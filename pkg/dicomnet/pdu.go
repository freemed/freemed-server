// Package dicomnet implements the DICOM network (DIMSE) protocol — Upper Layer
// PDUs and DIMSE message services — using only the Go standard library.
//
// The package is deliberately split so that protocol handling has no database
// dependency: pdu.go, dimse.go, element.go, mwl.go and scp.go implement the
// wire protocol and query semantics, while the worklist data itself is
// supplied through the WorklistSource interface. A database-backed
// implementation lives outside this package.
//
// Implemented: A-ASSOCIATE-RQ/AC/RJ, P-DATA-TF, A-RELEASE-RQ/RP, A-ABORT,
// Implicit VR Little Endian and Explicit VR Little Endian transfer syntaxes,
// and the C-FIND SCP side of the Modality Worklist Information Model.
//
// All peer-supplied input is treated as hostile: every length is bounds
// checked before it is used to slice or allocate, and no decoder panics.
package dicomnet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Upper Layer PDU types (PS3.8 §9.3).
const (
	PDUAssociateRQ byte = 0x01
	PDUAssociateAC byte = 0x02
	PDUAssociateRJ byte = 0x03
	PDUDataTF      byte = 0x04
	PDUReleaseRQ   byte = 0x05
	PDUReleaseRP   byte = 0x06
	PDUAbort       byte = 0x07
)

// Association item types (PS3.8 §9.3).
const (
	itemApplicationContext     byte = 0x10
	itemPresentationContextRQ  byte = 0x20
	itemPresentationContextAC  byte = 0x21
	itemAbstractSyntax         byte = 0x30
	itemTransferSyntax         byte = 0x40
	itemUserInformation        byte = 0x50
	itemMaxLength              byte = 0x51
	itemImplementationClassUID byte = 0x52
	itemAsynchronousOperations byte = 0x53
	itemSCPRoleSelection       byte = 0x54
	itemImplementationVersion  byte = 0x55
	itemSOPClassExtendedNegot  byte = 0x56
)

// Presentation context negotiation results (PS3.8 Table 9-18).
const (
	PresContextAcceptance          byte = 0x00
	PresContextUserRejection       byte = 0x01
	PresContextNoReason            byte = 0x02
	PresContextAbstractSyntaxUnsup byte = 0x03
	PresContextTransferSyntaxUnsup byte = 0x04
)

// A-ASSOCIATE-RJ result/source/reason values.
const (
	RJResultPermanent byte = 0x01
	RJResultTransient byte = 0x02

	RJSourceServiceUser     byte = 0x01
	RJSourceServiceProvider byte = 0x02

	RJReasonNoReason              byte = 0x01
	RJReasonAppContextNotSup      byte = 0x02
	RJReasonCallingAENotRecog     byte = 0x03
	RJReasonReserved              byte = 0x04
	RJReasonProtocolVersionNotSup byte = 0x06
	RJReasonCalledAENotRecog      byte = 0x07

	AbortSourceServiceUser     byte = 0x00
	AbortSourceServiceProvider byte = 0x02
)

// Well known UIDs used during association negotiation.
const (
	UIDApplicationContext = "1.2.840.10008.3.1.1.1"

	// UIDVerification is the Verification SOP Class (C-ECHO).
	UIDVerification = "1.2.840.10008.1.1"

	// UIDModalityWorklistFIND is the Modality Worklist Information Model –
	// FIND SOP Class used for MWL C-FIND.
	UIDModalityWorklistFIND = "1.2.840.10008.5.1.4.31"
)

// Size limits. These are hard ceilings: a peer may not make the server
// allocate more than this for a single PDU.
const (
	// DefaultMaxPDULength is the maximum PDU length we advertise when the
	// peer does not request a specific value.
	DefaultMaxPDULength uint32 = 16384

	// MaxPDUBodySize is the absolute ceiling on a received PDU body. Any PDU
	// claiming to be larger is rejected without allocating.
	MaxPDUBodySize uint32 = 16 << 20 // 16 MiB

	// pduHeaderSize is the fixed 6-byte UL PDU header.
	pduHeaderSize = 6

	// aeTitleLength is the fixed width of an AE title in an association PDU.
	aeTitleLength = 16

	// MaxAETitleLength is the maximum length of an AE title.
	MaxAETitleLength = 16
)

// Errors returned by the PDU layer. Callers should treat all of them as
// protocol violations and terminate the association.
var (
	ErrPDUTooLarge    = errors.New("dicomnet: PDU length exceeds maximum")
	ErrShortPDU       = errors.New("dicomnet: truncated PDU")
	ErrInvalidPDUType = errors.New("dicomnet: unknown PDU type")
	ErrMalformedItem  = errors.New("dicomnet: malformed association item")
	ErrInvalidAETitle = errors.New("dicomnet: invalid AE title")
)

// PresentationContext carries one proposed (RQ) or negotiated (AC)
// presentation context.
type PresentationContext struct {
	ID               byte
	AbstractSyntax   string
	TransferSyntaxes []string // populated in RQ; the chosen syntax in AC
	Result           byte     // acceptance result (AC only)
}

// PDV is a single Presentation Data Value inside a P-DATA-TF PDU.
//
// IsCommand is derived from bit 0 of the message control header and IsLast
// from bit 1 (PS3.8 §9.3.5).
type PDV struct {
	ContextID byte
	IsCommand bool
	IsLast    bool
	Data      []byte
}

// PDU is a decoded (or to-be-encoded) Upper Layer PDU.
type PDU struct {
	Type byte

	// Association RQ/AC fields.
	ProtocolVersion    uint16
	CalledAETitle      string
	CallingAETitle     string
	ApplicationContext string

	PresentationContexts []PresentationContext

	// User information.
	MaxPDULength              uint32
	ImplementationClassUID    string
	ImplementationVersionName string

	// A-ASSOCIATE-RJ fields.
	Result byte
	Source byte
	Reason byte

	// A-ABORT fields.
	AbortSource byte
	AbortReason byte

	// P-DATA-TF payload.
	PDVs []PDV
}

// NewAssociateRQ builds an A-ASSOCIATE-RQ PDU carrying the given presentation
// contexts and user information.
func NewAssociateRQ(calledAE, callingAE, appContext string, pcs []PresentationContext, maxPDU uint32, implClassUID, implVersion string) *PDU {
	return &PDU{
		Type:                      PDUAssociateRQ,
		ProtocolVersion:           1,
		CalledAETitle:             calledAE,
		CallingAETitle:            callingAE,
		ApplicationContext:        appContext,
		PresentationContexts:      pcs,
		MaxPDULength:              maxPDU,
		ImplementationClassUID:    implClassUID,
		ImplementationVersionName: implVersion,
	}
}

// NewAssociateRJ builds an A-ASSOCIATE-RJ PDU.
func NewAssociateRJ(result, source, reason byte) *PDU {
	return &PDU{Type: PDUAssociateRJ, Result: result, Source: source, Reason: reason}
}

// NewReleaseRQ builds an A-RELEASE-RQ PDU.
func NewReleaseRQ() *PDU { return &PDU{Type: PDUReleaseRQ} }

// NewReleaseRP builds an A-RELEASE-RP PDU.
func NewReleaseRP() *PDU { return &PDU{Type: PDUReleaseRP} }

// NewAbort builds an A-ABORT PDU.
func NewAbort(source, reason byte) *PDU {
	return &PDU{Type: PDUAbort, AbortSource: source, AbortReason: reason}
}

// NewDataTF builds a P-DATA-TF PDU from the given PDVs. Each PDV payload must
// already have been sized to fit inside the negotiated maximum PDU length —
// see Fragment.
func NewDataTF(pdvs ...PDV) *PDU { return &PDU{Type: PDUDataTF, PDVs: pdvs} }

// Encode serialises the PDU, including the 6-byte UL header.
func (p *PDU) Encode() ([]byte, error) {
	var body []byte
	switch p.Type {
	case PDUAssociateRQ:
		body = p.encodeAssociation(true)
	case PDUAssociateAC:
		body = p.encodeAssociation(false)
	case PDUAssociateRJ:
		if len(p.CalledAETitle) > 0 {
			_ = p.CalledAETitle // AE titles are not part of an RJ
		}
		body = make([]byte, 4)
		body[1] = p.Result
		body[2] = p.Source
		body[3] = p.Reason
	case PDUDataTF:
		for _, pdv := range p.PDVs {
			if len(pdv.Data) > int(MaxPDUBodySize) {
				return nil, ErrPDUTooLarge
			}
			hdr := make([]byte, 6)
			binary.BigEndian.PutUint32(hdr[0:4], uint32(len(pdv.Data)+2))
			hdr[4] = pdv.ContextID
			var mch byte
			if pdv.IsCommand {
				mch |= 0x01
			}
			if pdv.IsLast {
				mch |= 0x02
			}
			hdr[5] = mch
			body = append(body, hdr...)
			body = append(body, pdv.Data...)
		}
	case PDUReleaseRQ, PDUReleaseRP:
		body = make([]byte, 4)
	case PDUAbort:
		body = make([]byte, 4)
		body[2] = p.AbortSource
		body[3] = p.AbortReason
	default:
		return nil, fmt.Errorf("%w: 0x%02x", ErrInvalidPDUType, p.Type)
	}

	if uint32(len(body)) > MaxPDUBodySize {
		return nil, ErrPDUTooLarge
	}

	out := make([]byte, pduHeaderSize+len(body))
	out[0] = p.Type
	binary.BigEndian.PutUint32(out[2:6], uint32(len(body)))
	copy(out[pduHeaderSize:], body)
	return out, nil
}

// encodeAssociation builds the body of an A-ASSOCIATE-RQ (rq == true) or
// A-ASSOCIATE-AC PDU.
func (p *PDU) encodeAssociation(rq bool) []byte {
	body := make([]byte, 0, 128)
	var hdr [4]byte
	ver := p.ProtocolVersion
	if ver == 0 {
		ver = 1
	}
	binary.BigEndian.PutUint16(hdr[0:2], ver)
	body = append(body, hdr[0:4]...)

	// Called / Calling AE titles (fixed-width, space padded).
	body = append(body, padAETitle(p.CalledAETitle)...)
	body = append(body, padAETitle(p.CallingAETitle)...)
	body = append(body, make([]byte, 32)...) // reserved

	appCtx := p.ApplicationContext
	if appCtx == "" {
		appCtx = UIDApplicationContext
	}
	body = append(body, encodeItem(itemApplicationContext, []byte(appCtx))...)

	for _, pc := range p.PresentationContexts {
		body = append(body, encodePresentationContext(pc, rq)...)
	}

	var user []byte
	maxLen := p.MaxPDULength
	if maxLen == 0 {
		maxLen = DefaultMaxPDULength
	}
	ml := make([]byte, 4)
	binary.BigEndian.PutUint32(ml, maxLen)
	user = append(user, encodeItem(itemMaxLength, ml)...)

	implUID := p.ImplementationClassUID
	if implUID == "" {
		implUID = ImplementationClassUID
	}
	user = append(user, encodeItem(itemImplementationClassUID, []byte(implUID))...)

	ver2 := p.ImplementationVersionName
	if ver2 == "" {
		ver2 = ImplementationVersionName
	}
	user = append(user, encodeItem(itemImplementationVersion, []byte(ver2))...)

	body = append(body, encodeItem(itemUserInformation, user)...)
	return body
}

// encodePresentationContext encodes one presentation context item.
func encodePresentationContext(pc PresentationContext, rq bool) []byte {
	var body []byte
	body = append(body, pc.ID, 0, 0, 0)
	if rq {
		body = append(body, encodeItem(itemAbstractSyntax, []byte(pc.AbstractSyntax))...)
		for _, ts := range pc.TransferSyntaxes {
			body = append(body, encodeItem(itemTransferSyntax, []byte(ts))...)
		}
		return encodeItem(itemPresentationContextRQ, body)
	}
	body[2] = pc.Result
	ts := ""
	if len(pc.TransferSyntaxes) > 0 {
		ts = pc.TransferSyntaxes[0]
	}
	body = append(body, encodeItem(itemTransferSyntax, []byte(ts))...)
	return encodeItem(itemPresentationContextAC, body)
}

// encodeItem encodes a DICOM UL item with an explicit (long) length field.
func encodeItem(typ byte, body []byte) []byte {
	if len(body) > 0xFFFF {
		body = body[:0xFFFF]
	}
	out := make([]byte, 4+len(body))
	out[0] = typ
	binary.BigEndian.PutUint16(out[2:4], uint16(len(body)))
	copy(out[4:], body)
	return out
}

// padAETitle space pads (or truncates) an AE title to the fixed 16 byte field
// width used in association PDUs.
func padAETitle(s string) []byte {
	out := make([]byte, aeTitleLength)
	for i := range out {
		out[i] = ' '
	}
	s = strings.TrimSpace(s)
	if len(s) > aeTitleLength {
		s = s[:aeTitleLength]
	}
	copy(out, s)
	return out
}

// ReadPDU reads a single UL PDU from r. The declared body length is validated
// against maxBody before any buffer is allocated.
func ReadPDU(r io.Reader, maxBody uint32) (*PDU, error) {
	var hdr [pduHeaderSize]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return nil, err
	}
	typ := hdr[0]
	length := binary.BigEndian.Uint32(hdr[2:6])
	if maxBody == 0 || maxBody > MaxPDUBodySize {
		maxBody = MaxPDUBodySize
	}
	if length > maxBody {
		return nil, fmt.Errorf("%w: type 0x%02x length %d > %d", ErrPDUTooLarge, typ, length, maxBody)
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrShortPDU, err)
	}
	return DecodePDU(typ, body)
}

// DecodePDU decodes a PDU body (without the 6-byte header).
func DecodePDU(typ byte, body []byte) (*PDU, error) {
	if uint32(len(body)) > MaxPDUBodySize {
		return nil, ErrPDUTooLarge
	}
	switch typ {
	case PDUAssociateRQ:
		return decodeAssociation(typ, body, true)
	case PDUAssociateAC:
		return decodeAssociation(typ, body, false)
	case PDUAssociateRJ:
		if len(body) < 4 {
			return nil, fmt.Errorf("%w: association reject body %d bytes", ErrShortPDU, len(body))
		}
		return &PDU{Type: typ, Result: body[1], Source: body[2], Reason: body[3]}, nil
	case PDUDataTF:
		return decodeDataTF(body)
	case PDUReleaseRQ, PDUReleaseRP:
		if len(body) < 4 {
			return nil, fmt.Errorf("%w: release body %d bytes", ErrShortPDU, len(body))
		}
		return &PDU{Type: typ}, nil
	case PDUAbort:
		if len(body) < 4 {
			return nil, fmt.Errorf("%w: abort body %d bytes", ErrShortPDU, len(body))
		}
		return &PDU{Type: typ, AbortSource: body[2], AbortReason: body[3]}, nil
	default:
		return nil, fmt.Errorf("%w: 0x%02x", ErrInvalidPDUType, typ)
	}
}

// decodeDataTF splits a P-DATA-TF body into PDVs.
func decodeDataTF(body []byte) (*PDU, error) {
	p := &PDU{Type: PDUDataTF}
	pos := 0
	for pos < len(body) {
		if pos+6 > len(body) {
			return nil, fmt.Errorf("%w: PDV header truncated at %d", ErrShortPDU, pos)
		}
		plen := binary.BigEndian.Uint32(body[pos : pos+4])
		if plen < 2 {
			return nil, fmt.Errorf("%w: PDV length %d is too small", ErrMalformedItem, plen)
		}
		// plen includes the 1-byte context ID and 1-byte message control header.
		if uint64(pos)+4+uint64(plen) > uint64(len(body)) {
			return nil, fmt.Errorf("%w: PDV claims %d bytes, only %d remain", ErrShortPDU, plen, len(body)-pos-4)
		}
		mch := body[pos+5]
		data := body[pos+6 : pos+4+int(plen)]
		cp := make([]byte, len(data))
		copy(cp, data)
		p.PDVs = append(p.PDVs, PDV{
			ContextID: body[pos+4],
			IsCommand: mch&0x01 != 0,
			IsLast:    mch&0x02 != 0,
			Data:      cp,
		})
		pos += 4 + int(plen)
	}
	return p, nil
}

// decodeAssociation decodes an A-ASSOCIATE-RQ or A-ASSOCIATE-AC body.
func decodeAssociation(typ byte, body []byte, rq bool) (*PDU, error) {
	// protocol version (2) + reserved (2) + called AE (16) + calling AE (16) + reserved (32)
	const fixed = 2 + 2 + aeTitleLength + aeTitleLength + 32
	if len(body) < fixed {
		return nil, fmt.Errorf("%w: association body %d bytes", ErrShortPDU, len(body))
	}
	p := &PDU{Type: typ}
	p.ProtocolVersion = binary.BigEndian.Uint16(body[0:2])
	p.CalledAETitle = readAETitle(body[4 : 4+aeTitleLength])
	p.CallingAETitle = readAETitle(body[4+aeTitleLength : 4+2*aeTitleLength])

	pos := fixed
	for pos < len(body) {
		ityp, ibody, next, err := decodeItem(body, pos)
		if err != nil {
			return nil, err
		}
		pos = next
		switch ityp {
		case itemApplicationContext:
			p.ApplicationContext = strings.TrimRight(string(ibody), "\x00 ")
		case itemPresentationContextRQ, itemPresentationContextAC:
			pc, err := decodePresentationContext(ibody, rq)
			if err != nil {
				return nil, err
			}
			p.PresentationContexts = append(p.PresentationContexts, pc)
		case itemUserInformation:
			if err := decodeUserInformation(p, ibody); err != nil {
				return nil, err
			}
		default:
			// Unknown top-level item: ignored for forward compatibility.
		}
	}
	return p, nil
}

// decodePresentationContext decodes a presentation context RQ/AC item body.
func decodePresentationContext(body []byte, rq bool) (PresentationContext, error) {
	var pc PresentationContext
	if len(body) < 4 {
		return pc, fmt.Errorf("%w: presentation context body %d bytes", ErrShortPDU, len(body))
	}
	pc.ID = body[0]
	pc.Result = body[2]
	pos := 4
	for pos < len(body) {
		ityp, ibody, next, err := decodeItem(body, pos)
		if err != nil {
			return pc, err
		}
		pos = next
		switch ityp {
		case itemAbstractSyntax:
			pc.AbstractSyntax = strings.TrimRight(string(ibody), "\x00 ")
		case itemTransferSyntax:
			pc.TransferSyntaxes = append(pc.TransferSyntaxes, strings.TrimRight(string(ibody), "\x00 "))
		default:
			// ignore unknown sub-items
		}
	}
	if !rq && len(pc.TransferSyntaxes) == 0 {
		pc.TransferSyntaxes = append(pc.TransferSyntaxes, "")
	}
	return pc, nil
}

// decodeUserInformation decodes the user information item body.
func decodeUserInformation(p *PDU, body []byte) error {
	pos := 0
	for pos < len(body) {
		ityp, ibody, next, err := decodeItem(body, pos)
		if err != nil {
			return err
		}
		pos = next
		switch ityp {
		case itemMaxLength:
			if len(ibody) < 4 {
				return fmt.Errorf("%w: maximum length sub-item is %d bytes", ErrMalformedItem, len(ibody))
			}
			p.MaxPDULength = binary.BigEndian.Uint32(ibody[0:4])
		case itemImplementationClassUID:
			p.ImplementationClassUID = strings.TrimRight(string(ibody), "\x00 ")
		case itemImplementationVersion:
			p.ImplementationVersionName = strings.TrimRight(string(ibody), "\x00 ")
		default:
			// Asynchronous operations, role selection, extended negotiation:
			// accepted but not acted upon.
		}
	}
	return nil
}

// decodeItem reads one UL item at pos and returns its type, body and the
// offset of the following item. Every length is validated against the buffer.
func decodeItem(buf []byte, pos int) (byte, []byte, int, error) {
	if pos < 0 || pos+4 > len(buf) {
		return 0, nil, 0, fmt.Errorf("%w: item header truncated at %d", ErrShortPDU, pos)
	}
	typ := buf[pos]
	length := int(binary.BigEndian.Uint16(buf[pos+2 : pos+4]))
	start := pos + 4
	if length < 0 || start+length > len(buf) {
		return 0, nil, 0, fmt.Errorf("%w: item 0x%02x claims %d bytes, only %d remain", ErrMalformedItem, typ, length, len(buf)-start)
	}
	return typ, buf[start : start+length], start + length, nil
}

// readAETitle trims trailing padding from a fixed-width AE title field.
func readAETitle(b []byte) string {
	return strings.TrimRight(string(b), " \x00")
}
