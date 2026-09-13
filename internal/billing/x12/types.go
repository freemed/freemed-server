package x12

import (
	"bytes"
	"strings"
)

// X12 delimiters (standard EDI separators).
const (
	SegmentTerminator = "~"
	ElementSeparator  = "*"
	SubElementSep     = ":"
	RepetitionSep     = "^"
)

// Segment represents a single X12 segment (e.g. NM1*85*2*CLINIC NAME~).
type Segment struct {
	Tag      string   // e.g. "NM1"
	Elements []string // e.g. ["85", "2", "CLINIC NAME", ...]
}

// sanitizeElement neutralizes X12 delimiters embedded in a data element value.
// Unescaped delimiters coming from patient/provider free text (names, addresses)
// would otherwise be interpreted as structural separators — segment or element
// injection that silently corrupts (or forges) the transaction. The component
// element separator SubElementSep is deliberately NOT stripped: it is legal
// structure inside composite elements such as SV1's "HC:99213:25".
func sanitizeElement(v string) string {
	if !strings.ContainsAny(v, ElementSeparator+SegmentTerminator+RepetitionSep) {
		return v
	}
	return strings.NewReplacer(
		ElementSeparator, " ",
		SegmentTerminator, " ",
		RepetitionSep, " ",
	).Replace(v)
}

// String serializes the segment using the standard element separator and terminator.
func (s Segment) String() string {
	// The ISA segment is fixed-width and ISA16 legitimately carries the
	// component element separator, so the envelope must never be sanitized.
	if s.Tag == "ISA" {
		return s.Tag + ElementSeparator + strings.Join(s.Elements, ElementSeparator) + SegmentTerminator
	}

	parts := make([]string, len(s.Elements))
	for i, e := range s.Elements {
		parts[i] = sanitizeElement(e)
	}
	return s.Tag + ElementSeparator + strings.Join(parts, ElementSeparator) + SegmentTerminator
}

// TransactionSetBuffer accumulates serialized X12 segments and counts them, so
// the closing SE segment can report a conformant segment count (SE01 is the
// number of segments from ST through SE inclusive).
type TransactionSetBuffer struct {
	bytes.Buffer
	counting bool
	segments int
}

// CountSegments begins counting segments from this point forward. Call it
// immediately after the GS segment is written so ST is included in the count.
func (b *TransactionSetBuffer) CountSegments() { b.counting = true }

// SegmentCount returns the number of segments written since CountSegments.
func (b *TransactionSetBuffer) SegmentCount() int { return b.segments }

// WriteString writes s to the underlying buffer, counting segment terminators
// while counting is enabled.
func (b *TransactionSetBuffer) WriteString(s string) (int, error) {
	if b.counting {
		b.segments += strings.Count(s, SegmentTerminator)
	}
	return b.Buffer.WriteString(s)
}

// Envelope holds ISA/IEA and GS/GE loop segments plus the transaction set wrapper.
type Envelope struct {
	ISA Segment
	GS  Segment
	ST  Segment
	SE  Segment
	GE  Segment
	IEA Segment
}

// ---------- Loop 1000A – Submitter Name ----------

// Loop1000A identifies the entity submitting the transaction.
type Loop1000A struct {
	NM1 Segment // NM1*41*2*SUBMITTER NAME*****46*SUBMITTER_ID~
	PER Segment // PER*IC*CONTACT NAME*TE*PHONE~
}

// ---------- Loop 1000B – Receiver Name ----------

// Loop1000B identifies the entity receiving the transaction (typically the payer).
type Loop1000B struct {
	NM1 Segment // NM1*40*2*RECEIVER NAME*****46*RECEIVER_ID~
}

// ---------- Helper builders ----------

// NewSegment is a convenience constructor for Segment.
func NewSegment(tag string, elements ...string) Segment {
	return Segment{Tag: tag, Elements: elements}
}

// formatDollars converts a float64 dollar amount to the X12 monetary format
// (no decimal point, two implied decimals).  100.00 → "10000".
func formatDollars(amount float64) string {
	cents := int(amount*100 + 0.5)
	return intToStr(cents)
}

// intToStr is a simple int-to-string helper (avoiding fmt for performance).
func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte(n%10) + '0'
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
