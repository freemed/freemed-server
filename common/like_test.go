package common

import "testing"

func TestEscapeLikePattern(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		// Ordinary values pass through untouched, including non-ASCII and
		// characters that are not LIKE metacharacters.
		{"", ""},
		{"smith", "smith"},
		{"O'Brien", "O'Brien"},
		{"123-45-6789", "123-45-6789"},
		{"Müller", "Müller"},
		// Metacharacters are escaped with a backslash.
		{"%", `\%`},
		{"_", `\_`},
		{`\`, `\\`},
		{"100%", `100\%`},
		{"a_b", `a\_b`},
		{"%_\\", `\%\_\\`},
		{"%%", `\%\%`},
		// A wildcard-only value must no longer be a bare pattern.
		{"%smith%", `\%smith\%`},
	}
	for _, tc := range cases {
		if got := EscapeLikePattern(tc.in); got != tc.want {
			t.Errorf("EscapeLikePattern(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestEscapeLikePatternIsIdempotentForCleanInput documents that escaping is not
// applied twice to a value that contains nothing to escape.
func TestEscapeLikePatternNoOpForCleanInput(t *testing.T) {
	in := "Patient Name"
	if got := EscapeLikePattern(in); got != in {
		t.Errorf("EscapeLikePattern(%q) = %q, want unchanged", in, got)
	}
}
