package api

import "testing"

func TestParseControls(t *testing.T) {
	sample := `<?xml version="1.0"?>
<form>
  <controls>
    <control variable="patient_name" name="Patient Name" type="string" default="" options="" limits="30" uuid="abc-123" />
    <control variable="is_smoker" name="Smoker" type="boolean" />
    <control variable="gender" name="Gender" type="select" default="male" options="male|female|other" />
  </controls>
</form>`

	controls, err := parseTemplateControls(sample)
	if err != nil {
		t.Fatalf("parseTemplateControls returned error: %v", err)
	}

	if len(controls) != 3 {
		t.Fatalf("expected 3 controls, got %d", len(controls))
	}

	// First control has all attributes present.
	c0 := controls[0]
	if c0.Variable != "patient_name" {
		t.Errorf("c0.Variable = %q, want %q", c0.Variable, "patient_name")
	}
	if c0.Name != "Patient Name" {
		t.Errorf("c0.Name = %q, want %q", c0.Name, "Patient Name")
	}
	if c0.Type != "string" {
		t.Errorf("c0.Type = %q, want %q", c0.Type, "string")
	}
	if c0.Limits != "30" {
		t.Errorf("c0.Limits = %q, want %q", c0.Limits, "30")
	}
	if c0.UUID != "abc-123" {
		t.Errorf("c0.UUID = %q, want %q", c0.UUID, "abc-123")
	}

	// Second control omits several attributes; they must be empty strings.
	c1 := controls[1]
	if c1.Variable != "is_smoker" || c1.Type != "boolean" {
		t.Errorf("c1 = %+v, want variable=is_smoker type=boolean", c1)
	}
	if c1.Default != "" || c1.Options != "" || c1.Limits != "" || c1.UUID != "" {
		t.Errorf("c1 optional attributes should be empty, got %+v", c1)
	}

	// Third control exercises select options.
	c2 := controls[2]
	if c2.Variable != "gender" || c2.Type != "select" {
		t.Errorf("c2 = %+v, want variable=gender type=select", c2)
	}
	if c2.Default != "male" {
		t.Errorf("c2.Default = %q, want %q", c2.Default, "male")
	}
	if c2.Options != "male|female|other" {
		t.Errorf("c2.Options = %q, want %q", c2.Options, "male|female|other")
	}
}

func TestParseControlsEmptyAndMissingUUID(t *testing.T) {
	// A control with no uuid must fall back to variable via controlUUID.
	controls, err := parseTemplateControls(`<controls><control variable="x" name="X" type="string"/></controls>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(controls) != 1 {
		t.Fatalf("expected 1 control, got %d", len(controls))
	}
	if got := controlUUID(controls[0]); got != "x" {
		t.Errorf("controlUUID fallback = %q, want %q", got, "x")
	}

	// Empty string parses to zero controls without error.
	empty, err := parseTemplateControls("")
	if err != nil {
		t.Fatalf("unexpected error on empty input: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("expected 0 controls for empty input, got %d", len(empty))
	}
}
