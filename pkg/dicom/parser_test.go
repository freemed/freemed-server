package dicom

import (
	"encoding/binary"
	"testing"
)

// appendExplicitElement appends an Explicit VR Little Endian element.
func appendExplicitElement(buf []byte, group, elem uint16, vr string, value []byte) []byte {
	h := make([]byte, 4)
	binary.LittleEndian.PutUint16(h[0:2], group)
	binary.LittleEndian.PutUint16(h[2:4], elem)
	buf = append(buf, h...)
	buf = append(buf, vr...)
	if isLongVR(vr) {
		buf = append(buf, 0, 0) // reserved bytes
		l := make([]byte, 4)
		binary.LittleEndian.PutUint32(l, uint32(len(value)))
		buf = append(buf, l...)
	} else {
		l := make([]byte, 2)
		binary.LittleEndian.PutUint16(l, uint16(len(value)))
		buf = append(buf, l...)
	}
	return append(buf, value...)
}

// appendImplicitElement appends an Implicit VR Little Endian element (no VR
// field, 4-byte length).
func appendImplicitElement(buf []byte, group, elem uint16, value []byte) []byte {
	h := make([]byte, 4)
	binary.LittleEndian.PutUint16(h[0:2], group)
	binary.LittleEndian.PutUint16(h[2:4], elem)
	buf = append(buf, h...)
	l := make([]byte, 4)
	binary.LittleEndian.PutUint32(l, uint32(len(value)))
	buf = append(buf, l...)
	return append(buf, value...)
}

// item wraps raw bytes in a defined-length DICOM Item (FFFE,E000).
func item(content []byte) []byte {
	out := make([]byte, 8)
	binary.LittleEndian.PutUint16(out[0:2], 0xFFFE)
	binary.LittleEndian.PutUint16(out[2:4], 0xE000)
	binary.LittleEndian.PutUint32(out[4:8], uint32(len(content)))
	return append(out, content...)
}

// undefinedSequence builds an SQ element with undefined length whose value is
// the given item content, terminated by a Sequence Delimitation Item.
func undefinedSequence(group, elem uint16, itemContent []byte) []byte {
	out := make([]byte, 4)
	binary.LittleEndian.PutUint16(out[0:2], group)
	binary.LittleEndian.PutUint16(out[2:4], elem)
	out = append(out, "SQ"...)
	out = append(out, 0, 0) // reserved bytes
	l := make([]byte, 4)
	binary.LittleEndian.PutUint32(l, 0xFFFFFFFF)
	out = append(out, l...)
	out = append(out, item(itemContent)...)
	delim := make([]byte, 8)
	binary.LittleEndian.PutUint16(delim[0:2], 0xFFFE)
	binary.LittleEndian.PutUint16(delim[2:4], 0xE0DD)
	out = append(out, delim...)
	return out
}

// wrapPart10 wraps a raw data set in a DICOM Part-10 preamble + "DICM" magic.
func wrapPart10(dataset []byte) []byte {
	out := make([]byte, 132)
	copy(out[128:132], "DICM")
	return append(out, dataset...)
}

func buildExplicitDataset() []byte {
	var b []byte
	b = appendExplicitElement(b, 0x0010, 0x0010, "PN", []byte("Doe^Jane"))
	b = appendExplicitElement(b, 0x0010, 0x0020, "LO", []byte("12345"))
	b = appendExplicitElement(b, 0x0010, 0x0030, "DA", []byte("19800101"))
	b = appendExplicitElement(b, 0x0020, 0x000D, "UI", []byte("1.2.3.4.5.6.7.8.9.100"))
	b = appendExplicitElement(b, 0x0020, 0x000E, "UI", []byte("1.2.3.4.5.6.7.8.9.100.1"))
	b = appendExplicitElement(b, 0x0008, 0x0018, "UI", []byte("1.2.3.4.5.6.7.8.9.100.1.1"))
	b = appendExplicitElement(b, 0x0008, 0x0060, "CS", []byte("CT"))
	b = appendExplicitElement(b, 0x0008, 0x0020, "DA", []byte("20240101"))
	b = appendExplicitElement(b, 0x0008, 0x1030, "LO", []byte("Chest CT"))
	b = appendExplicitElement(b, 0x0008, 0x0080, "LO", []byte("Test Hospital"))
	b = appendExplicitElement(b, 0x0008, 0x0090, "PN", []byte("Referring^Doc"))
	return b
}

func buildFileMeta(transferSyntax string) []byte {
	var b []byte
	b = appendExplicitElement(b, 0x0002, 0x0001, "OB", []byte{0x00, 0x01})
	b = appendExplicitElement(b, 0x0002, 0x0010, "UI", []byte(transferSyntax))
	return b
}

func TestParseExplicitVRLE(t *testing.T) {
	data := wrapPart10(append(buildFileMeta(ExplicitVRLittleEndian), buildExplicitDataset()...))

	md, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []struct {
		name, got, want string
	}{
		{"PatientName", md.PatientName, "Doe^Jane"},
		{"PatientID", md.PatientID, "12345"},
		{"PatientBirthDate", md.PatientBirthDate, "19800101"},
		{"StudyInstanceUID", md.StudyInstanceUID, "1.2.3.4.5.6.7.8.9.100"},
		{"SeriesInstanceUID", md.SeriesInstanceUID, "1.2.3.4.5.6.7.8.9.100.1"},
		{"SOPInstanceUID", md.SOPInstanceUID, "1.2.3.4.5.6.7.8.9.100.1.1"},
		{"Modality", md.Modality, "CT"},
		{"StudyDate", md.StudyDate, "20240101"},
		{"StudyDescription", md.StudyDescription, "Chest CT"},
		{"InstitutionName", md.InstitutionName, "Test Hospital"},
		{"ReferringPhysicianName", md.ReferringPhysicianName, "Referring^Doc"},
		{"TransferSyntaxUID", md.TransferSyntaxUID, ExplicitVRLittleEndian},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, c.got, c.want)
		}
	}
}

func TestParseImplicitVRLE(t *testing.T) {
	var dataset []byte
	dataset = appendImplicitElement(dataset, 0x0010, 0x0010, []byte("Smith^John"))
	dataset = appendImplicitElement(dataset, 0x0010, 0x0020, []byte("99999"))
	dataset = appendImplicitElement(dataset, 0x0020, 0x000D, []byte("2.3.4.5.6.7.8.9.200"))
	dataset = appendImplicitElement(dataset, 0x0020, 0x000E, []byte("2.3.4.5.6.7.8.9.200.1"))
	dataset = appendImplicitElement(dataset, 0x0008, 0x0018, []byte("2.3.4.5.6.7.8.9.200.1.1"))
	dataset = appendImplicitElement(dataset, 0x0008, 0x0060, []byte("MR"))
	dataset = appendImplicitElement(dataset, 0x0008, 0x0020, []byte("20240315"))

	data := wrapPart10(append(buildFileMeta(ImplicitVRLittleEndian), dataset...))

	md, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if md.PatientName != "Smith^John" {
		t.Errorf("PatientName: got %q, want %q", md.PatientName, "Smith^John")
	}
	if md.PatientID != "99999" {
		t.Errorf("PatientID: got %q, want %q", md.PatientID, "99999")
	}
	if md.StudyInstanceUID != "2.3.4.5.6.7.8.9.200" {
		t.Errorf("StudyInstanceUID: got %q", md.StudyInstanceUID)
	}
	if md.Modality != "MR" {
		t.Errorf("Modality: got %q, want %q", md.Modality, "MR")
	}
	if md.StudyDate != "20240315" {
		t.Errorf("StudyDate: got %q, want %q", md.StudyDate, "20240315")
	}
	if md.SOPInstanceUID != "2.3.4.5.6.7.8.9.200.1.1" {
		t.Errorf("SOPInstanceUID: got %q", md.SOPInstanceUID)
	}
}

func TestParseSkipsSequences(t *testing.T) {
	// Insert a defined-length sequence and an undefined-length sequence
	// between StudyDate and StudyDescription; both must be skipped so the
	// trailing tags are still extracted correctly.
	var b []byte
	b = appendExplicitElement(b, 0x0008, 0x0020, "DA", []byte("20240101"))

	definedSeqValue := item(appendExplicitElement(nil, 0x0008, 0x1150, "UI", []byte("1.2.3.4.5")))
	b = appendExplicitElement(b, 0x0008, 0x1140, "SQ", definedSeqValue)

	b = append(b, undefinedSequence(0x0008, 0x1115, appendExplicitElement(nil, 0x0008, 0x1155, "UI", []byte("1.2.3.4.6")))...)

	b = appendExplicitElement(b, 0x0008, 0x1030, "LO", []byte("Chest CT"))
	b = appendExplicitElement(b, 0x0008, 0x0018, "UI", []byte("9.9.9.9.9"))

	data := wrapPart10(append(buildFileMeta(ExplicitVRLittleEndian), b...))

	md, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if md.StudyDescription != "Chest CT" {
		t.Errorf("StudyDescription: got %q, want %q", md.StudyDescription, "Chest CT")
	}
	if md.SOPInstanceUID != "9.9.9.9.9" {
		t.Errorf("SOPInstanceUID: got %q, want %q", md.SOPInstanceUID, "9.9.9.9.9")
	}
	if md.StudyDate != "20240101" {
		t.Errorf("StudyDate: got %q, want %q", md.StudyDate, "20240101")
	}
}

func TestParseTooShort(t *testing.T) {
	if _, err := Parse([]byte{0x01, 0x02, 0x03}); err == nil {
		t.Error("expected error for short input, got nil")
	}
}
