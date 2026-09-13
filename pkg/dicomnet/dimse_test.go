package dicomnet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// Command set construction / parsing
// ---------------------------------------------------------------------------

func TestCommandSetRoundTrip(t *testing.T) {
	cs := NewCFindRequest(0x1234, UIDModalityWorklistFIND, 0)
	raw, err := cs.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	got, err := ParseCommandSet(raw)
	if err != nil {
		t.Fatalf("ParseCommandSet: %v", err)
	}
	if v, ok := got.CommandField(); !ok || v != CmdCFindRQ {
		t.Errorf("command field = 0x%04X/%v, want 0x%04X", v, ok, CmdCFindRQ)
	}
	if v, ok := got.MessageID(); !ok || v != 0x1234 {
		t.Errorf("message ID = 0x%04X, want 0x1234", v)
	}
	if v, ok := got.UI(TagRequestedSOPClassUID); !ok || v != UIDModalityWorklistFIND {
		t.Errorf("requested SOP class = %q", v)
	}
	if got.IsResponse() {
		t.Error("C-FIND-RQ classified as a response")
	}
	if !got.HasDataSet() {
		t.Error("C-FIND-RQ with a data set reported no data set")
	}
	if !IsCommandTag(TagCommandField) || IsCommandTag(TagPatientName) {
		t.Error("IsCommandTag misclassified a tag")
	}
	if vr := CommandVR(TagStatus); vr != "US" {
		t.Errorf("CommandVR(Status) = %q", vr)
	}
	if vr := CommandVR(0x00009999); vr != "UN" {
		t.Errorf("CommandVR(unknown) = %q", vr)
	}
}

// TestCommandSetMatchesReferenceBytes pins the implicit VR wire layout
// including the (0000,0000) Command Group Length element.
func TestCommandSetMatchesReferenceBytes(t *testing.T) {
	cs := NewCommandSet()
	cs.SetUS(TagCommandField, CmdCFindRQ)
	cs.SetUS(TagMessageID, 1)
	cs.SetUS(TagCommandDataSetType, DataSetTypePresent)
	cs.SetUI(TagRequestedSOPClassUID, "1.2.840.10008.5.1.4.31")

	got, err := cs.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	var want []byte
	want = append(want, refImplicitElement(0x0000, 0x0003, []byte("1.2.840.10008.5.1.4.31"))...)
	want = append(want, refImplicitElement(0x0000, 0x0100, []byte{0x20, 0x00})...)
	want = append(want, refImplicitElement(0x0000, 0x0110, []byte{0x01, 0x00})...)
	want = append(want, refImplicitElement(0x0000, 0x0800, []byte{0x01, 0x00})...)

	var gl []byte
	gl = append(gl, refImplicitElement(0x0000, 0x0000, make([]byte, 4))...)
	binary.LittleEndian.PutUint32(gl[8:12], uint32(len(want)))
	want = append(gl, want...)

	if !bytes.Equal(got, want) {
		t.Fatalf("command set bytes differ\n got=%x\nwant=%x", got, want)
	}
}

func TestCommandSetStatusesAndDataSetFlags(t *testing.T) {
	cases := []struct {
		status     uint16
		identifier bool
	}{
		{StatusPending, true},
		{StatusSuccess, false},
		{StatusCancel, false},
		{StatusRefusedResources, false},
		{StatusDataSetMismatch, false},
	}
	for _, tc := range cases {
		resp := NewCFindResponse(9, UIDModalityWorklistFIND, tc.status, tc.identifier)
		raw, err := resp.Encode()
		if err != nil {
			t.Fatalf("status 0x%04X Encode: %v", tc.status, err)
		}
		got, err := ParseCommandSet(raw)
		if err != nil {
			t.Fatalf("status 0x%04X ParseCommandSet: %v", tc.status, err)
		}
		if v, ok := got.Status(); !ok || v != tc.status {
			t.Errorf("status = 0x%04X, want 0x%04X", v, tc.status)
		}
		if v, ok := got.MessageIDBeingRespondedTo(); !ok || v != 9 {
			t.Errorf("message ID being responded to = %d", v)
		}
		if cf, _ := got.CommandField(); cf != CmdCFindRSP {
			t.Errorf("command field = 0x%04X, want 0x%04X", cf, CmdCFindRSP)
		}
		if !got.IsResponse() {
			t.Error("C-FIND-RSP not classified as a response")
		}
		if got.HasDataSet() != tc.identifier {
			t.Errorf("HasDataSet = %v, want %v", got.HasDataSet(), tc.identifier)
		}
		if v, _ := got.US(TagCommandDataSetType); v != DataSetTypeNone && !tc.identifier {
			t.Errorf("data set type = 0x%04X, want 0x%04X", v, DataSetTypeNone)
		}
		if v, _ := got.UI(TagAffectedSOPClassUID); v != UIDModalityWorklistFIND {
			t.Errorf("affected SOP class = %q", v)
		}
	}
}

func TestCEchoResponseRoundTrip(t *testing.T) {
	raw, err := NewCEchoResponse(7, UIDVerification, StatusSuccess).Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, err := ParseCommandSet(raw)
	if err != nil {
		t.Fatalf("ParseCommandSet: %v", err)
	}
	if cf, _ := got.CommandField(); cf != CmdCEchoRSP {
		t.Errorf("command field = 0x%04X", cf)
	}
	if got.HasDataSet() {
		t.Error("C-ECHO-RSP must not carry a data set")
	}
}

func TestParseCommandSetRejectsBadInput(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", nil},
		{"truncated element header", []byte{0x00, 0x00, 0x00, 0x00, 0x04}},
		{
			"data element in command set",
			refImplicitElement(0x0010, 0x0010, []byte("SMITH")),
		},
		{
			"command field missing",
			refImplicitElement(0x0000, 0x0110, []byte{0x01, 0x00}),
		},
		{
			"group length larger than the buffer",
			func() []byte {
				b := refImplicitElement(0x0000, 0x0000, []byte{0xFF, 0xFF, 0xFF, 0x7F})
				return append(b, refImplicitElement(0x0000, 0x0100, []byte{0x20, 0x00})...)
			}(),
		},
		{
			"value length past the end of the buffer",
			[]byte{0x00, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 'A'},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("ParseCommandSet panicked: %v", r)
				}
			}()
			if _, err := ParseCommandSet(tc.data); err == nil {
				t.Fatalf("ParseCommandSet accepted malformed command set %x", tc.data)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Message control header / PDV discrimination
// ---------------------------------------------------------------------------

func TestMessageControlHeaderBits(t *testing.T) {
	cases := []struct {
		cmd, last bool
		want      byte
	}{
		{false, false, 0x00},
		{true, false, 0x01},
		{false, true, 0x02},
		{true, true, 0x03},
	}
	for _, tc := range cases {
		got := MessageControlHeader(tc.cmd, tc.last)
		if got != tc.want {
			t.Errorf("MessageControlHeader(%v,%v) = 0x%02X, want 0x%02X", tc.cmd, tc.last, got, tc.want)
		}
		pdv := PDV{IsCommand: tc.cmd, IsLast: tc.last}
		if pdv.ControlHeader() != tc.want {
			t.Errorf("PDV.ControlHeader = 0x%02X, want 0x%02X", pdv.ControlHeader(), tc.want)
		}
		if pdv.IsCommand != (got&MCHCommandBit != 0) {
			t.Error("command bit round trip failed")
		}
		if pdv.IsLast != (got&MCHLastBit != 0) {
			t.Error("last bit round trip failed")
		}
	}
}

// TestDataTFDistinguishesCommandAndDataPDVs checks the message control header
// survives a full PDU encode/decode cycle.
func TestDataTFDistinguishesCommandAndDataPDVs(t *testing.T) {
	raw, err := NewDataTF(
		PDV{ContextID: 3, IsCommand: true, IsLast: true, Data: []byte("CMD")},
		PDV{ContextID: 3, IsCommand: false, IsLast: true, Data: []byte("DATA")},
	).Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pdu, err := DecodePDU(PDUDataTF, raw[6:])
	if err != nil {
		t.Fatalf("DecodePDU: %v", err)
	}
	if len(pdu.PDVs) != 2 {
		t.Fatalf("PDVs = %d, want 2", len(pdu.PDVs))
	}
	if !pdu.PDVs[0].IsCommand || string(pdu.PDVs[0].Data) != "CMD" {
		t.Errorf("PDV 0 = %+v", pdu.PDVs[0])
	}
	if pdu.PDVs[1].IsCommand || string(pdu.PDVs[1].Data) != "DATA" {
		t.Errorf("PDV 1 = %+v", pdu.PDVs[1])
	}
	if pdu.PDVs[0].ContextID != 3 || pdu.PDVs[1].ContextID != 3 {
		t.Errorf("context IDs = %d/%d", pdu.PDVs[0].ContextID, pdu.PDVs[1].ContextID)
	}
}

// ---------------------------------------------------------------------------
// Fragmentation
// ---------------------------------------------------------------------------

func TestFragmentRespectsPeerMaxPDU(t *testing.T) {
	command, err := NewCFindRequest(1, UIDModalityWorklistFIND, 0).Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// A data set much larger than the peer's maximum PDU length.
	dataSet := bytes.Repeat([]byte("A"), 4000)

	for _, peerMax := range []uint32{512, 1024, DefaultMaxPDULength, 0} {
		pdvs, err := Fragment(1, command, dataSet, peerMax)
		if err != nil {
			t.Fatalf("peerMax=%d Fragment: %v", peerMax, err)
		}
		limit := int(peerMax)
		if peerMax == 0 {
			limit = int(DefaultMaxPDULength)
		}
		for i, pdv := range pdvs {
			pduLen := pduHeaderSize + 4 + len(pdv.Data) + 2
			if pduLen > limit {
				t.Errorf("peerMax=%d: PDV %d would produce a %d byte PDU (limit %d)", peerMax, i, pduLen, limit)
			}
			if pdv.ContextID != 1 {
				t.Errorf("peerMax=%d: PDV %d context = %d", peerMax, i, pdv.ContextID)
			}
		}
		// The final command PDV and the final data PDV must be marked last.
		if !pdvs[0].IsCommand || !pdvs[0].IsLast && len(command) <= len(pdvs[0].Data) {
			t.Errorf("peerMax=%d: unexpected first PDV %+v", peerMax, pdvs[0])
		}
		last := pdvs[len(pdvs)-1]
		if last.IsCommand || !last.IsLast {
			t.Errorf("peerMax=%d: last PDV = %+v, want a final data PDV", peerMax, last)
		}

		// Reassembling the PDVs must reproduce the original message.
		acc := NewMessageAccumulator()
		for _, pdv := range pdvs {
			if err := acc.Feed(pdv); err != nil {
				t.Fatalf("peerMax=%d: Feed: %v", peerMax, err)
			}
		}
		if !acc.Complete() {
			t.Errorf("peerMax=%d: message not complete", peerMax)
		}
		if !bytes.Equal(acc.DataSetBytes(), dataSet) {
			t.Errorf("peerMax=%d: data set did not survive fragmentation", peerMax)
		}
		cs := acc.Command()
		if cs == nil {
			t.Fatalf("peerMax=%d: command set did not survive fragmentation", peerMax)
		}
		if v, _ := cs.MessageID(); v != 1 {
			t.Errorf("peerMax=%d: message ID = %d", peerMax, v)
		}
	}
}

func TestFragmentRejectsAbsurdlySmallPDU(t *testing.T) {
	if _, err := Fragment(1, []byte{0x01}, nil, 8); err == nil {
		t.Fatal("Fragment accepted a maximum PDU length smaller than its own headers")
	}
}

func TestFragmentWithoutDataSet(t *testing.T) {
	command, err := NewCFindResponse(1, UIDModalityWorklistFIND, StatusSuccess, false).Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pdvs, err := Fragment(2, command, nil, 0)
	if err != nil {
		t.Fatalf("Fragment: %v", err)
	}
	if len(pdvs) != 1 {
		t.Fatalf("PDVs = %d, want 1", len(pdvs))
	}
	if !pdvs[0].IsCommand || !pdvs[0].IsLast {
		t.Errorf("PDV = %+v", pdvs[0])
	}
	if !bytes.Equal(pdvs[0].Data, command) {
		t.Error("command bytes were altered by fragmentation")
	}
	acc := NewMessageAccumulator()
	if err := acc.Feed(pdvs[0]); err != nil {
		t.Fatalf("Feed: %v", err)
	}
	if !acc.Complete() {
		t.Error("message without a data set not complete")
	}
	if acc.WantDataSet() {
		t.Error("WantDataSet = true for a message without a data set")
	}
	if len(acc.DataSetBytes()) != 0 {
		t.Errorf("data set bytes = %x", acc.DataSetBytes())
	}
	if v, _ := acc.Command().Status(); v != StatusSuccess {
		t.Errorf("status = 0x%04X", v)
	}
}

func TestFragmentEmptyCommandValue(t *testing.T) {
	pdvs, err := Fragment(1, nil, []byte("X"), 0)
	if err != nil {
		t.Fatalf("Fragment: %v", err)
	}
	if len(pdvs) != 2 {
		t.Fatalf("PDVs = %d, want 2 (empty command + data)", len(pdvs))
	}
	if len(pdvs[0].Data) != 0 || !pdvs[0].IsLast || !pdvs[0].IsCommand {
		t.Errorf("empty command PDV = %+v", pdvs[0])
	}
}

// ---------------------------------------------------------------------------
// Message assembly
// ---------------------------------------------------------------------------

func TestMessageAccumulatorCommandAndData(t *testing.T) {
	command, err := NewCFindRequest(5, UIDModalityWorklistFIND, 0).Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	identifier := []byte("IDENTIFIER-BYTES")

	acc := NewMessageAccumulator()
	// Split both values at arbitrary points to force multi-PDV assembly.
	if err := acc.Feed(PDV{ContextID: 1, IsCommand: true, IsLast: false, Data: command[:5]}); err != nil {
		t.Fatalf("Feed: %v", err)
	}
	if acc.Complete() {
		t.Fatal("message reported complete before the command set finished")
	}
	if err := acc.Feed(PDV{ContextID: 1, IsCommand: true, IsLast: true, Data: command[5:]}); err != nil {
		t.Fatalf("Feed: %v", err)
	}
	if !acc.CommandComplete() {
		t.Fatal("command set not complete")
	}
	if acc.Complete() {
		t.Fatal("message reported complete before the data set arrived")
	}
	if !acc.WantDataSet() {
		t.Fatal("data set expected but not reported")
	}
	if err := acc.Feed(PDV{ContextID: 1, IsCommand: false, IsLast: true, Data: identifier}); err != nil {
		t.Fatalf("Feed: %v", err)
	}
	if !acc.Complete() {
		t.Fatal("message not complete")
	}
	if !bytes.Equal(acc.DataSetBytes(), identifier) {
		t.Errorf("identifier = %q", acc.DataSetBytes())
	}
	if v, _ := acc.Command().MessageID(); v != 5 {
		t.Errorf("message ID = %d", v)
	}
}

func TestMessageAccumulatorRejectsProtocolErrors(t *testing.T) {
	command, err := NewCFindRequest(1, UIDModalityWorklistFIND, 0).Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	noDataSet, err := NewCFindResponse(1, UIDModalityWorklistFIND, StatusSuccess, false).Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	t.Run("data PDV before the command set", func(t *testing.T) {
		acc := NewMessageAccumulator()
		if err := acc.Feed(PDV{ContextID: 1, IsLast: true, Data: []byte("X")}); err == nil {
			t.Fatal("accepted a data PDV before the command set")
		}
	})

	t.Run("command PDV after the command set", func(t *testing.T) {
		acc := NewMessageAccumulator()
		if err := acc.Feed(PDV{ContextID: 1, IsCommand: true, IsLast: true, Data: command}); err != nil {
			t.Fatalf("Feed: %v", err)
		}
		if err := acc.Feed(PDV{ContextID: 1, IsCommand: true, IsLast: true, Data: command}); err == nil {
			t.Fatal("accepted a second command set in one message")
		}
	})

	t.Run("data PDV when no data set was announced", func(t *testing.T) {
		acc := NewMessageAccumulator()
		if err := acc.Feed(PDV{ContextID: 1, IsCommand: true, IsLast: true, Data: noDataSet}); err != nil {
			t.Fatalf("Feed: %v", err)
		}
		if err := acc.Feed(PDV{ContextID: 1, IsLast: true, Data: []byte("X")}); err == nil {
			t.Fatal("accepted an unsolicited data set")
		}
	})

	t.Run("malformed command set", func(t *testing.T) {
		acc := NewMessageAccumulator()
		if err := acc.Feed(PDV{ContextID: 1, IsCommand: true, IsLast: true, Data: []byte{0x01}}); err == nil {
			t.Fatal("accepted a malformed command set")
		}
	})

	t.Run("oversized message", func(t *testing.T) {
		acc := NewMessageAccumulator()
		acc.total = MaxElementLength
		if err := acc.Feed(PDV{ContextID: 1, IsCommand: true, IsLast: true, Data: command}); !errors.Is(err, ErrMessageTooLarge) {
			t.Fatalf("err = %v, want ErrMessageTooLarge", err)
		}
	})
}
