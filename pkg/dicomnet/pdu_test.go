package dicomnet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Reference encoders. These build PDU bytes by hand from PS3.8 §9.3, without
// touching the production encoder, so that encoder/decoder symmetric bugs are
// caught rather than cancelled out.
// ---------------------------------------------------------------------------

func refItem(typ byte, body []byte) []byte {
	b := make([]byte, 4, 4+len(body))
	b[0] = typ
	b[2] = byte(len(body) >> 8)
	b[3] = byte(len(body))
	return append(b, body...)
}

func refAETitle(s string) []byte {
	b := bytes.Repeat([]byte{' '}, 16)
	copy(b, s)
	return b
}

func refPDUBytes(typ byte, body []byte) []byte {
	out := make([]byte, 6, 6+len(body))
	out[0] = typ
	binary.BigEndian.PutUint32(out[2:6], uint32(len(body)))
	return append(out, body...)
}

// refAssociateRQ builds the byte image of an A-ASSOCIATE-RQ for a single
// presentation context.
func refAssociateRQ(called, calling, abstract string, transferSyntaxes []string, maxPDU uint32) []byte {
	var body []byte
	body = append(body, 0x00, 0x01) // protocol version 1
	body = append(body, 0x00, 0x00) // reserved
	body = append(body, refAETitle(called)...)
	body = append(body, refAETitle(calling)...)
	body = append(body, make([]byte, 32)...)

	body = append(body, refItem(itemApplicationContext, []byte(UIDApplicationContext))...)

	var pc []byte
	pc = append(pc, 0x01, 0x00, 0x00, 0x00) // ID 1 + reserved
	pc = append(pc, refItem(itemAbstractSyntax, []byte(abstract))...)
	for _, ts := range transferSyntaxes {
		pc = append(pc, refItem(itemTransferSyntax, []byte(ts))...)
	}
	body = append(body, refItem(itemPresentationContextRQ, pc)...)

	var ui []byte
	ml := make([]byte, 4)
	binary.BigEndian.PutUint32(ml, maxPDU)
	ui = append(ui, refItem(itemMaxLength, ml)...)
	ui = append(ui, refItem(itemImplementationClassUID, []byte(ImplementationClassUID))...)
	ui = append(ui, refItem(itemImplementationVersion, []byte(ImplementationVersionName))...)
	body = append(body, refItem(itemUserInformation, ui)...)

	return refPDUBytes(PDUAssociateRQ, body)
}

// ---------------------------------------------------------------------------
// A-ASSOCIATE-RQ
// ---------------------------------------------------------------------------

func TestAssociateRQEncoderMatchesReferenceBytes(t *testing.T) {
	rq := NewAssociateRQ("FREEMED", "SCU_AE", UIDApplicationContext, []PresentationContext{{
		ID:               1,
		AbstractSyntax:   UIDModalityWorklistFIND,
		TransferSyntaxes: []string{ImplicitVRLittleEndian, ExplicitVRLittleEndian},
	}}, 16384, ImplementationClassUID, ImplementationVersionName)

	got, err := rq.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	want := refAssociateRQ("FREEMED", "SCU_AE", UIDModalityWorklistFIND,
		[]string{ImplicitVRLittleEndian, ExplicitVRLittleEndian}, 16384)

	if !bytes.Equal(got, want) {
		t.Fatalf("A-ASSOCIATE-RQ bytes differ from hand-built reference\n got=%x\nwant=%x", got, want)
	}
}

func TestDecodeAssociateRQReferenceBytes(t *testing.T) {
	raw := refAssociateRQ("FREEMED", "SCU_AE", UIDModalityWorklistFIND, []string{ImplicitVRLittleEndian}, 32768)

	pdu, err := ReadPDU(bytes.NewReader(raw), MaxPDUBodySize)
	if err != nil {
		t.Fatalf("ReadPDU: %v", err)
	}
	if pdu.Type != PDUAssociateRQ {
		t.Fatalf("type = 0x%02X, want 0x01", pdu.Type)
	}
	if pdu.ProtocolVersion != 1 {
		t.Errorf("protocol version = %d, want 1", pdu.ProtocolVersion)
	}
	if pdu.CalledAETitle != "FREEMED" {
		t.Errorf("called AE = %q, want FREEMED", pdu.CalledAETitle)
	}
	if pdu.CallingAETitle != "SCU_AE" {
		t.Errorf("calling AE = %q, want SCU_AE", pdu.CallingAETitle)
	}
	if pdu.ApplicationContext != UIDApplicationContext {
		t.Errorf("application context = %q", pdu.ApplicationContext)
	}
	if len(pdu.PresentationContexts) != 1 {
		t.Fatalf("presentation contexts = %d, want 1", len(pdu.PresentationContexts))
	}
	pc := pdu.PresentationContexts[0]
	if pc.ID != 1 || pc.AbstractSyntax != UIDModalityWorklistFIND {
		t.Errorf("presentation context = %+v", pc)
	}
	if len(pc.TransferSyntaxes) != 1 || pc.TransferSyntaxes[0] != ImplicitVRLittleEndian {
		t.Errorf("transfer syntaxes = %v", pc.TransferSyntaxes)
	}
	if pdu.MaxPDULength != 32768 {
		t.Errorf("max PDU length = %d, want 32768", pdu.MaxPDULength)
	}
	if pdu.ImplementationClassUID != ImplementationClassUID {
		t.Errorf("implementation class UID = %q", pdu.ImplementationClassUID)
	}
	if pdu.ImplementationVersionName != ImplementationVersionName {
		t.Errorf("implementation version = %q", pdu.ImplementationVersionName)
	}
}

func TestAssociateRQEncodeDecodeRoundTrip(t *testing.T) {
	pcs := []PresentationContext{
		{ID: 1, AbstractSyntax: UIDModalityWorklistFIND, TransferSyntaxes: []string{ImplicitVRLittleEndian}},
		{ID: 3, AbstractSyntax: UIDVerification, TransferSyntaxes: []string{ExplicitVRLittleEndian, ImplicitVRLittleEndian}},
		{ID: 5, AbstractSyntax: "1.2.840.10008.5.1.4.1.1.2", TransferSyntaxes: []string{ExplicitVRLittleEndian}},
	}
	rq := NewAssociateRQ("MWL_SCP", "MODALITY1", "", pcs, 65536, "", "")

	raw, err := rq.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, err := ReadPDU(bytes.NewReader(raw), MaxPDUBodySize)
	if err != nil {
		t.Fatalf("ReadPDU: %v", err)
	}
	if got.ApplicationContext != UIDApplicationContext {
		t.Errorf("default application context = %q", got.ApplicationContext)
	}
	if got.MaxPDULength != 65536 {
		t.Errorf("max PDU length = %d", got.MaxPDULength)
	}
	if got.ImplementationClassUID != ImplementationClassUID {
		t.Errorf("default implementation class UID = %q", got.ImplementationClassUID)
	}
	if len(got.PresentationContexts) != 3 {
		t.Fatalf("presentation contexts = %d, want 3", len(got.PresentationContexts))
	}
	for i, pc := range got.PresentationContexts {
		if pc.ID != pcs[i].ID || pc.AbstractSyntax != pcs[i].AbstractSyntax {
			t.Errorf("pc[%d] = %+v, want ID %d abstract %s", i, pc, pcs[i].ID, pcs[i].AbstractSyntax)
		}
		if len(pc.TransferSyntaxes) != len(pcs[i].TransferSyntaxes) {
			t.Errorf("pc[%d] transfer syntaxes = %v", i, pc.TransferSyntaxes)
		}
	}
}

// ---------------------------------------------------------------------------
// A-ASSOCIATE-AC / RJ
// ---------------------------------------------------------------------------

func TestAssociateACRoundTrip(t *testing.T) {
	ac := &PDU{
		Type:            PDUAssociateAC,
		ProtocolVersion: 1,
		CalledAETitle:   "FREEMED",
		CallingAETitle:  "SCU_AE",
		PresentationContexts: []PresentationContext{
			{ID: 1, Result: PresContextAcceptance, TransferSyntaxes: []string{ImplicitVRLittleEndian}},
			{ID: 3, Result: PresContextAbstractSyntaxUnsup, TransferSyntaxes: []string{ImplicitVRLittleEndian}},
			{ID: 5, Result: PresContextTransferSyntaxUnsup, TransferSyntaxes: []string{ImplicitVRLittleEndian}},
		},
		MaxPDULength: DefaultMaxPDULength,
	}
	raw, err := ac.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, err := ReadPDU(bytes.NewReader(raw), MaxPDUBodySize)
	if err != nil {
		t.Fatalf("ReadPDU: %v", err)
	}
	if got.Type != PDUAssociateAC {
		t.Fatalf("type = 0x%02X", got.Type)
	}
	if got.CalledAETitle != "FREEMED" || got.CallingAETitle != "SCU_AE" {
		t.Errorf("AE titles = %q / %q", got.CalledAETitle, got.CallingAETitle)
	}
	if len(got.PresentationContexts) != 3 {
		t.Fatalf("presentation contexts = %d", len(got.PresentationContexts))
	}
	wantResults := []byte{PresContextAcceptance, PresContextAbstractSyntaxUnsup, PresContextTransferSyntaxUnsup}
	for i, pc := range got.PresentationContexts {
		if pc.Result != wantResults[i] {
			t.Errorf("pc[%d] result = 0x%02X, want 0x%02X", i, pc.Result, wantResults[i])
		}
		if len(pc.TransferSyntaxes) != 1 || pc.TransferSyntaxes[0] != ImplicitVRLittleEndian {
			t.Errorf("pc[%d] transfer syntax = %v", i, pc.TransferSyntaxes)
		}
	}
	if got.MaxPDULength != DefaultMaxPDULength {
		t.Errorf("max PDU length = %d", got.MaxPDULength)
	}
}

func TestAssociateRJRJRoundTrip(t *testing.T) {
	for _, tc := range []struct{ result, source, reason byte }{
		{RJResultPermanent, RJSourceServiceUser, RJReasonCalledAENotRecog},
		{RJResultTransient, RJSourceServiceProvider, RJReasonNoReason},
	} {
		rj := NewAssociateRJ(tc.result, tc.source, tc.reason)
		raw, err := rj.Encode()
		if err != nil {
			t.Fatalf("Encode: %v", err)
		}
		if len(raw) != 10 {
			t.Errorf("RJ PDU length = %d, want 10", len(raw))
		}
		got, err := ReadPDU(bytes.NewReader(raw), MaxPDUBodySize)
		if err != nil {
			t.Fatalf("ReadPDU: %v", err)
		}
		if got.Result != tc.result || got.Source != tc.source || got.Reason != tc.reason {
			t.Errorf("RJ = (%d,%d,%d), want (%d,%d,%d)", got.Result, got.Source, got.Reason, tc.result, tc.source, tc.reason)
		}
	}
}

// ---------------------------------------------------------------------------
// P-DATA-TF / release / abort
// ---------------------------------------------------------------------------

func TestDataTFRoundTrip(t *testing.T) {
	pdvs := []PDV{
		{ContextID: 1, IsCommand: true, IsLast: false, Data: []byte{0x00, 0x00, 0x00, 0x00}},
		{ContextID: 1, IsCommand: true, IsLast: true, Data: []byte{0x20, 0x00}},
		{ContextID: 1, IsCommand: false, IsLast: false, Data: []byte("ABCD")},
		{ContextID: 1, IsCommand: false, IsLast: true, Data: []byte("EF")},
	}
	raw, err := NewDataTF(pdvs...).Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, err := ReadPDU(bytes.NewReader(raw), MaxPDUBodySize)
	if err != nil {
		t.Fatalf("ReadPDU: %v", err)
	}
	if got.Type != PDUDataTF {
		t.Fatalf("type = 0x%02X", got.Type)
	}
	if len(got.PDVs) != len(pdvs) {
		t.Fatalf("PDVs = %d, want %d", len(got.PDVs), len(pdvs))
	}
	for i, want := range pdvs {
		g := got.PDVs[i]
		if g.ContextID != want.ContextID {
			t.Errorf("PDV[%d] context = %d", i, g.ContextID)
		}
		if g.IsCommand != want.IsCommand {
			t.Errorf("PDV[%d] IsCommand = %v, want %v", i, g.IsCommand, want.IsCommand)
		}
		if g.IsLast != want.IsLast {
			t.Errorf("PDV[%d] IsLast = %v, want %v", i, g.IsLast, want.IsLast)
		}
		if !bytes.Equal(g.Data, want.Data) {
			t.Errorf("PDV[%d] data = %x, want %x", i, g.Data, want.Data)
		}
	}
}

func TestDataTFMessageControlHeaderBytes(t *testing.T) {
	// A command PDV carries bit 0 set; the last fragment carries bit 1.
	raw, err := NewDataTF(
		PDV{ContextID: 2, IsCommand: true, IsLast: false, Data: []byte{0xAA}},
		PDV{ContextID: 2, IsCommand: true, IsLast: true, Data: []byte{0xBB}},
		PDV{ContextID: 2, IsCommand: false, IsLast: true, Data: []byte{0xCC}},
	).Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	body := raw[6:]
	// PDV 1: length(4) ctx(1) mch(1) data(1)
	if body[5] != MCHCommandBit {
		t.Errorf("PDV 1 message control header = 0x%02X, want 0x01", body[5])
	}
	off := 4 + 3
	if body[off+5] != MCHCommandBit|MCHLastBit {
		t.Errorf("PDV 2 message control header = 0x%02X, want 0x03", body[off+5])
	}
	off += 4 + 3
	if body[off+5] != MCHLastBit {
		t.Errorf("PDV 3 message control header = 0x%02X, want 0x02", body[off+5])
	}
}

func TestReleaseAndAbortRoundTrip(t *testing.T) {
	for _, pdu := range []*PDU{NewReleaseRQ(), NewReleaseRP()} {
		raw, err := pdu.Encode()
		if err != nil {
			t.Fatalf("Encode: %v", err)
		}
		if len(raw) != 10 {
			t.Fatalf("release PDU length = %d, want 10", len(raw))
		}
		got, err := ReadPDU(bytes.NewReader(raw), MaxPDUBodySize)
		if err != nil {
			t.Fatalf("ReadPDU: %v", err)
		}
		if got.Type != pdu.Type {
			t.Errorf("release type = 0x%02X, want 0x%02X", got.Type, pdu.Type)
		}
	}

	raw, err := NewAbort(AbortSourceServiceUser, 0x00).Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, err := ReadPDU(bytes.NewReader(raw), MaxPDUBodySize)
	if err != nil {
		t.Fatalf("ReadPDU: %v", err)
	}
	if got.Type != PDUAbort || got.AbortSource != AbortSourceServiceUser || got.AbortReason != 0 {
		t.Errorf("abort = %+v", got)
	}
}

func TestPDUEncodeRejectsUnknownType(t *testing.T) {
	if _, err := (&PDU{Type: 0x42}).Encode(); !errors.Is(err, ErrInvalidPDUType) {
		t.Fatalf("err = %v, want ErrInvalidPDUType", err)
	}
	if _, err := DecodePDU(0x42, nil); !errors.Is(err, ErrInvalidPDUType) {
		t.Fatalf("DecodePDU err = %v, want ErrInvalidPDUType", err)
	}
}

// ---------------------------------------------------------------------------
// Hostile input
// ---------------------------------------------------------------------------

func TestReadPDURejectsHostileLengths(t *testing.T) {
	tests := []struct {
		name string
		raw  []byte
	}{
		{"truncated header", []byte{0x01, 0x00, 0x00}},
		{"empty", nil},
		{
			// 2 GiB body claimed with no body present.
			"absurd length",
			[]byte{PDUAssociateRQ, 0x00, 0x7F, 0xFF, 0xFF, 0xFF},
		},
		{
			// 0xFFFFFFFF is "negative" as a signed 32-bit length.
			"negative length",
			[]byte{PDUAssociateRQ, 0x00, 0xFF, 0xFF, 0xFF, 0xFF},
		},
		{
			// Length past 16 MiB ceiling.
			"above ceiling",
			[]byte{PDUAssociateRQ, 0x00, 0x02, 0x00, 0x00, 0x01},
		},
		{
			// Declared body longer than the available data.
			"short body",
			[]byte{PDUAssociateRQ, 0x00, 0x00, 0x00, 0x10, 0x00, 0x01, 0x00, 0x00},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("ReadPDU panicked on hostile input: %v", r)
					}
				}()
				if _, err := ReadPDU(bytes.NewReader(tc.raw), MaxPDUBodySize); err == nil {
					t.Fatalf("ReadPDU accepted hostile input %x", tc.raw)
				}
			}()
		})
	}
}

func TestDecodePDURejectsMalformedBodies(t *testing.T) {
	tests := []struct {
		name string
		typ  byte
		body []byte
	}{
		{"short RJ", PDUAssociateRJ, []byte{0x00, 0x01}},
		{"short release", PDUReleaseRQ, []byte{0x00, 0x00}},
		{"short abort", PDUAbort, []byte{0x00}},
		{"short association", PDUAssociateRQ, []byte{0x00, 0x01, 0x00, 0x00}},
		{
			"item length beyond buffer",
			PDUAssociateRQ,
			func() []byte {
				b := make([]byte, 4+16+16+32)
				b[1] = 1
				out := append(b, itemApplicationContext, 0x00, 0xFF) // claims 255 bytes
				return out
			}(),
		},
		{
			"truncated item header",
			PDUAssociateRQ,
			func() []byte {
				b := make([]byte, 4+16+16+32)
				b[1] = 1
				return append(b, 0x10, 0x00)
			}(),
		},
		{
			"PDV length smaller than its own header",
			PDUDataTF,
			[]byte{0x00, 0x00, 0x00, 0x01, 0x01, 0x01},
		},
		{
			"PDV claims more than the PDU holds",
			PDUDataTF,
			[]byte{0x00, 0x00, 0x10, 0x00, 0x01, 0x01, 0xAA, 0xBB},
		},
		{
			"PDV header truncated",
			PDUDataTF,
			[]byte{0x00, 0x00, 0x00, 0x03, 0x01},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("DecodePDU panicked on %s: %v", tc.name, r)
				}
			}()
			if _, err := DecodePDU(tc.typ, tc.body); err == nil {
				t.Fatalf("DecodePDU accepted malformed body for %s", tc.name)
			}
		})
	}
}

func TestDecodePDUAcceptsEmptyDataTF(t *testing.T) {
	pdu, err := DecodePDU(PDUDataTF, nil)
	if err != nil {
		t.Fatalf("DecodePDU: %v", err)
	}
	if len(pdu.PDVs) != 0 {
		t.Errorf("PDVs = %d, want 0", len(pdu.PDVs))
	}
}

func TestAETitlePadding(t *testing.T) {
	padded := padAETitle("FREEMED")
	if len(padded) != 16 {
		t.Fatalf("padded length = %d", len(padded))
	}
	if string(padded) != "FREEMED         " {
		t.Errorf("padded = %q", string(padded))
	}
	if got := readAETitle(padded); got != "FREEMED" {
		t.Errorf("readAETitle = %q", got)
	}
	// Longer than 16 characters is truncated rather than overrunning the field.
	long := strings.Repeat("X", 40)
	if got := len(padAETitle(long)); got != 16 {
		t.Errorf("oversized AE title encodes to %d bytes, want 16", got)
	}
	// Space and NUL padding is stripped.
	if got := readAETitle([]byte("AET\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")); got != "AET" {
		t.Errorf("readAETitle with NUL padding = %q", got)
	}
}
