// Package hl7 implements HL7v2 message parsing, building, and MLLP transport.
//
// HL7v2 is a pipe-delimited healthcare messaging format. This package supports
// ADT (patient admit/discharge/transfer) and ORU (lab results) message types.
package hl7

import (
	"bytes"
	"fmt"
	"strings"
)

// HL7v2 delimiter constants.
const (
	SegmentSep    = '\r' // segment separator (carriage return)
	FieldSep      = '|'  // field separator (pipe)
	ComponentSep  = '^'  // component separator (caret)
	SubCompSep    = '&'  // sub-component separator (ampersand)
	RepetitionSep = '~'  // repetition separator (tilde)
	EscapeChar    = '\\' // escape character (backslash)
)

// Message represents a parsed HL7v2 message comprising ordered segments.
type Message struct {
	Segments []Segment
}

// Segment is one logical line of an HL7v2 message (e.g. MSH, PID, PV1).
// Fields is a slice of fields where each field is a slice of components.
// Fields[0] = first field, Fields[0][0] = first component of first field.
type Segment struct {
	Name   string       // segment identifier, e.g. "MSH", "PID"
	Fields [][]string   // Fields[f] = field f (0-based), Fields[f][c] = component c (0-based)
}

// GetField returns the full value of a field by its 1-based HL7 index.
// Components are joined with the component separator (^).
// Returns empty string if the field doesn't exist.
func (s *Segment) GetField(idx int) string {
	if idx < 1 || idx > len(s.Fields) {
		return ""
	}
	return strings.Join(s.Fields[idx-1], string(ComponentSep))
}

// GetComponent returns a component value from a field.
// Both indexes are 1-based (HL7 convention).
func (s *Segment) GetComponent(fieldIdx, compIdx int) string {
	if fieldIdx < 1 || fieldIdx > len(s.Fields) {
		return ""
	}
	components := s.Fields[fieldIdx-1]
	if compIdx < 1 || compIdx > len(components) {
		return ""
	}
	return components[compIdx-1]
}

// splitField splits a single HL7 field value into its components.
func splitField(field string) []string {
	if field == "" {
		return []string{""}
	}
	return strings.Split(field, string(ComponentSep))
}

// Unmarshal parses raw HL7v2 wire-format bytes into a Message.
// The input should already have MLLP framing (<VT> and <FS><CR>) stripped.
func Unmarshal(data []byte) (*Message, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("hl7: empty message")
	}

	msg := &Message{}
	// Split on segment separator (\r)
	// Also handle \n variants that real-world senders may use.
	raw := string(data)
	// Normalize: replace \r\n with \r, then split on \r
	raw = strings.ReplaceAll(raw, "\r\n", "\r")
	lines := strings.Split(raw, string(SegmentSep))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		seg, err := parseSegment(line)
		if err != nil {
			return nil, err
		}
		msg.Segments = append(msg.Segments, seg)
	}

	if len(msg.Segments) == 0 {
		return nil, fmt.Errorf("hl7: no segments found")
	}

	return msg, nil
}

// parseSegment parses a single segment line into a Segment.
// Fields[0] corresponds to HL7 field index 1.
// For MSH segments, Fields[0] is always "|" (the field separator itself, MSH-1).
func parseSegment(line string) (Segment, error) {
	parts := strings.Split(line, string(FieldSep))
	if len(parts) == 0 {
		return Segment{}, fmt.Errorf("hl7: empty segment line")
	}

	name := strings.TrimSpace(parts[0])
	seg := Segment{
		Name:   name,
		Fields: make([][]string, 0, len(parts)),
	}

	if strings.EqualFold(name, "MSH") {
		// MSH-1 is the field separator itself — always "|"
		seg.Fields = append(seg.Fields, []string{"|"})
	}

	// Add remaining fields (skip parts[0] which is the segment name)
	for i := 1; i < len(parts); i++ {
		seg.Fields = append(seg.Fields, splitField(parts[i]))
	}

	return seg, nil
}

// Marshal serializes a Message to HL7v2 wire format.
// Uses \r as segment separator.
func (m *Message) Marshal() []byte {
	var buf bytes.Buffer
	for i, seg := range m.Segments {
		if i > 0 {
			buf.WriteByte(SegmentSep)
		}
		buf.WriteString(seg.Name)

		isMSH := strings.EqualFold(seg.Name, "MSH")
		for j, field := range seg.Fields {
			// MSH-1 is the field separator itself — skip writing its value
			// but still write the pipe delimiter (handled at start of loop)
			if isMSH && j == 0 {
				continue
			}
			buf.WriteByte(FieldSep)
			buf.WriteString(strings.Join(field, string(ComponentSep)))
		}
	}
	return buf.Bytes()
}

// FindSegment returns the first segment with the given name, or nil if not found.
func (m *Message) FindSegment(name string) *Segment {
	for i := range m.Segments {
		if strings.EqualFold(m.Segments[i].Name, name) {
			return &m.Segments[i]
		}
	}
	return nil
}

// FindAllSegments returns all segments with the given name.
func (m *Message) FindAllSegments(name string) []*Segment {
	var result []*Segment
	for i := range m.Segments {
		if strings.EqualFold(m.Segments[i].Name, name) {
			result = append(result, &m.Segments[i])
		}
	}
	return result
}
