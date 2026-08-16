package common

import (
	"testing"
	"time"
)

func TestParseDate_Formats(t *testing.T) {
	cases := []struct {
		in   string
		want time.Time
	}{
		{"2024-01-15", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{"01/15/2024", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{"2024-01-15 13:45:00", time.Date(2024, 1, 15, 13, 45, 0, 0, time.UTC)},
		{"01/15/2024 13:45:00", time.Date(2024, 1, 15, 13, 45, 0, 0, time.UTC)},
		{"20240115", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{"01-15-2024", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
	}
	for _, tc := range cases {
		got, err := ParseDate(tc.in)
		if err != nil {
			t.Errorf("ParseDate(%q) unexpected error: %v", tc.in, err)
			continue
		}
		if !got.Equal(tc.want) {
			t.Errorf("ParseDate(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestParseDate_Empty(t *testing.T) {
	if _, err := ParseDate(""); err == nil {
		t.Error(`ParseDate("") expected error, got nil`)
	}
}
