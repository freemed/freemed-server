// Package dicom implements a minimal DICOM Part-10 reader sufficient to
// extract patient/study/series/SOP identifiers, modality and dates from a
// DICOM object without any external DICOM library.
//
// Supported transfer syntaxes:
//   - Explicit VR Little Endian (1.2.840.10008.1.2.1) — the common default
//   - Implicit VR Little Endian (1.2.840.10008.1.2)
//
// Only top-level (non-sequence) elements are decoded; sequence (SQ) values are
// skipped correctly, including undefined-length sequences encoded with the
// Item / Sequence Delimitation Item structure.
package dicom

import (
	"encoding/binary"
	"errors"
	"strings"
)

const (
	// TransferSyntaxUID values we understand.
	ImplicitVRLittleEndian = "1.2.840.10008.1.2"
	ExplicitVRLittleEndian = "1.2.840.10008.1.2.1"
)

// Metadata holds the DICOM attributes extracted from a Part-10 object.
type Metadata struct {
	PatientName            string
	PatientID              string
	PatientBirthDate       string
	StudyInstanceUID       string
	SeriesInstanceUID      string
	SOPInstanceUID         string
	Modality               string
	StudyDate              string
	StudyDescription       string
	InstitutionName        string
	ReferringPhysicianName string
	TransferSyntaxUID      string
}

// Parse reads a DICOM Part-10 object and returns the extracted metadata.
// It is deliberately lenient: unknown tags are skipped, and a best-effort
// result is returned even when the stream is truncated or an unsupported
// transfer syntax is encountered.
func Parse(data []byte) (Metadata, error) {
	var md Metadata
	if len(data) < 8 {
		return md, errors.New("dicom: data too short to be a DICOM object")
	}

	// Skip the 128-byte preamble and verify the "DICM" magic, when present.
	offset := 0
	if len(data) >= 132 && string(data[128:132]) == "DICM" {
		offset = 132
	}

	// The file meta information group (0002) is always Explicit VR Little
	// Endian. The remainder of the data set uses the transfer syntax declared
	// in (0002,0010). We default to Explicit VR LE, the most common case.
	explicit := true

	for {
		if offset+8 > len(data) {
			break
		}

		group := binary.LittleEndian.Uint16(data[offset : offset+2])
		element := binary.LittleEndian.Uint16(data[offset+2 : offset+4])

		// File-meta group is always explicit VR LE.
		isFileMeta := group == 0x0002
		useExplicit := isFileMeta || explicit

		var vr string
		var valueLen uint32
		headerLen := 0

		if useExplicit {
			if offset+8 > len(data) {
				break
			}
			vr = string(data[offset+4 : offset+6])
			if isLongVR(vr) {
				if offset+12 > len(data) {
					break
				}
				valueLen = binary.LittleEndian.Uint32(data[offset+8 : offset+12])
				headerLen = 12
			} else {
				valueLen = uint32(binary.LittleEndian.Uint16(data[offset+6 : offset+8]))
				headerLen = 8
			}
		} else {
			// Implicit VR Little Endian: no VR field, 4-byte length.
			valueLen = binary.LittleEndian.Uint32(data[offset+4 : offset+8])
			headerLen = 8
		}

		valueStart := offset + headerLen
		tag := (uint32(group) << 16) | uint32(element)

		if valueLen == 0xFFFFFFFF {
			// Undefined length: only valid for sequences (SQ) in our scope.
			offset = skipUndefinedSequence(data, valueStart)
		} else if useExplicit && vr == "SQ" {
			// Defined-length sequence: skip the whole value.
			offset = valueStart + int(valueLen)
		} else {
			if valueStart+int(valueLen) > len(data) {
				break
			}
			if valueLen > 0 {
				assignMetadata(&md, tag, data[valueStart:valueStart+int(valueLen)])
			}
			offset = valueStart + int(valueLen)
		}

		// Once we read the Transfer Syntax UID, switch the encoding used for
		// the remainder of the data set.
		if tag == 0x00020010 {
			if md.TransferSyntaxUID == ImplicitVRLittleEndian {
				explicit = false
			} else {
				explicit = true
			}
		}

		if offset >= len(data) {
			break
		}
	}

	return md, nil
}

// isLongVR reports whether vr uses the 16-bit reserved / 32-bit length layout
// instead of the 16-bit length layout (PS3.5 §7.1.2).
func isLongVR(vr string) bool {
	switch vr {
	case "OB", "OD", "OF", "OL", "OV", "OW", "SQ", "UC", "UR", "UT", "UN":
		return true
	}
	return false
}

// maxSequenceDepth bounds nesting of undefined-length sequence items while
// skipping. Each nesting level costs only a few bytes of input, so without a
// ceiling a crafted object (~47 bytes per level) drives this recursion past the
// goroutine stack limit, raising an UNCATCHABLE "fatal error: stack overflow"
// that kills the whole process — a gin.Recovery() cannot catch it. This is a
// security bound, not a formatting choice.
const maxSequenceDepth = 64

// skipUndefinedSequence advances past an undefined-length sequence value that
// is encoded as a series of Item (FFFE,E000) tags terminated by a Sequence
// Delimitation Item (FFFE,E0DD).
func skipUndefinedSequence(data []byte, pos int) int {
	return skipUndefinedSequenceDepth(data, pos, 0)
}

// skipUndefinedSequenceDepth is skipUndefinedSequence with an explicit nesting
// depth so that hostile input cannot exhaust the stack.
func skipUndefinedSequenceDepth(data []byte, pos, depth int) int {
	if depth >= maxSequenceDepth {
		// Nested implausibly deeply for a real object; give up on this sequence
		// the same way the malformed-tag path below does.
		return len(data)
	}
	for {
		if pos+8 > len(data) {
			return len(data)
		}
		group := binary.LittleEndian.Uint16(data[pos : pos+2])
		element := binary.LittleEndian.Uint16(data[pos+2 : pos+4])
		length := binary.LittleEndian.Uint32(data[pos+4 : pos+8])
		pos += 8

		if group != 0xFFFE {
			// Not an item/delimiter tag — malformed, bail out.
			return pos
		}
		switch element {
		case 0xE000: // Item
			if length == 0xFFFFFFFF {
				pos = skipUndefinedSequenceDepth(data, pos, depth+1) // undefined item length: recurse
			} else {
				pos += int(length)
			}
		case 0xE00D: // Item Delimitation Item
			return pos
		case 0xE0DD: // Sequence Delimitation Item
			return pos
		default:
			return pos
		}
	}
}

// assignMetadata maps a decoded element tag to the corresponding Metadata field.
func assignMetadata(md *Metadata, tag uint32, raw []byte) {
	s := strings.TrimRight(string(raw), " \x00")
	switch tag {
	case 0x00020010:
		md.TransferSyntaxUID = s
	case 0x00100010:
		md.PatientName = s
	case 0x00100020:
		md.PatientID = s
	case 0x00100030:
		md.PatientBirthDate = s
	case 0x0020000D:
		md.StudyInstanceUID = s
	case 0x0020000E:
		md.SeriesInstanceUID = s
	case 0x00080018:
		md.SOPInstanceUID = s
	case 0x00080060:
		md.Modality = s
	case 0x00080020:
		md.StudyDate = s
	case 0x00081030:
		md.StudyDescription = s
	case 0x00080080:
		md.InstitutionName = s
	case 0x00080090:
		md.ReferringPhysicianName = s
	}
}
