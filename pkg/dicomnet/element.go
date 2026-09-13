package dicomnet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

// Transfer syntaxes implemented by the codec.
const (
	ImplicitVRLittleEndian = "1.2.840.10008.1.2"
	ExplicitVRLittleEndian = "1.2.840.10008.1.2.1"
)

// MaxElementLength is the absolute ceiling on a single element value or
// sequence item. Anything larger is rejected before allocating.
const MaxElementLength = 16 << 20 // 16 MiB

// MaxSequenceDepth bounds sequence nesting while decoding. Decoding recurses per
// level, so without a ceiling a data set of nested undefined-length sequence
// items (~32 bytes per level) drives the decoder past the goroutine stack limit
// and raises an UNCATCHABLE "fatal error: stack overflow" that kills the whole
// process. A hostile SCU can reach ~400k levels well inside the 16 MiB message
// cap, so this bound is a security control, not a formatting nicety.
const MaxSequenceDepth = 64

// UndefinedLength is the 0xFFFFFFFF length marker used by undefined-length
// sequences and items.
const UndefinedLength uint32 = 0xFFFFFFFF

// Item / delimiter tags (group 0xFFFE).
const (
	TagItem              uint32 = 0xFFFEE000
	TagItemDelimitation  uint32 = 0xFFFEE00D
	TagSequenceDelimiter uint32 = 0xFFFEE0DD
)

// Errors returned by the element codec. All of them indicate hostile or
// corrupt input; none of the decoder entry points panic.
var (
	ErrTruncatedElement  = errors.New("dicomnet: truncated element")
	ErrElementTooLarge   = errors.New("dicomnet: element value exceeds maximum length")
	ErrInvalidVR         = errors.New("dicomnet: unknown value representation")
	ErrUnexpectedItemTag = errors.New("dicomnet: item tag outside a sequence")
	ErrUnterminatedSeq   = errors.New("dicomnet: unterminated undefined-length sequence")
	ErrInvalidLength     = errors.New("dicomnet: invalid element length")
	ErrInvalidTag        = errors.New("dicomnet: invalid tag")
)

// Element is a single DICOM data element. For primitive VRs the value is in
// Value; for sequences the items are in Items (each item being a DataSet).
type Element struct {
	Tag   uint32
	VR    string
	Value []byte
	Items []DataSet

	// UndefinedLength is true when the element (a sequence) was encoded with
	// the 0xFFFFFFFF length marker.
	UndefinedLength bool
}

// DataSet is an ordered list of data elements. Order is preserved so that a
// decoded identifier can be re-encoded faithfully.
type DataSet []Element

// NewStringElement builds a primitive element from a string value.
func NewStringElement(tag uint32, vr, value string) Element {
	return Element{Tag: tag, VR: vr, Value: []byte(value)}
}

// NewSequenceElement builds a sequence element with the given items.
func NewSequenceElement(tag uint32, items ...DataSet) Element {
	return Element{Tag: tag, VR: "SQ", Items: items}
}

// NewUSElement builds a US (unsigned short) element.
func NewUSElement(tag uint32, v uint16) Element {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, v)
	return Element{Tag: tag, VR: "US", Value: b}
}

// NewULElement builds a UL (unsigned long) element.
func NewULElement(tag uint32, v uint32) Element {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return Element{Tag: tag, VR: "UL", Value: b}
}

// Group returns the element's group number.
func (e Element) Group() uint16 { return uint16(e.Tag >> 16) }

// Element returns the element's element number.
func (e Element) Element() uint16 { return uint16(e.Tag & 0xFFFF) }

// String returns the element value with DICOM trailing padding removed.
// Multi-valued VRs keep their backslash separators. Non-text VRs return the
// raw bytes as a string.
func (e Element) String() string {
	return strings.TrimRight(string(e.Value), " \x00")
}

// Values returns the multi-valued interpretation of a text element (the value
// split on backslash). A single-valued element returns a one element slice.
func (e Element) Values() []string {
	return SplitValues(e.String())
}

// Uint16 returns the element's value interpreted as a little endian US.
func (e Element) Uint16() (uint16, bool) {
	if len(e.Value) < 2 {
		return 0, false
	}
	return binary.LittleEndian.Uint16(e.Value[:2]), true
}

// Uint32 returns the element's value interpreted as a little endian UL.
func (e Element) Uint32() (uint32, bool) {
	if len(e.Value) < 4 {
		return 0, false
	}
	return binary.LittleEndian.Uint32(e.Value[:4]), true
}

// SplitValues splits a multi-valued DICOM value on backslash.
func SplitValues(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\\")
}

// JoinValues joins values into a multi-valued DICOM value.
func JoinValues(v []string) string { return strings.Join(v, "\\") }

// Get returns the element with the given tag.
func (ds DataSet) Get(tag uint32) (Element, bool) {
	for _, e := range ds {
		if e.Tag == tag {
			return e, true
		}
	}
	return Element{}, false
}

// GetString returns the string value of the first element with tag, or "".
func (ds DataSet) GetString(tag uint32) string {
	if e, ok := ds.Get(tag); ok {
		return e.String()
	}
	return ""
}

// GetSequence returns the first item of the sequence element with the given
// tag. It reports false when the tag is absent, is not a sequence, or has no
// items.
func (ds DataSet) GetSequence(tag uint32) (DataSet, bool) {
	e, ok := ds.Get(tag)
	if !ok || len(e.Items) == 0 {
		return nil, false
	}
	return e.Items[0], true
}

// Set replaces (or appends) an element by tag.
func (ds *DataSet) Set(e Element) {
	for i := range *ds {
		if (*ds)[i].Tag == e.Tag {
			(*ds)[i] = e
			return
		}
	}
	*ds = append(*ds, e)
}

// isLongVR reports whether a VR uses the 16-bit reserved field plus 32-bit
// length layout instead of the 16-bit length layout (PS3.5 §7.1.2).
func isLongVR(vr string) bool {
	switch vr {
	case "OB", "OD", "OF", "OL", "OV", "OW", "SQ", "UC", "UR", "UT", "UN":
		return true
	}
	return false
}

// validVR reports whether vr is a known value representation. Unknown VRs are
// treated as corrupt input rather than silently accepted.
func validVR(vr string) bool {
	switch vr {
	case "AE", "AS", "AT", "CS", "DA", "DS", "DT", "FL", "FD", "IS", "LO",
		"LT", "OB", "OD", "OF", "OL", "OV", "OW", "PN", "SH", "SL", "SQ",
		"SS", "ST", "SV", "TM", "UC", "UI", "UL", "UN", "UR", "US", "UT",
		"UV":
		return true
	}
	return false
}

// vrDictionary maps tags to VRs for Implicit VR Little Endian decoding, where
// the VR is not present on the wire. Tags outside the dictionary are treated
// as UN (raw bytes), which is safe: the value round-trips through Value.
var vrDictionary = map[uint32]string{
	TagSpecificCharacterSet:          "CS",
	TagSOPClassUID:                   "UI",
	TagSOPInstanceUID:                "UI",
	TagStudyDate:                     "DA",
	TagSeriesDate:                    "DA",
	TagModality:                      "CS",
	TagStudyDescription:              "LO",
	TagInstitutionName:               "LO",
	TagReferringPhysicianName:        "PN",
	TagAccessionNumber:               "SH",
	TagPatientName:                   "PN",
	TagPatientID:                     "LO",
	TagPatientBirthDate:              "DA",
	TagPatientSex:                    "CS",
	TagStudyInstanceUID:              "UI",
	TagRequestedProcedureDescription: "LO",
	TagRequestedContrastAgent:        "LO",
	TagScheduledProcedureStepSeq:     "SQ",
	TagScheduledStationAETitle:       "AE",
	TagSPSStartDate:                  "DA",
	TagSPSStartTime:                  "TM",
	TagSPSEndDate:                    "DA",
	TagSPSEndTime:                    "TM",
	TagScheduledPerformingPhysName:   "PN",
	TagSPSDescription:                "LO",
	TagSPSID:                         "SH",
	TagSPSLocation:                   "SH",
	TagSPSStatus:                     "CS",
	TagRequestedProcedureID:          "SH",
	TagRequestedProcedurePriority:    "SH",
}

// getTag reads a data element tag. DICOM writes the group number first, then
// the element number, each as a little endian uint16 (PS3.5 §7.1.1), so the
// tag value is rebuilt as (group << 16 | element).
func getTag(b []byte) uint32 {
	return uint32(binary.LittleEndian.Uint16(b[0:2]))<<16 | uint32(binary.LittleEndian.Uint16(b[2:4]))
}

// putTag writes a data element tag in wire order (group first).
func putTag(b []byte, tag uint32) {
	binary.LittleEndian.PutUint16(b[0:2], uint16(tag>>16))
	binary.LittleEndian.PutUint16(b[2:4], uint16(tag))
}

// VRForTag returns the VR used when decoding the tag from an implicit VR
// stream. Unknown tags decode as UN.
func VRForTag(tag uint32) string {
	if vr, ok := vrDictionary[tag]; ok {
		return vr
	}
	return "UN"
}

// EncodeDataSet encodes a data set using the requested transfer syntax.
func EncodeDataSet(ds DataSet, explicit bool) ([]byte, error) {
	var out []byte
	for _, e := range ds {
		b, err := EncodeElement(e, explicit)
		if err != nil {
			return nil, err
		}
		out = append(out, b...)
	}
	return out, nil
}

// EncodeElement encodes a single element. explicit selects Explicit VR Little
// Endian encoding; otherwise Implicit VR Little Endian is used.
func EncodeElement(e Element, explicit bool) ([]byte, error) {
	vr := e.VR
	if vr == "" {
		vr = "UN"
	}
	if e.Group() == 0xFFFE {
		// Item / delimiter tags never carry a VR, in either syntax.
		switch e.Tag {
		case TagItem, TagItemDelimitation, TagSequenceDelimiter:
		default:
			return nil, fmt.Errorf("%w: 0x%08X", ErrInvalidTag, e.Tag)
		}
		length := uint32(len(e.Value))
		if len(e.Items) > 0 {
			length = 0
		}
		if e.UndefinedLength {
			length = UndefinedLength
		}
		out := make([]byte, 8)
		putTag(out[0:4], e.Tag)
		binary.LittleEndian.PutUint32(out[4:8], length)
		return append(out, e.Value...), nil
	}

	if vr == "SQ" {
		return encodeSequence(e, explicit)
	}

	value := e.Value
	if uint32(len(value)) > MaxElementLength {
		return nil, ErrElementTooLarge
	}

	out := make([]byte, 0, 8+len(value))
	var tagBuf [4]byte
	putTag(tagBuf[:], e.Tag)
	out = append(out, tagBuf[:]...)

	if explicit {
		if !validVR(vr) {
			return nil, fmt.Errorf("%w: %q", ErrInvalidVR, vr)
		}
		out = append(out, vr...)
		if isLongVR(vr) {
			out = append(out, 0, 0)
			var l [4]byte
			binary.LittleEndian.PutUint32(l[:], uint32(len(value)))
			out = append(out, l[:]...)
		} else {
			if len(value) > 0xFFFF {
				return nil, ErrElementTooLarge
			}
			var l [2]byte
			binary.LittleEndian.PutUint16(l[:], uint16(len(value)))
			out = append(out, l[:]...)
		}
	} else {
		var l [4]byte
		binary.LittleEndian.PutUint32(l[:], uint32(len(value)))
		out = append(out, l[:]...)
	}
	return append(out, value...), nil
}

// encodeSequence encodes an SQ element, choosing defined or undefined length.
func encodeSequence(e Element, explicit bool) ([]byte, error) {
	var items []byte
	for _, it := range e.Items {
		body, err := EncodeDataSet(it, explicit)
		if err != nil {
			return nil, err
		}
		var hdr [8]byte
		putTag(hdr[0:4], TagItem)
		binary.LittleEndian.PutUint32(hdr[4:8], uint32(len(body)))
		items = append(items, hdr[:]...)
		items = append(items, body...)
	}

	delim := make([]byte, 8)
	putTag(delim[0:4], TagSequenceDelimiter)
	// Delimitation items always carry zero length.

	out := make([]byte, 0, 12+len(items))
	var tagBuf [4]byte
	putTag(tagBuf[:], e.Tag)
	out = append(out, tagBuf[:]...)

	if explicit {
		out = append(out, "SQ"...)
		out = append(out, 0, 0)
		var l [4]byte
		if e.UndefinedLength {
			binary.LittleEndian.PutUint32(l[:], UndefinedLength)
			out = append(out, l[:]...)
			out = append(out, items...)
			out = append(out, delim...)
			return out, nil
		}
		binary.LittleEndian.PutUint32(l[:], uint32(len(items)))
		out = append(out, l[:]...)
		return append(out, items...), nil
	}

	var l [4]byte
	if e.UndefinedLength {
		binary.LittleEndian.PutUint32(l[:], UndefinedLength)
		out = append(out, l[:]...)
		out = append(out, items...)
		out = append(out, delim...)
		return out, nil
	}
	binary.LittleEndian.PutUint32(l[:], uint32(len(items)))
	out = append(out, l[:]...)
	return append(out, items...), nil
}

// DecodeDataSet decodes every element in data. It returns an error (never a
// panic) on truncated, oversized or otherwise malformed input.
func DecodeDataSet(data []byte, explicit bool) (DataSet, error) {
	ds, next, err := decodeDataSetRange(data, explicit, 0)
	if err != nil {
		return nil, err
	}
	if next != len(data) {
		return nil, fmt.Errorf("%w: %d trailing bytes", ErrTruncatedElement, len(data)-next)
	}
	return ds, nil
}

// decodeDataSetRange decodes elements until the end of data, returning the
// offset consumed.
func decodeDataSetRange(data []byte, explicit bool, depth int) (DataSet, int, error) {
	var ds DataSet
	pos := 0
	for pos < len(data) {
		e, next, err := decodeElementAt(data, pos, explicit, depth)
		if err != nil {
			return nil, 0, err
		}
		if e != nil {
			ds = append(ds, *e)
		}
		pos = next
	}
	return ds, pos, nil
}

// decodeElementAt decodes a single element at pos, returning the offset of the
// next element. Item and delimitation tags are rejected here: they are only
// meaningful inside a sequence, where decodeSequenceItems consumes them.
func decodeElementAt(data []byte, pos int, explicit bool, depth int) (*Element, int, error) {
	if pos < 0 || pos+8 > len(data) {
		return nil, 0, fmt.Errorf("%w: element header at %d", ErrTruncatedElement, pos)
	}
	tag := getTag(data[pos : pos+4])

	if uint16(tag>>16) == 0xFFFE {
		switch tag {
		case TagItem:
			return nil, 0, fmt.Errorf("%w: 0x%08X at %d", ErrUnexpectedItemTag, tag, pos)
		case TagItemDelimitation, TagSequenceDelimiter:
			return nil, 0, fmt.Errorf("%w: delimiter 0x%08X at %d", ErrUnexpectedItemTag, tag, pos)
		default:
			return nil, 0, fmt.Errorf("%w: 0x%08X at %d", ErrInvalidTag, tag, pos)
		}
	}

	var vr string
	var length uint32
	headerLen := 0
	if explicit {
		if pos+8 > len(data) {
			return nil, 0, fmt.Errorf("%w: explicit element header at %d", ErrTruncatedElement, pos)
		}
		vr = string(data[pos+4 : pos+6])
		if !validVR(vr) {
			return nil, 0, fmt.Errorf("%w: %q at %d", ErrInvalidVR, vr, pos)
		}
		if isLongVR(vr) {
			if pos+12 > len(data) {
				return nil, 0, fmt.Errorf("%w: long VR header at %d", ErrTruncatedElement, pos)
			}
			length = binary.LittleEndian.Uint32(data[pos+8 : pos+12])
			headerLen = 12
		} else {
			length = uint32(binary.LittleEndian.Uint16(data[pos+6 : pos+8]))
			headerLen = 8
		}
	} else {
		vr = VRForTag(tag)
		length = binary.LittleEndian.Uint32(data[pos+4 : pos+8])
		headerLen = 8
	}

	valueStart := pos + headerLen

	if vr == "SQ" || length == UndefinedLength {
		if explicit && vr != "SQ" {
			return nil, 0, fmt.Errorf("%w: undefined length on non-SQ VR %q at %d", ErrInvalidLength, vr, pos)
		}
		if depth >= MaxSequenceDepth {
			return nil, 0, fmt.Errorf("%w: sequence nesting exceeds %d at %d", ErrElementTooLarge, MaxSequenceDepth, pos)
		}
		items, next, err := decodeSequenceItems(data, valueStart, length, explicit, depth+1)
		if err != nil {
			return nil, 0, err
		}
		return &Element{
			Tag:             tag,
			VR:              "SQ",
			Items:           items,
			UndefinedLength: length == UndefinedLength,
		}, next, nil
	}

	if length > MaxElementLength {
		return nil, 0, fmt.Errorf("%w: %d bytes at %d", ErrElementTooLarge, length, pos)
	}
	if uint64(valueStart)+uint64(length) > uint64(len(data)) {
		return nil, 0, fmt.Errorf("%w: value %d bytes at %d, only %d remain", ErrTruncatedElement, length, pos, len(data)-valueStart)
	}
	value := make([]byte, length)
	copy(value, data[valueStart:valueStart+int(length)])
	return &Element{Tag: tag, VR: vr, Value: value}, valueStart + int(length), nil
}

// decodeSequenceItems decodes the items of a sequence value. When the sequence
// length is undefined, decoding stops at the Sequence Delimitation Item. The
// transfer syntax of the enclosing data set applies inside the sequence.
func decodeSequenceItems(data []byte, pos int, length uint32, explicit bool, depth int) ([]DataSet, int, error) {
	undefined := length == UndefinedLength
	end := len(data)
	if !undefined {
		if length > MaxElementLength {
			return nil, 0, fmt.Errorf("%w: sequence length %d", ErrElementTooLarge, length)
		}
		if uint64(pos)+uint64(length) > uint64(len(data)) {
			return nil, 0, fmt.Errorf("%w: sequence value %d bytes at %d", ErrTruncatedElement, length, pos)
		}
		end = pos + int(length)
	}

	var items []DataSet
	for pos < end {
		if pos+8 > end {
			if undefined {
				return nil, 0, fmt.Errorf("%w: sequence item header at %d", ErrUnterminatedSeq, pos)
			}
			return nil, 0, fmt.Errorf("%w: sequence item header at %d", ErrTruncatedElement, pos)
		}
		tag := getTag(data[pos : pos+4])
		itemLen := binary.LittleEndian.Uint32(data[pos+4 : pos+8])
		bodyStart := pos + 8

		switch tag {
		case TagSequenceDelimiter:
			return items, bodyStart, nil
		case TagItemDelimitation:
			// Stray item delimiter: treat as the end of the enclosing item.
			return items, bodyStart, nil
		case TagItem:
			if itemLen == UndefinedLength {
				item, next, err := decodeUndefinedItem(data, bodyStart, explicit, depth)
				if err != nil {
					return nil, 0, err
				}
				items = append(items, item)
				pos = next
				continue
			}
			if itemLen > MaxElementLength {
				return nil, 0, fmt.Errorf("%w: item length %d at %d", ErrElementTooLarge, itemLen, pos)
			}
			if uint64(bodyStart)+uint64(itemLen) > uint64(end) {
				return nil, 0, fmt.Errorf("%w: item value %d bytes at %d", ErrTruncatedElement, itemLen, pos)
			}
			item, next, err := decodeDataSetRange(data[bodyStart:bodyStart+int(itemLen)], explicit, depth+1)
			if err != nil {
				return nil, 0, err
			}
			if next != int(itemLen) {
				return nil, 0, fmt.Errorf("%w: item at %d has %d trailing bytes", ErrTruncatedElement, pos, int(itemLen)-next)
			}
			items = append(items, item)
			pos = bodyStart + int(itemLen)
		default:
			return nil, 0, fmt.Errorf("%w: 0x%08X inside sequence at %d", ErrUnexpectedItemTag, tag, pos)
		}
	}
	if undefined {
		return nil, 0, fmt.Errorf("%w: reached end of data at %d", ErrUnterminatedSeq, pos)
	}
	return items, pos, nil
}

// decodeUndefinedItem decodes an item with undefined length, terminated by an
// Item Delimitation Item.
func decodeUndefinedItem(data []byte, pos int, explicit bool, depth int) (DataSet, int, error) {
	var ds DataSet
	for pos < len(data) {
		if pos+8 > len(data) {
			return nil, 0, fmt.Errorf("%w: undefined item at %d", ErrTruncatedElement, pos)
		}
		tag := getTag(data[pos : pos+4])
		if tag == TagItemDelimitation {
			return ds, pos + 8, nil
		}
		if uint16(tag>>16) == 0xFFFE {
			return nil, 0, fmt.Errorf("%w: 0x%08X in undefined item at %d", ErrUnexpectedItemTag, tag, pos)
		}
		e, next, err := decodeElementAt(data, pos, explicit, depth)
		if err != nil {
			return nil, 0, err
		}
		if e != nil {
			ds = append(ds, *e)
		}
		pos = next
	}
	return nil, 0, fmt.Errorf("%w: undefined item at %d", ErrUnterminatedSeq, pos)
}
