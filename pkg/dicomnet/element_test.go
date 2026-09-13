package dicomnet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
	"testing"
)

// refExplicitElement builds an Explicit VR Little Endian element by hand.
func refExplicitElement(group, elem uint16, vr string, value []byte) []byte {
	out := make([]byte, 4)
	binary.LittleEndian.PutUint16(out[0:2], group)
	binary.LittleEndian.PutUint16(out[2:4], elem)
	out = append(out, vr...)
	if isLongVR(vr) {
		out = append(out, 0, 0)
		var l [4]byte
		binary.LittleEndian.PutUint32(l[:], uint32(len(value)))
		out = append(out, l[:]...)
	} else {
		var l [2]byte
		binary.LittleEndian.PutUint16(l[:], uint16(len(value)))
		out = append(out, l[:]...)
	}
	return append(out, value...)
}

// refImplicitElement builds an Implicit VR Little Endian element by hand.
func refImplicitElement(group, elem uint16, value []byte) []byte {
	out := make([]byte, 8)
	binary.LittleEndian.PutUint16(out[0:2], group)
	binary.LittleEndian.PutUint16(out[2:4], elem)
	binary.LittleEndian.PutUint32(out[4:8], uint32(len(value)))
	return append(out, value...)
}

// refItemTag builds an Item / Delimitation / Sequence-Delimitation tag. DICOM
// writes the group number before the element number.
func refItemTag(tag uint32, length uint32) []byte {
	out := make([]byte, 8)
	binary.LittleEndian.PutUint16(out[0:2], uint16(tag>>16))
	binary.LittleEndian.PutUint16(out[2:4], uint16(tag))
	binary.LittleEndian.PutUint32(out[4:8], length)
	return out
}

// sampleDataSet returns a data set exercising text VRs, a zero-length return
// key, a US and a UL element.
func sampleDataSet() DataSet {
	return DataSet{
		NewStringElement(TagPatientName, "PN", "SMITH^JOHN"),
		NewStringElement(TagPatientID, "LO", "MRN0001234"),
		NewStringElement(TagPatientBirthDate, "DA", "19700101"),
		NewStringElement(TagPatientSex, "CS", "M"),
		NewStringElement(TagAccessionNumber, "SH", ""), // return key, zero length
		NewStringElement(TagStudyInstanceUID, "UI", "1.2.840.10008.5.1.4.31.1"),
		NewUSElement(TagSPSID, 42),
		NewULElement(0x00090010, 0xDEADBEEF),
	}
}

func TestElementExplicitVRLittleEndianRoundTrip(t *testing.T) {
	ds := sampleDataSet()
	raw, err := EncodeDataSet(ds, true)
	if err != nil {
		t.Fatalf("EncodeDataSet: %v", err)
	}
	got, err := DecodeDataSet(raw, true)
	if err != nil {
		t.Fatalf("DecodeDataSet: %v", err)
	}
	if len(got) != len(ds) {
		t.Fatalf("elements = %d, want %d", len(got), len(ds))
	}
	for i, want := range ds {
		if got[i].Tag != want.Tag {
			t.Errorf("element %d tag = 0x%08X, want 0x%08X", i, got[i].Tag, want.Tag)
		}
		if got[i].VR != want.VR {
			t.Errorf("element %d VR = %q, want %q", i, got[i].VR, want.VR)
		}
		if !bytes.Equal(got[i].Value, want.Value) {
			t.Errorf("element %d value = %x, want %x", i, got[i].Value, want.Value)
		}
	}
}

func TestElementImplicitVRLittleEndianRoundTrip(t *testing.T) {
	ds := sampleDataSet()
	raw, err := EncodeDataSet(ds, false)
	if err != nil {
		t.Fatalf("EncodeDataSet: %v", err)
	}
	got, err := DecodeDataSet(raw, false)
	if err != nil {
		t.Fatalf("DecodeDataSet: %v", err)
	}
	if len(got) != len(ds) {
		t.Fatalf("elements = %d, want %d", len(got), len(ds))
	}
	for i, want := range ds {
		if got[i].Tag != want.Tag {
			t.Errorf("element %d tag = 0x%08X, want 0x%08X", i, got[i].Tag, want.Tag)
		}
		if !bytes.Equal(got[i].Value, want.Value) {
			t.Errorf("element %d value = %x, want %x", i, got[i].Value, want.Value)
		}
	}
	// Implicit VR has no VR on the wire, so the dictionary must recover it.
	if vr := got[0].VR; vr != "PN" {
		t.Errorf("implicit PatientName VR = %q, want PN", vr)
	}
	if vr := got[6].VR; vr != "SH" {
		t.Errorf("implicit SPS ID VR = %q, want SH", vr)
	}
	// Unknown tag decodes as UN with the raw value preserved.
	if vr := got[7].VR; vr != "UN" {
		t.Errorf("unknown tag VR = %q, want UN", vr)
	}
}

// TestExplicitVRLengthQuirks checks the 8-byte vs 12-byte element header
// layout required by PS3.5 §7.1.2.
func TestExplicitVRLengthQuirks(t *testing.T) {
	shortVRs := []string{"AE", "CS", "DA", "LO", "PN", "SH", "TM", "UI", "US"}
	for _, vr := range shortVRs {
		b, err := EncodeElement(NewStringElement(0x00100010, vr, "AB"), true)
		if err != nil {
			t.Fatalf("EncodeElement(%s): %v", vr, err)
		}
		if len(b) != 8+2 {
			t.Errorf("VR %s header length = %d, want 8 (+2 value)", vr, len(b)-2)
		}
		// The 2-byte length field sits at offset 6.
		if got := binary.LittleEndian.Uint16(b[6:8]); got != 2 {
			t.Errorf("VR %s length = %d, want 2", vr, got)
		}
	}

	longVRs := []string{"OB", "OD", "OF", "OL", "OV", "OW", "UC", "UR", "UT", "UN"}
	for _, vr := range longVRs {
		b, err := EncodeElement(Element{Tag: 0x00090010, VR: vr, Value: []byte("A")}, true)
		if err != nil {
			t.Fatalf("EncodeElement(%s): %v", vr, err)
		}
		if len(b) != 12+1 {
			t.Errorf("VR %s header length = %d, want 12 (+1 value)", vr, len(b)-1)
		}
		if b[6] != 0 || b[7] != 0 {
			t.Errorf("VR %s reserved bytes = %02X %02X, want 00 00", vr, b[6], b[7])
		}
		if got := binary.LittleEndian.Uint32(b[8:12]); got != 1 {
			t.Errorf("VR %s 32-bit length = %d, want 1", vr, got)
		}
	}

	// SQ always uses the 4-byte reserved + 32-bit length layout, whether its
	// length is defined or undefined.
	b, err := EncodeElement(NewSequenceElement(0x00400100), true)
	if err != nil {
		t.Fatalf("EncodeElement(SQ): %v", err)
	}
	if len(b) != 12 {
		t.Errorf("empty SQ element length = %d, want 12", len(b))
	}
	if string(b[4:6]) != "SQ" || b[6] != 0 || b[7] != 0 {
		t.Errorf("empty SQ header = %x", b[:12])
	}
	b, err = EncodeElement(Element{Tag: 0x00400100, VR: "SQ", UndefinedLength: true}, true)
	if err != nil {
		t.Fatalf("EncodeElement(SQ undefined): %v", err)
	}
	if got := binary.LittleEndian.Uint32(b[8:12]); got != UndefinedLength {
		t.Errorf("undefined SQ length = 0x%08X, want 0xFFFFFFFF", got)
	}
	if len(b) != 20 || getTag(b[12:16]) != TagSequenceDelimiter {
		t.Errorf("undefined SQ must end with a sequence delimitation item: %x", b)
	}
}

// TestExplicitHeaderMatchesReference cross-checks the encoder against the
// hand-built byte image for a short and a long VR.
func TestExplicitHeaderMatchesReference(t *testing.T) {
	got, err := EncodeElement(NewStringElement(TagPatientName, "PN", "SMITH^JOHN"), true)
	if err != nil {
		t.Fatalf("EncodeElement: %v", err)
	}
	want := refExplicitElement(0x0010, 0x0010, "PN", []byte("SMITH^JOHN"))
	if !bytes.Equal(got, want) {
		t.Errorf("PN element bytes differ\n got=%x\nwant=%x", got, want)
	}

	got, err = EncodeElement(Element{Tag: 0x00080005, VR: "CS", Value: []byte("ISO_IR 100")}, true)
	if err != nil {
		t.Fatalf("EncodeElement: %v", err)
	}
	want = refExplicitElement(0x0008, 0x0005, "CS", []byte("ISO_IR 100"))
	if !bytes.Equal(got, want) {
		t.Errorf("CS element bytes differ\n got=%x\nwant=%x", got, want)
	}

	got, err = EncodeElement(Element{Tag: 0x00090010, VR: "OB", Value: []byte{1, 2, 3}}, true)
	if err != nil {
		t.Fatalf("EncodeElement: %v", err)
	}
	want = refExplicitElement(0x0009, 0x0010, "OB", []byte{1, 2, 3})
	if !bytes.Equal(got, want) {
		t.Errorf("OB element bytes differ\n got=%x\nwant=%x", got, want)
	}
}

func TestImplicitHeaderMatchesReference(t *testing.T) {
	got, err := EncodeElement(NewStringElement(TagPatientName, "PN", "DOE^JANE"), false)
	if err != nil {
		t.Fatalf("EncodeElement: %v", err)
	}
	want := refImplicitElement(0x0010, 0x0010, []byte("DOE^JANE"))
	if !bytes.Equal(got, want) {
		t.Errorf("implicit element bytes differ\n got=%x\nwant=%x", got, want)
	}
}

// ---------------------------------------------------------------------------
// Sequences
// ---------------------------------------------------------------------------

func TestSequenceDefinedLengthRoundTrip(t *testing.T) {
	sps := DataSet{
		NewStringElement(TagModality, "CS", "CT"),
		NewStringElement(TagSPSStartDate, "DA", "20240115"),
		NewStringElement(TagScheduledStationAETitle, "AE", "CT1"),
	}
	ds := DataSet{NewSequenceElement(TagScheduledProcedureStepSeq, sps)}

	for _, explicit := range []bool{true, false} {
		raw, err := EncodeDataSet(ds, explicit)
		if err != nil {
			t.Fatalf("explicit=%v EncodeDataSet: %v", explicit, err)
		}
		got, err := DecodeDataSet(raw, explicit)
		if err != nil {
			t.Fatalf("explicit=%v DecodeDataSet: %v", explicit, err)
		}
		item, ok := got.GetSequence(TagScheduledProcedureStepSeq)
		if !ok {
			t.Fatalf("explicit=%v: ScheduledProcedureStepSequence missing", explicit)
		}
		if v := item.GetString(TagModality); v != "CT" {
			t.Errorf("explicit=%v: Modality = %q", explicit, v)
		}
		if v := item.GetString(TagSPSStartDate); v != "20240115" {
			t.Errorf("explicit=%v: start date = %q", explicit, v)
		}
		if v := item.GetString(TagScheduledStationAETitle); v != "CT1" {
			t.Errorf("explicit=%v: station AE = %q", explicit, v)
		}
		if got[0].UndefinedLength {
			t.Errorf("explicit=%v: defined-length sequence marked undefined", explicit)
		}
	}
}

func TestSequenceUndefinedLengthRoundTrip(t *testing.T) {
	sps := DataSet{NewStringElement(TagModality, "CS", "MR")}
	ds := DataSet{Element{Tag: TagScheduledProcedureStepSeq, VR: "SQ", Items: []DataSet{sps}, UndefinedLength: true}}

	for _, explicit := range []bool{true, false} {
		raw, err := EncodeDataSet(ds, explicit)
		if err != nil {
			t.Fatalf("explicit=%v EncodeDataSet: %v", explicit, err)
		}
		// The delimiter item must be present on the wire.
		if !bytes.Contains(raw, []byte{0xFE, 0xFF, 0xDD, 0xE0}) {
			t.Fatalf("explicit=%v: sequence delimitation item not found in %x", explicit, raw)
		}
		got, err := DecodeDataSet(raw, explicit)
		if err != nil {
			t.Fatalf("explicit=%v DecodeDataSet: %v", explicit, err)
		}
		item, ok := got.GetSequence(TagScheduledProcedureStepSeq)
		if !ok {
			t.Fatalf("explicit=%v: sequence missing", explicit)
		}
		if v := item.GetString(TagModality); v != "MR" {
			t.Errorf("explicit=%v: Modality = %q", explicit, v)
		}
		if !got[0].UndefinedLength {
			t.Errorf("explicit=%v: undefined-length sequence not reported as such", explicit)
		}
	}
}

func TestSequenceUndefinedLengthItemRoundTrip(t *testing.T) {
	// A defined-length sequence whose single item has undefined length,
	// terminated by an Item Delimitation Item.
	item := refItemTag(TagItem, UndefinedLength)
	item = append(item, refExplicitElement(0x0008, 0x0060, "CS", []byte("US"))...)
	item = append(item, refItemTag(TagItemDelimitation, 0)...)
	// refExplicitElement computes the sequence length from the value given.
	raw := refExplicitElement(0x0040, 0x0100, "SQ", item)

	got, err := DecodeDataSet(raw, true)
	if err != nil {
		t.Fatalf("DecodeDataSet: %v", err)
	}
	inner, ok := got.GetSequence(TagScheduledProcedureStepSeq)
	if !ok {
		t.Fatalf("sequence missing, decoded %+v", got)
	}
	if v := inner.GetString(TagModality); v != "US" {
		t.Errorf("Modality = %q, want US", v)
	}
}

func TestEmptySequenceRoundTrip(t *testing.T) {
	ds := DataSet{NewSequenceElement(TagScheduledProcedureStepSeq)}
	for _, explicit := range []bool{true, false} {
		raw, err := EncodeDataSet(ds, explicit)
		if err != nil {
			t.Fatalf("explicit=%v EncodeDataSet: %v", explicit, err)
		}
		got, err := DecodeDataSet(raw, explicit)
		if err != nil {
			t.Fatalf("explicit=%v DecodeDataSet: %v", explicit, err)
		}
		if len(got) != 1 || len(got[0].Items) != 0 {
			t.Errorf("explicit=%v: empty sequence = %+v", explicit, got)
		}
	}
}

// ---------------------------------------------------------------------------
// Multi-valued VRs and character set passthrough
// ---------------------------------------------------------------------------

func TestMultiValuedVRRoundTrip(t *testing.T) {
	ds := DataSet{NewStringElement(TagPatientName, "PN", "SMITH^JOHN\\DOE^JANE")}
	raw, err := EncodeDataSet(ds, true)
	if err != nil {
		t.Fatalf("EncodeDataSet: %v", err)
	}
	got, err := DecodeDataSet(raw, true)
	if err != nil {
		t.Fatalf("DecodeDataSet: %v", err)
	}
	if got[0].String() != "SMITH^JOHN\\DOE^JANE" {
		t.Fatalf("value = %q", got[0].String())
	}
	values := got[0].Values()
	if len(values) != 2 || values[0] != "SMITH^JOHN" || values[1] != "DOE^JANE" {
		t.Errorf("values = %v", values)
	}
	if JoinValues(values) != "SMITH^JOHN\\DOE^JANE" {
		t.Errorf("JoinValues = %q", JoinValues(values))
	}
	if SplitValues("") != nil {
		t.Errorf("SplitValues(\"\") = %v, want nil", SplitValues(""))
	}
}

func TestCharacterSetPassthrough(t *testing.T) {
	name := "MÜLLER^JÖRG"
	ds := DataSet{NewStringElement(TagPatientName, "PN", name)}
	for _, explicit := range []bool{true, false} {
		raw, err := EncodeDataSet(ds, explicit)
		if err != nil {
			t.Fatalf("EncodeDataSet: %v", err)
		}
		got, err := DecodeDataSet(raw, explicit)
		if err != nil {
			t.Fatalf("DecodeDataSet: %v", err)
		}
		if got[0].String() != name {
			t.Errorf("explicit=%v: name = %q, want %q", explicit, got[0].String(), name)
		}
	}
}

func TestTrailingPaddingStripped(t *testing.T) {
	e := NewStringElement(TagPatientID, "LO", "MRN123   \x00")
	if got := e.String(); got != "MRN123" {
		t.Errorf("String() = %q, want MRN123", got)
	}
}

// ---------------------------------------------------------------------------
// Malformed / hostile input
// ---------------------------------------------------------------------------

func TestDecodeDataSetMalformed(t *testing.T) {
	tests := []struct {
		name     string
		explicit bool
		data     []byte
	}{
		{
			name:     "truncated element header",
			explicit: true,
			data:     []byte{0x10, 0x00, 0x10, 0x00, 'P'},
		},
		{
			name:     "truncated implicit element header",
			explicit: false,
			data:     []byte{0x10, 0x00, 0x10, 0x00, 0x02, 0x00},
		},
		{
			name:     "huge element length implicit",
			explicit: false,
			data:     func() []byte { return refItemTag(0x00100010, 0x7FFFFFFF) }(),
		},
		{
			name:     "huge element length explicit short VR",
			explicit: true,
			data:     []byte{0x10, 0x00, 0x10, 0x00, 'P', 'N', 0xFF, 0xFF},
		},
		{
			name:     "value length past the end of the buffer",
			explicit: true,
			data:     []byte{0x10, 0x00, 0x10, 0x00, 'P', 'N', 0x20, 0x00, 'A'},
		},
		{
			name:     "unknown VR",
			explicit: true,
			data:     []byte{0x10, 0x00, 0x10, 0x00, 'Z', 'Z', 0x00, 0x00},
		},
		{
			name:     "zero-length item tag at data set level",
			explicit: true,
			data:     refItemTag(TagItem, 0),
		},
		{
			name:     "item delimiter at data set level",
			explicit: false,
			data:     refItemTag(TagItemDelimitation, 0),
		},
		{
			name:     "bogus group FFFE tag",
			explicit: false,
			data:     refItemTag(0xFFFEE999, 0),
		},
		{
			name:     "unterminated undefined-length sequence",
			explicit: true,
			data: func() []byte {
				var b []byte
				b = append(b, 0x40, 0x00, 0x00, 0x01)
				b = append(b, "SQ"...)
				b = append(b, 0, 0)
				l := make([]byte, 4)
				binary.LittleEndian.PutUint32(l, UndefinedLength)
				b = append(b, l...)
				// one item, no sequence delimiter
				b = append(b, refItemTag(TagItem, 0)...)
				return b
			}(),
		},
		{
			name:     "sequence item length past the sequence end",
			explicit: true,
			data: func() []byte {
				var b []byte
				b = append(b, 0x40, 0x00, 0x00, 0x01)
				b = append(b, "SQ"...)
				b = append(b, 0, 0)
				l := make([]byte, 4)
				binary.LittleEndian.PutUint32(l, 8)
				b = append(b, l...)
				b = append(b, refItemTag(TagItem, 0x1000)...)
				return b
			}(),
		},
		{
			name:     "undefined length on a non-SQ explicit element",
			explicit: true,
			data: func() []byte {
				b := []byte{0x10, 0x00, 0x10, 0x00, 'P', 'N', 0, 0}
				l := make([]byte, 4)
				binary.LittleEndian.PutUint32(l, UndefinedLength)
				return append(b, l...)
			}(),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("DecodeDataSet panicked: %v", r)
				}
			}()
			if _, err := DecodeDataSet(tc.data, tc.explicit); err == nil {
				t.Fatalf("DecodeDataSet accepted malformed input %x", tc.data)
			}
		})
	}
}

func TestDecodeDataSetEmptyInput(t *testing.T) {
	for _, explicit := range []bool{true, false} {
		ds, err := DecodeDataSet(nil, explicit)
		if err != nil {
			t.Fatalf("explicit=%v: %v", explicit, err)
		}
		if len(ds) != 0 {
			t.Errorf("explicit=%v: %d elements, want 0", explicit, len(ds))
		}
	}
}

func TestEncodeElementRejectsOversizedValue(t *testing.T) {
	huge := make([]byte, MaxElementLength+1)
	if _, err := EncodeElement(Element{Tag: 0x00090010, VR: "OB", Value: huge}, true); !errors.Is(err, ErrElementTooLarge) {
		t.Fatalf("err = %v, want ErrElementTooLarge", err)
	}
	if _, err := EncodeElement(Element{Tag: 0x00100010, VR: "PN", Value: make([]byte, 0x10000)}, true); !errors.Is(err, ErrElementTooLarge) {
		t.Fatalf("short VR overflow err = %v, want ErrElementTooLarge", err)
	}
}

func TestEncodeElementRejectsUnknownVR(t *testing.T) {
	if _, err := EncodeElement(Element{Tag: 0x00100010, VR: "ZZ", Value: []byte("x")}, true); !errors.Is(err, ErrInvalidVR) {
		t.Fatalf("err = %v, want ErrInvalidVR", err)
	}
}

func TestDataSetAccessors(t *testing.T) {
	ds := sampleDataSet()
	if v := ds.GetString(TagPatientName); v != "SMITH^JOHN" {
		t.Errorf("GetString(PatientName) = %q", v)
	}
	if v := ds.GetString(TagStudyDescription); v != "" {
		t.Errorf("GetString(missing) = %q", v)
	}
	if _, ok := ds.Get(0x99999999); ok {
		t.Error("Get(unknown) reported ok")
	}
	e, ok := ds.Get(TagSPSID)
	if !ok {
		t.Fatal("Get(SPSID) failed")
	}
	if v, ok := e.Uint16(); !ok || v != 42 {
		t.Errorf("SPSID Uint16 = %d/%v", v, ok)
	}
	ul, ok := ds.Get(0x00090010)
	if !ok {
		t.Fatal("Get(UL element) failed")
	}
	if v, ok := ul.Uint32(); !ok || v != 0xDEADBEEF {
		t.Errorf("UL value = %#x/%v", v, ok)
	}
	if _, ok := (Element{Tag: 1, VR: "US"}).Uint16(); ok {
		t.Error("short US element reported ok")
	}
	if _, ok := (Element{Tag: 1, VR: "UL"}).Uint32(); ok {
		t.Error("short UL element reported ok")
	}

	var target DataSet
	target.Set(NewStringElement(TagModality, "CS", "CT"))
	target.Set(NewStringElement(TagModality, "CS", "MR"))
	if len(target) != 1 || target[0].String() != "MR" {
		t.Errorf("DataSet.Set did not replace: %+v", target)
	}
	if e := (Element{Tag: TagPatientName}); e.Group() != 0x0010 || e.Element() != 0x0010 {
		t.Errorf("Group/Element = %04X/%04X", e.Group(), e.Element())
	}
	if vr := VRForTag(0x99999999); vr != "UN" {
		t.Errorf("VRForTag(unknown) = %q", vr)
	}
	if !validVR("PN") || validVR("ZZ") {
		t.Error("validVR misclassified")
	}
	if !isLongVR("SQ") || isLongVR("PN") {
		t.Error("isLongVR misclassified")
	}
	if !strings.EqualFold(ImplicitVRLittleEndian, "1.2.840.10008.1.2") {
		t.Error("transfer syntax constant mismatch")
	}
}

// TestDecodeDataSetRejectsUnboundedSequenceNesting guards a remote
// denial-of-service: decoding is recursive per sequence level, and each level
// costs only ~32 bytes of input, so a C-FIND identifier of nested
// undefined-length sequences drives the decoder past the goroutine stack limit.
// The resulting "fatal error: stack overflow" is thrown by the runtime, NOT a
// panic, so gin.Recovery() cannot catch it and the whole server process dies.
// Measured before the cap: 400,000 levels (12.8 MB) killed the process with
// "goroutine stack exceeds 1000000000-byte limit". The accumulator permits a
// 16 MiB identifier, so this was reachable from the network.
func TestDecodeDataSetRejectsUnboundedSequenceNesting(t *testing.T) {
	const levels = 400000

	// (0008,1111) undefined-length SQ + undefined-length item, per level.
	buf := make([]byte, 0, levels*32)
	for i := 0; i < levels; i++ {
		buf = append(buf, 0x08, 0x00, 0x11, 0x11, 0xFF, 0xFF, 0xFF, 0xFF)
		buf = append(buf, 0xFE, 0xFF, 0x00, 0xE0, 0xFF, 0xFF, 0xFF, 0xFF)
	}
	for i := 0; i < levels; i++ {
		buf = append(buf, 0xFE, 0xFF, 0x0D, 0xE0, 0x00, 0x00, 0x00, 0x00)
		buf = append(buf, 0xFE, 0xFF, 0xDD, 0xE0, 0x00, 0x00, 0x00, 0x00)
	}

	_, err := DecodeDataSet(buf, false)
	if err == nil {
		t.Fatal("expected the nesting cap to reject this input, got nil error")
	}
	if !errors.Is(err, ErrElementTooLarge) {
		t.Fatalf("expected ErrElementTooLarge, got %v", err)
	}
}

// TestMaxSequenceDepthIsBounded pins the cap itself so it cannot be raised back
// to an unsafe value unnoticed.
func TestMaxSequenceDepthIsBounded(t *testing.T) {
	if MaxSequenceDepth <= 0 || MaxSequenceDepth > 256 {
		t.Fatalf("MaxSequenceDepth = %d; must be a small positive bound", MaxSequenceDepth)
	}
}
