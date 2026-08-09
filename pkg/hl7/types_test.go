package hl7

import (
	"testing"
)

// TestUnmarshal_ADT_A04 verifies parsing a standard ADT^A04 registration message.
func TestUnmarshal_ADT_A04(t *testing.T) {
	raw := "MSH|^~\\&|SENDING_APP|SENDING_FAC|RECEIVING_APP|RECEIVING_FAC|20240101080000||ADT^A04|MSG00001|P|2.5\r" +
		"PID|1||MRN12345^^^HOSPITAL^MR||DOE^JOHN^||19800101|M|||123 MAIN ST^^HARTFORD^CT^06101||(860)555-1212|||S|||123456789\r" +
		"PV1|1|I|||||||||||||||||V00001"

	msg, err := Unmarshal([]byte(raw))
	if err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if len(msg.Segments) != 3 {
		t.Errorf("expected 3 segments, got %d", len(msg.Segments))
	}

	msh := msg.FindSegment("MSH")
	if msh == nil {
		t.Fatal("MSH segment not found")
	}
	if msh.GetField(9) != "ADT^A04" {
		t.Errorf("expected MSH-9 = 'ADT^A04', got %q", msh.GetField(9))
	}

	pid := msg.FindSegment("PID")
	if pid == nil {
		t.Fatal("PID segment not found")
	}
	if pid.GetComponent(3, 1) != "MRN12345" {
		t.Errorf("expected PID-3.1 = 'MRN12345', got %q", pid.GetComponent(3, 1))
	}
	if pid.GetComponent(5, 1) != "DOE" {
		t.Errorf("expected PID-5.1 = 'DOE', got %q", pid.GetComponent(5, 1))
	}
	if pid.GetComponent(5, 2) != "JOHN" {
		t.Errorf("expected PID-5.2 = 'JOHN', got %q", pid.GetComponent(5, 2))
	}
	if pid.GetField(7) != "19800101" {
		t.Errorf("expected PID-7 = '19800101', got %q", pid.GetField(7))
	}
	if pid.GetField(8) != "M" {
		t.Errorf("expected PID-8 = 'M', got %q", pid.GetField(8))
	}
	if pid.GetField(19) != "123456789" {
		t.Errorf("expected PID-19 = '123456789', got %q", pid.GetField(19))
	}
}

// TestUnmarshal_empty returns error.
func TestUnmarshal_empty(t *testing.T) {
	_, err := Unmarshal([]byte{})
	if err == nil {
		t.Error("expected error for empty input")
	}
}

// TestUnmarshal_whitespace_only returns error.
func TestUnmarshal_whitespace_only(t *testing.T) {
	_, err := Unmarshal([]byte("   \r   \r"))
	if err == nil {
		t.Error("expected error for whitespace-only input")
	}
}

// TestMarshalRoundTrip verifies that marshal(unmarshal(x)) == x.
func TestMarshalRoundTrip(t *testing.T) {
	raw := "MSH|^~\\&|SENDING_APP|SENDING_FAC|RECEIVING_APP|RECEIVING_FAC|20240101080000||ADT^A04|MSG00001|P|2.5\r" +
		"PID|1||MRN12345^^^HOSPITAL^MR||DOE^JOHN^||19800101|M|||123 MAIN ST^^HARTFORD^CT^06101||(860)555-1212|||S|||123456789"

	msg, err := Unmarshal([]byte(raw))
	if err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	marshaled := string(msg.Marshal())
	if marshaled != raw {
		t.Errorf("round-trip mismatch:\n  original: %q\n marshaled: %q", raw, marshaled)
	}
}

// TestSegment_GetField_outOfBounds returns empty.
func TestSegment_GetField_outOfBounds(t *testing.T) {
	seg := &Segment{
		Name:   "PID",
		Fields: [][]string{{"1"}, {"DOE^JOHN"}},
	}

	if got := seg.GetField(0); got != "" {
		t.Errorf("expected empty for index 0, got %q", got)
	}
	if got := seg.GetField(3); got != "" {
		t.Errorf("expected empty for index 3, got %q", got)
	}
}

// TestSegment_GetComponent returns correct subfield.
func TestSegment_GetComponent(t *testing.T) {
	seg := &Segment{
		Name: "PID",
		Fields: [][]string{
			{"1"},                         // field 1
			{""},                          // field 2
			{"DOE", "JOHN", "Q"},          // field 3
		},
	}

	if got := seg.GetComponent(3, 1); got != "DOE" {
		t.Errorf("expected 'DOE', got %q", got)
	}
	if got := seg.GetComponent(3, 2); got != "JOHN" {
		t.Errorf("expected 'JOHN', got %q", got)
	}
	if got := seg.GetComponent(3, 3); got != "Q" {
		t.Errorf("expected 'Q', got %q", got)
	}
	if got := seg.GetComponent(3, 4); got != "" {
		t.Errorf("expected empty for out-of-bounds component, got %q", got)
	}
	if got := seg.GetComponent(0, 1); got != "" {
		t.Errorf("expected empty for field 0, got %q", got)
	}
}

// TestFindSegment verifies FindSegment and FindAllSegments.
func TestFindSegment(t *testing.T) {
	raw := "MSH|^~\\&|SENDING_APP|SENDING_FAC|||20240101080000||ADT^A04|MSG00001|P|2.5\r" +
		"PID|1||MRN12345\r" +
		"OBX|1|NM|12345-6^Glucose||100|mg/dL|70-110|H|||F\r" +
		"OBX|2|NM|12345-7^Sodium||140|mmol/L|135-145|N|||F"

	msg, err := Unmarshal([]byte(raw))
	if err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if s := msg.FindSegment("MSH"); s == nil {
		t.Error("FindSegment('MSH') returned nil")
	}
	if s := msg.FindSegment("NONEXISTENT"); s != nil {
		t.Error("FindSegment('NONEXISTENT') should return nil")
	}

	obx := msg.FindAllSegments("OBX")
	if len(obx) != 2 {
		t.Errorf("expected 2 OBX segments, got %d", len(obx))
	}
}

// TestUnmarshal_CRLF handles \r\n segment separators.
func TestUnmarshal_CRLF(t *testing.T) {
	raw := "MSH|^~\\&|APP|FAC|||20240101080000||ADT^A04|MSG00001|P|2.5\r\n" +
		"PID|1||MRN12345"

	msg, err := Unmarshal([]byte(raw))
	if err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if len(msg.Segments) != 2 {
		t.Errorf("expected 2 segments, got %d", len(msg.Segments))
	}
}

// TestMarshal_emptyMessage produces empty bytes.
func TestMarshal_emptyMessage(t *testing.T) {
	msg := &Message{}
	got := string(msg.Marshal())
	if got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}
