package dicomnet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// DIMSE command set tags (group 0x0000).
const (
	TagCommandGroupLength             uint32 = 0x00000000
	TagAffectedSOPClassUID            uint32 = 0x00000002
	TagRequestedSOPClassUID           uint32 = 0x00000003
	TagCommandField                   uint32 = 0x00000100
	TagMessageID                      uint32 = 0x00000110
	TagMessageIDBeingRespondedTo      uint32 = 0x00000120
	TagMoveDestination                uint32 = 0x00000600
	TagPriority                       uint32 = 0x00000700
	TagCommandDataSetType             uint32 = 0x00000800
	TagStatus                         uint32 = 0x00000900
	TagOffendingElement               uint32 = 0x00000901
	TagErrorComment                   uint32 = 0x00000902
	TagErrorID                        uint32 = 0x00000903
	TagAffectedSOPInstanceUID         uint32 = 0x00001000
	TagRequestedSOPInstanceUID        uint32 = 0x00001001
	TagNumberOfRemainingSuboperations uint32 = 0x00001020
	TagNumberOfCompletedSuboperations uint32 = 0x00001021
	TagNumberOfFailedSuboperations    uint32 = 0x00001022
	TagNumberOfWarningSuboperations   uint32 = 0x00001023
)

// Command field values (PS3.7 §9.1). A response is the request value with bit
// 15 set.
const (
	CmdResponseBit uint16 = 0x8000

	CmdCStoreRQ  uint16 = 0x0001
	CmdCStoreRSP uint16 = 0x8001
	CmdCGetRQ    uint16 = 0x0010
	CmdCGetRSP   uint16 = 0x8010
	CmdCFindRQ   uint16 = 0x0020
	CmdCFindRSP  uint16 = 0x8020
	CmdCMoveRQ   uint16 = 0x0021
	CmdCMoveRSP  uint16 = 0x8021
	CmdCCancelRQ uint16 = 0x0FFF
	CmdCEchoRQ   uint16 = 0x0030
	CmdCEchoRSP  uint16 = 0x8030
)

// CommandDataSetType values (PS3.7 §9.3.1). Any value other than
// DataSetTypeNone means the message is followed by a data set.
const (
	DataSetTypePresent uint16 = 0x0001
	DataSetTypeNone    uint16 = 0x0101
)

// DIMSE status codes (PS3.7 Annex C and PS3.4 service definitions).
const (
	StatusSuccess          uint16 = 0x0000
	StatusPending          uint16 = 0xFF00
	StatusPendingWarning   uint16 = 0xFF01
	StatusCancel           uint16 = 0xFE00
	StatusCancelRefused    uint16 = 0xFE01
	StatusRefusedResources uint16 = 0xA700
	StatusDataSetMismatch  uint16 = 0xA900
	StatusSOPClassUnsup    uint16 = 0x0122
	StatusProcessingFail   uint16 = 0x0110
	StatusCannotUnderstand uint16 = 0xC000
)

// Message control header bits (PS3.8 §9.3.5).
const (
	MCHCommandBit byte = 0x01
	MCHLastBit    byte = 0x02
)

// Command element VRs. Command sets are always encoded Implicit VR Little
// Endian; the VR is only used locally for typed accessors.
var commandVRs = map[uint32]string{
	TagCommandGroupLength:             "UL",
	TagAffectedSOPClassUID:            "UI",
	TagRequestedSOPClassUID:           "UI",
	TagCommandField:                   "US",
	TagMessageID:                      "US",
	TagMessageIDBeingRespondedTo:      "US",
	TagMoveDestination:                "AE",
	TagPriority:                       "US",
	TagCommandDataSetType:             "US",
	TagStatus:                         "US",
	TagOffendingElement:               "AT",
	TagErrorComment:                   "LO",
	TagErrorID:                        "US",
	TagAffectedSOPInstanceUID:         "UI",
	TagRequestedSOPInstanceUID:        "UI",
	TagNumberOfRemainingSuboperations: "US",
	TagNumberOfCompletedSuboperations: "US",
	TagNumberOfFailedSuboperations:    "US",
	TagNumberOfWarningSuboperations:   "US",
}

// Errors returned by the DIMSE layer.
var (
	ErrNotCommandSet   = errors.New("dicomnet: not a DIMSE command set")
	ErrMissingAttr     = errors.New("dicomnet: required command attribute missing")
	ErrMessageTooLarge = errors.New("dicomnet: DIMSE message exceeds maximum length")
	ErrProtocolState   = errors.New("dicomnet: unexpected PDV during message assembly")
)

// CommandVR returns the VR of a command element tag.
func CommandVR(tag uint32) string {
	if vr, ok := commandVRs[tag]; ok {
		return vr
	}
	return "UN"
}

// IsCommandTag reports whether a tag belongs to the command group.
func IsCommandTag(tag uint32) bool { return tag>>16 == 0x0000 }

// CommandSet is a DIMSE command set, always encoded Implicit VR Little Endian.
type CommandSet struct {
	Fields map[uint32][]byte
}

// NewCommandSet returns an empty command set.
func NewCommandSet() *CommandSet {
	return &CommandSet{Fields: make(map[uint32][]byte)}
}

// SetUS stores a US (unsigned short) value.
func (c *CommandSet) SetUS(tag uint32, v uint16) {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, v)
	c.Fields[tag] = b
}

// SetUL stores a UL (unsigned long) value.
func (c *CommandSet) SetUL(tag uint32, v uint32) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	c.Fields[tag] = b
}

// SetUI stores a UI (UID) string value.
func (c *CommandSet) SetUI(tag uint32, v string) { c.Fields[tag] = []byte(v) }

// SetString stores a string value with the given VR.
func (c *CommandSet) SetString(tag uint32, v string) { c.Fields[tag] = []byte(v) }

// US returns a US value.
func (c *CommandSet) US(tag uint32) (uint16, bool) {
	b, ok := c.Fields[tag]
	if !ok || len(b) < 2 {
		return 0, false
	}
	return binary.LittleEndian.Uint16(b[:2]), true
}

// UL returns a UL value.
func (c *CommandSet) UL(tag uint32) (uint32, bool) {
	b, ok := c.Fields[tag]
	if !ok || len(b) < 4 {
		return 0, false
	}
	return binary.LittleEndian.Uint32(b[:4]), true
}

// UI returns a UI/string value with trailing padding removed.
func (c *CommandSet) UI(tag uint32) (string, bool) {
	b, ok := c.Fields[tag]
	if !ok {
		return "", false
	}
	return strings.TrimRight(string(b), " \x00"), true
}

// CommandField returns the (0000,0100) command field value.
func (c *CommandSet) CommandField() (uint16, bool) { return c.US(TagCommandField) }

// Status returns the (0000,0900) status value.
func (c *CommandSet) Status() (uint16, bool) { return c.US(TagStatus) }

// MessageID returns the (0000,0110) message ID.
func (c *CommandSet) MessageID() (uint16, bool) { return c.US(TagMessageID) }

// MessageIDBeingRespondedTo returns the (0000,0120) value.
func (c *CommandSet) MessageIDBeingRespondedTo() (uint16, bool) {
	return c.US(TagMessageIDBeingRespondedTo)
}

// IsResponse reports whether the command field has the response bit set.
func (c *CommandSet) IsResponse() bool {
	cf, ok := c.CommandField()
	return ok && cf&CmdResponseBit != 0
}

// HasDataSet reports whether a data set follows the command set. A missing
// (0000,0800) element is treated as "data set present", which is the
// conservative interpretation: the PDVs that follow will be consumed.
func (c *CommandSet) HasDataSet() bool {
	v, ok := c.US(TagCommandDataSetType)
	if !ok {
		return true
	}
	return v != DataSetTypeNone
}

// Encode serialises the command set as Implicit VR Little Endian, computing
// (0000,0000) Command Group Length correctly. Fields are emitted in ascending
// tag order for deterministic output.
func (c *CommandSet) Encode() ([]byte, error) {
	tags := make([]uint32, 0, len(c.Fields))
	for tag := range c.Fields {
		if tag == TagCommandGroupLength {
			continue
		}
		tags = append(tags, tag)
	}
	sort.Slice(tags, func(i, j int) bool { return tags[i] < tags[j] })

	var body []byte
	for _, tag := range tags {
		v := c.Fields[tag]
		if uint32(len(v)) > MaxElementLength {
			return nil, ErrMessageTooLarge
		}
		var hdr [8]byte
		putTag(hdr[0:4], tag)
		binary.LittleEndian.PutUint32(hdr[4:8], uint32(len(v)))
		body = append(body, hdr[:]...)
		body = append(body, v...)
	}

	out := make([]byte, 0, 12+len(body))
	var gl [12]byte
	putTag(gl[0:4], TagCommandGroupLength)
	binary.LittleEndian.PutUint32(gl[4:8], 4) // UL
	binary.LittleEndian.PutUint32(gl[8:12], uint32(len(body)))
	out = append(out, gl[:]...)
	return append(out, body...), nil
}

// ParseCommandSet decodes a DIMSE command set from Implicit VR Little Endian
// bytes. Fatal protocol violations are returned as errors; the (0000,0000)
// group length is validated when present.
func ParseCommandSet(data []byte) (*CommandSet, error) {
	if len(data) == 0 {
		return nil, ErrNotCommandSet
	}
	ds, err := DecodeDataSet(data, false)
	if err != nil {
		return nil, err
	}
	c := &CommandSet{Fields: make(map[uint32][]byte)}
	for _, e := range ds {
		if !IsCommandTag(e.Tag) {
			return nil, fmt.Errorf("%w: tag 0x%08X in command set", ErrNotCommandSet, e.Tag)
		}
		if e.Tag == TagCommandGroupLength {
			if gl, ok := e.Uint32(); ok {
				// Group length counts the bytes after this element.
				if uint64(gl) > uint64(len(data)) {
					return nil, fmt.Errorf("%w: group length %d > %d", ErrNotCommandSet, gl, len(data))
				}
			}
			continue
		}
		cp := make([]byte, len(e.Value))
		copy(cp, e.Value)
		c.Fields[e.Tag] = cp
	}
	if _, ok := c.Fields[TagCommandField]; !ok {
		return nil, fmt.Errorf("%w: (0000,0100)", ErrMissingAttr)
	}
	return c, nil
}

// NewCFindResponse builds a C-FIND-RSP command set.
func NewCFindResponse(messageIDBeingRespondedTo uint16, sopClassUID string, status uint16, hasIdentifier bool) *CommandSet {
	c := NewCommandSet()
	c.SetUS(TagCommandField, CmdCFindRSP)
	c.SetUS(TagMessageIDBeingRespondedTo, messageIDBeingRespondedTo)
	if sopClassUID != "" {
		c.SetUI(TagAffectedSOPClassUID, sopClassUID)
	}
	c.SetUS(TagStatus, status)
	if hasIdentifier {
		c.SetUS(TagCommandDataSetType, DataSetTypePresent)
	} else {
		c.SetUS(TagCommandDataSetType, DataSetTypeNone)
	}
	return c
}

// NewCFindRequest builds a C-FIND-RQ command set (used by clients and tests).
func NewCFindRequest(messageID uint16, sopClassUID string, priority uint16) *CommandSet {
	c := NewCommandSet()
	c.SetUS(TagCommandField, CmdCFindRQ)
	c.SetUS(TagMessageID, messageID)
	if sopClassUID != "" {
		c.SetUI(TagRequestedSOPClassUID, sopClassUID)
	}
	c.SetUS(TagPriority, priority)
	c.SetUS(TagCommandDataSetType, DataSetTypePresent)
	return c
}

// NewCEchoResponse builds a C-ECHO-RSP command set.
func NewCEchoResponse(messageIDBeingRespondedTo uint16, sopClassUID string, status uint16) *CommandSet {
	c := NewCommandSet()
	c.SetUS(TagCommandField, CmdCEchoRSP)
	c.SetUS(TagMessageIDBeingRespondedTo, messageIDBeingRespondedTo)
	if sopClassUID != "" {
		c.SetUI(TagAffectedSOPClassUID, sopClassUID)
	}
	c.SetUS(TagStatus, status)
	c.SetUS(TagCommandDataSetType, DataSetTypeNone)
	return c
}

// MessageControlHeader builds the PDV message control header byte from the
// command / last-fragment flags.
func MessageControlHeader(isCommand, isLast bool) byte {
	var mch byte
	if isCommand {
		mch |= MCHCommandBit
	}
	if isLast {
		mch |= MCHLastBit
	}
	return mch
}

// ControlHeader returns the PDV's message control header byte.
func (p PDV) ControlHeader() byte { return MessageControlHeader(p.IsCommand, p.IsLast) }

// Fragment splits a DIMSE command set and optional data set into PDVs that fit
// the peer's negotiated maximum PDU length. peerMaxPDU is the value the peer
// sent in its association request (0 means "no maximum", in which case our own
// default is used). The function never returns a PDV whose enclosing P-DATA-TF
// PDU would exceed that length.
func Fragment(contextID byte, command, dataSet []byte, peerMaxPDU uint32) ([]PDV, error) {
	limit := peerMaxPDU
	if limit == 0 || limit > MaxPDUBodySize {
		limit = DefaultMaxPDULength
	}
	// PDU header (6) + PDV header (4) + context ID + message control header (2).
	if limit <= pduHeaderSize+6 {
		return nil, fmt.Errorf("%w: negotiated maximum PDU length %d too small", ErrProtocolState, limit)
	}
	chunk := int(limit) - pduHeaderSize - 6
	if chunk > int(MaxPDUBodySize) {
		chunk = int(MaxPDUBodySize)
	}

	var pdvs []PDV
	pdvs = append(pdvs, fragmentValue(contextID, command, true, chunk)...)
	if len(dataSet) > 0 {
		pdvs = append(pdvs, fragmentValue(contextID, dataSet, false, chunk)...)
	}
	if len(pdvs) == 0 {
		return nil, ErrProtocolState
	}
	return pdvs, nil
}

// fragmentValue splits one value into PDVs no larger than chunk bytes, marking
// the final PDV as the last fragment.
func fragmentValue(contextID byte, value []byte, isCommand bool, chunk int) []PDV {
	if len(value) == 0 {
		return []PDV{{ContextID: contextID, IsCommand: isCommand, IsLast: true}}
	}
	var out []PDV
	for len(value) > 0 {
		n := chunk
		if n > len(value) {
			n = len(value)
		}
		part := make([]byte, n)
		copy(part, value[:n])
		out = append(out, PDV{
			ContextID: contextID,
			IsCommand: isCommand,
			IsLast:    n == len(value),
			Data:      part,
		})
		value = value[n:]
	}
	return out
}

// MessageAccumulator reassembles a DIMSE message from a sequence of PDVs: one
// command set optionally followed by one data set.
type MessageAccumulator struct {
	cmd      []byte
	data     []byte
	cmdDone  bool
	dataLast bool
	wantData bool
	total    int
}

// NewMessageAccumulator returns an empty accumulator.
func NewMessageAccumulator() *MessageAccumulator { return &MessageAccumulator{} }

// Feed adds a PDV. The command-vs-data distinction comes from bit 0 of the
// message control header, and the end of each value from bit 1.
func (a *MessageAccumulator) Feed(pdv PDV) error {
	a.total += len(pdv.Data)
	if a.total > MaxElementLength {
		return ErrMessageTooLarge
	}
	if !a.cmdDone {
		if !pdv.IsCommand {
			return fmt.Errorf("%w: data PDV before command set", ErrProtocolState)
		}
		a.cmd = append(a.cmd, pdv.Data...)
		if pdv.IsLast {
			a.cmdDone = true
			// Whether a data set follows is decided from the command set.
			cs, err := ParseCommandSet(a.cmd)
			if err != nil {
				return err
			}
			a.wantData = cs.HasDataSet()
			if !a.wantData {
				a.dataLast = true
			}
		}
		return nil
	}
	if pdv.IsCommand {
		return fmt.Errorf("%w: command PDV after command set", ErrProtocolState)
	}
	if !a.wantData {
		return fmt.Errorf("%w: data PDV after data set complete", ErrProtocolState)
	}
	a.data = append(a.data, pdv.Data...)
	if pdv.IsLast {
		a.dataLast = true
	}
	return nil
}

// CommandComplete reports whether the command set has been fully received.
func (a *MessageAccumulator) CommandComplete() bool { return a.cmdDone }

// Complete reports whether the whole message (command plus any data set) has
// been received.
func (a *MessageAccumulator) Complete() bool { return a.cmdDone && a.dataLast }

// WantDataSet reports whether the command set indicated a following data set.
func (a *MessageAccumulator) WantDataSet() bool { return a.wantData }

// Command returns the accumulated command set bytes.
func (a *MessageAccumulator) Command() *CommandSet {
	cs, err := ParseCommandSet(a.cmd)
	if err != nil {
		return nil
	}
	return cs
}

// DataSetBytes returns the accumulated identifier/data set bytes.
func (a *MessageAccumulator) DataSetBytes() []byte { return a.data }
