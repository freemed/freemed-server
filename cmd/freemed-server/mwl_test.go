package main

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/freemed/freemed-server/pkg/dicomnet"
)

// TestMwlQuerySQLIsParameterised guards the security property that no value is
// ever concatenated into the worklist statement.
func TestMwlQuerySQLIsParameterised(t *testing.T) {
	// The only quoted literals allowed are empty strings (COALESCE defaults);
	// after removing them no quote may remain, so no value can be inlined.
	stripped := strings.ReplaceAll(mwlQuerySQL, "''", "")
	if strings.ContainsRune(stripped, '\'') {
		t.Error("the worklist statement contains a quoted literal; it must be fully parameterised")
	}
	if strings.Contains(mwlQuerySQL, "%") {
		t.Error("the worklist statement looks like it was built with a format string")
	}
	if got := strings.Count(mwlQuerySQL, "?"); got != 3 {
		t.Errorf("placeholders = %d, want 3 (window start, window end, limit)", got)
	}
}

func TestMwlWindow(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	t.Run("no query uses the default window", func(t *testing.T) {
		start, end := mwlWindow(nil)
		if !start.Equal(today) {
			t.Errorf("start = %v, want today (%v)", start, today)
		}
		if want := today.AddDate(0, 0, mwlDefaultWindowDays); !end.Equal(want) {
			t.Errorf("end = %v, want %v", end, want)
		}
	})

	t.Run("explicit range is honoured, end inclusive", func(t *testing.T) {
		q, err := dicomnet.ParseMWLQuery(dicomnet.DataSet{
			dicomnet.NewSequenceElement(dicomnet.TagScheduledProcedureStepSeq, dicomnet.DataSet{
				dicomnet.NewStringElement(dicomnet.TagSPSStartDate, "DA", "20240101-20240131"),
			}),
		})
		if err != nil {
			t.Fatalf("ParseMWLQuery: %v", err)
		}
		start, end := mwlWindow(q)
		if want := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local); !start.Equal(want) {
			t.Errorf("start = %v, want %v", start, want)
		}
		if want := time.Date(2024, 2, 1, 0, 0, 0, 0, time.Local); !end.Equal(want) {
			t.Errorf("end = %v, want %v", end, want)
		}
	})

	t.Run("single day range covers exactly that day", func(t *testing.T) {
		q, err := dicomnet.ParseMWLQuery(dicomnet.DataSet{
			dicomnet.NewSequenceElement(dicomnet.TagScheduledProcedureStepSeq, dicomnet.DataSet{
				dicomnet.NewStringElement(dicomnet.TagSPSStartDate, "DA", "20240115"),
			}),
		})
		if err != nil {
			t.Fatalf("ParseMWLQuery: %v", err)
		}
		start, end := mwlWindow(q)
		if !start.Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.Local)) {
			t.Errorf("start = %v", start)
		}
		if !end.Equal(time.Date(2024, 1, 16, 0, 0, 0, 0, time.Local)) {
			t.Errorf("end = %v", end)
		}
	})

	t.Run("open ended range starts today and is capped", func(t *testing.T) {
		q, err := dicomnet.ParseMWLQuery(dicomnet.DataSet{
			dicomnet.NewSequenceElement(dicomnet.TagScheduledProcedureStepSeq, dicomnet.DataSet{
				dicomnet.NewStringElement(dicomnet.TagSPSStartDate, "DA", "20240101-"),
			}),
		})
		if err != nil {
			t.Fatalf("ParseMWLQuery: %v", err)
		}
		start, end := mwlWindow(q)
		if !start.Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)) {
			t.Errorf("start = %v", start)
		}
		if got := end.Sub(start); got > mwlMaxWindowDays*24*time.Hour {
			t.Errorf("window = %v, exceeds the %d day cap", got, mwlMaxWindowDays)
		}
	})

	t.Run("a decade long range is capped", func(t *testing.T) {
		q, err := dicomnet.ParseMWLQuery(dicomnet.DataSet{
			dicomnet.NewSequenceElement(dicomnet.TagScheduledProcedureStepSeq, dicomnet.DataSet{
				dicomnet.NewStringElement(dicomnet.TagSPSStartDate, "DA", "20100101-20240101"),
			}),
		})
		if err != nil {
			t.Fatalf("ParseMWLQuery: %v", err)
		}
		start, end := mwlWindow(q)
		if got := end.Sub(start); got != mwlMaxWindowDays*24*time.Hour {
			t.Errorf("window = %v, want the %d day cap", got, mwlMaxWindowDays)
		}
	})

	t.Run("unparseable range falls back to the default window", func(t *testing.T) {
		q, err := dicomnet.ParseMWLQuery(dicomnet.DataSet{
			dicomnet.NewSequenceElement(dicomnet.TagScheduledProcedureStepSeq, dicomnet.DataSet{
				dicomnet.NewStringElement(dicomnet.TagSPSStartDate, "DA", "not-a-date"),
			}),
		})
		if err != nil {
			t.Fatalf("ParseMWLQuery: %v", err)
		}
		start, end := mwlWindow(q)
		if !start.Equal(today) {
			t.Errorf("start = %v, want today", start)
		}
		if want := today.AddDate(0, 0, mwlDefaultWindowDays); !end.Equal(want) {
			t.Errorf("end = %v, want %v", end, want)
		}
	})
}

func TestMwlModality(t *testing.T) {
	tests := []struct {
		apptType string
		prenote  string
		want     string
	}{
		{"CT Scan", "", "CT"},
		{"ct scan chest", "", "CT"},
		{"CT Scan with contrast", "rule out PE", "CT"},
		{"MRI Brain", "", "MR"},
		{"magnetic resonance imaging", "", "MR"},
		{"Ultrasound abdomen", "", "US"},
		{"Echocardiogram", "", "US"},
		{"Mammogram screening", "", "MG"},
		{"DEXA scan", "", "BMD"},
		{"Nuclear medicine scan", "", "NM"},
		{"PET Scan", "", "PT"},
		{"Cardiac catheterization", "", "XA"},
		{"Chest X-ray", "", "DX"},
		{"XRAY knee", "", "DX"},
		{"EKG", "", "ECG"},
		{"Office Visit", "", "OT"},
		{"", "", "OT"},
		{"Follow up", "physical therapy", "OT"},
	}
	for _, tc := range tests {
		if got := mwlModality(tc.apptType, tc.prenote); got != tc.want {
			t.Errorf("mwlModality(%q, %q) = %q, want %q", tc.apptType, tc.prenote, got, tc.want)
		}
	}
}

func TestMwlStepStatus(t *testing.T) {
	tests := map[string]string{
		"":             "SCHEDULED",
		"scheduled":    "SCHEDULED",
		"   ":          "SCHEDULED",
		"cancelled":    "CANCELLED",
		"Canceled":     "CANCELLED",
		"no show":      "CANCELLED",
		"completed":    "COMPLETED",
		"Checked Out":  "COMPLETED",
		"arrived":      "ARRIVED",
		"checked in":   "ARRIVED",
		"in progress":  "STARTED",
		"unrecognised": "SCHEDULED",
	}
	for in, want := range tests {
		if got := mwlStepStatus(in); got != want {
			t.Errorf("mwlStepStatus(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMwlPersonName(t *testing.T) {
	tests := []struct {
		last, first, middle string
		want                string
	}{
		{"Smith", "John", "Q", "Smith^John^Q"},
		{"Smith", "John", "", "Smith^John"},
		{"Smith", "", "", "Smith"},
		{"", "", "", ""},
		{"  Smith ", " John ", "", "Smith^John"},
		// Component separators in the source data must not forge components.
		{"Smi^th", "Jo\\hn", "=Q", "Smi th^Jo hn^Q"},
		{"Smith\x00", "John", "", "Smith^John"},
	}
	for _, tc := range tests {
		if got := mwlPersonName(tc.last, tc.first, tc.middle); got != tc.want {
			t.Errorf("mwlPersonName(%q,%q,%q) = %q, want %q", tc.last, tc.first, tc.middle, got, tc.want)
		}
	}
}

func TestMwlUIDIsDeterministicAndValid(t *testing.T) {
	first := mwlUID("study", 42)
	second := mwlUID("study", 42)
	if first != second {
		t.Fatalf("mwlUID is not deterministic: %q != %q", first, second)
	}
	if other := mwlUID("study", 43); other == first {
		t.Error("distinct rows produced the same worklist UID")
	}
	if !strings.HasPrefix(first, "2.25.") {
		t.Errorf("UID %q does not use the 2.25 numeric root", first)
	}
	if len(first) > 64 {
		t.Errorf("UID %q is %d characters (DICOM limit is 64)", first, len(first))
	}
	for _, r := range first[5:] {
		if r < '0' || r > '9' {
			t.Fatalf("UID %q contains a non-numeric character", first)
		}
	}
}

func TestMwlSexAndText(t *testing.T) {
	tests := map[string]string{
		"M": "M", "m": "M", "Male": "M", "male": "M",
		"F": "F", "Female": "F",
		"O": "O", "Other": "O",
		"": "", "unknown": "",
	}
	for in, want := range tests {
		if got := mwlSex(in); got != want {
			t.Errorf("mwlSex(%q) = %q, want %q", in, got, want)
		}
	}
	if got := mwlText("  value  ", 32); got != "value" {
		t.Errorf("mwlText trim = %q", got)
	}
	if got := mwlText("abcdef", 3); got != "abc" {
		t.Errorf("mwlText truncate = %q", got)
	}
	if got := mwlText("abcdef", 0); got != "abcdef" {
		t.Errorf("mwlText unbounded = %q", got)
	}
	// The backslash is the DICOM multi-value separator: a value containing one
	// must not be able to split the returned attribute into extra values.
	if got := mwlText("CHEST\\PA", 32); got != "CHEST PA" {
		t.Errorf("mwlText backslash = %q, want \"CHEST PA\"", got)
	}
	if got := mwlText("CHEST\x00X", 32); got != "CHESTX" {
		t.Errorf("mwlText NUL = %q, want CHESTX", got)
	}
	if got := mwlText("CHEST	PA", 32); got != "CHEST PA" {
		t.Errorf("mwlText control character = %q, want CHEST PA", got)
	}
}

// TestMwlItemProjection checks the scheduler row to MWL entry mapping without a
// database: the date/hour/minute columns drive the scheduled start, and the
// patient, physician and facility joins populate the response attributes.
func TestMwlItemProjection(t *testing.T) {
	src := newMwlSource(nil, "FREEMED")
	calDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.Local)
	item := src.item(
		17, calDate, 9, 30, 45,
		"scheduled", "CT Scan Chest", "rule out PE",
		"MRN0001234", "Smith", "John", "Q",
		sql.NullTime{Time: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true},
		"male",
		"Wilson", "James", "",
		"FREEMED CLINIC",
	)

	if item.PatientName != "Smith^John^Q" {
		t.Errorf("PatientName = %q", item.PatientName)
	}
	if item.PatientID != "MRN0001234" {
		t.Errorf("PatientID = %q", item.PatientID)
	}
	if item.PatientBirthDate != "19700101" {
		t.Errorf("PatientBirthDate = %q", item.PatientBirthDate)
	}
	if item.PatientSex != "M" {
		t.Errorf("PatientSex = %q", item.PatientSex)
	}
	if item.ScheduledProcedureStepStartDate != "20240115" {
		t.Errorf("start date = %q", item.ScheduledProcedureStepStartDate)
	}
	if item.ScheduledProcedureStepStartTime != "093000" {
		t.Errorf("start time = %q", item.ScheduledProcedureStepStartTime)
	}
	if item.ScheduledProcedureStepEndTime != "101500" {
		t.Errorf("end time = %q (45 minute appointment)", item.ScheduledProcedureStepEndTime)
	}
	if item.Modality != "CT" {
		t.Errorf("Modality = %q", item.Modality)
	}
	if item.ScheduledStationAETitle != "FREEMED" {
		t.Errorf("ScheduledStationAETitle = %q", item.ScheduledStationAETitle)
	}
	if item.ScheduledPerformingPhysicianName != "Wilson^James" {
		t.Errorf("performing physician = %q", item.ScheduledPerformingPhysicianName)
	}
	if item.InstitutionName != "FREEMED CLINIC" {
		t.Errorf("InstitutionName = %q", item.InstitutionName)
	}
	if item.ScheduledProcedureStepStatus != "SCHEDULED" {
		t.Errorf("SPS status = %q", item.ScheduledProcedureStepStatus)
	}
	if item.AccessionNumber == "" || item.StudyInstanceUID == "" ||
		item.RequestedProcedureID == "" || item.ScheduledProcedureStepID == "" {
		t.Errorf("identifiers are incomplete: %+v", item)
	}
	if !strings.Contains(item.ScheduledProcedureStepDescription, "CT Scan Chest") {
		t.Errorf("SPS description = %q", item.ScheduledProcedureStepDescription)
	}

	// Requested return keys must be reproducible through the query model.
	q, err := dicomnet.ParseMWLQuery(dicomnet.DataSet{
		dicomnet.NewStringElement(dicomnet.TagPatientName, "PN", "Smith*"),
		dicomnet.NewStringElement(dicomnet.TagPatientID, "LO", ""),
		dicomnet.NewSequenceElement(dicomnet.TagScheduledProcedureStepSeq, dicomnet.DataSet{
			dicomnet.NewStringElement(dicomnet.TagModality, "CS", "CT"),
			dicomnet.NewStringElement(dicomnet.TagSPSStartDate, "DA", "20240115"),
			dicomnet.NewStringElement(dicomnet.TagScheduledStationAETitle, "AE", ""),
		}),
	})
	if err != nil {
		t.Fatalf("ParseMWLQuery: %v", err)
	}
	if !q.MatchItem(item) {
		t.Fatal("the projected item does not satisfy a matching CT query")
	}
	resp, err := q.BuildResponseIdentifier(item)
	if err != nil {
		t.Fatalf("BuildResponseIdentifier: %v", err)
	}
	if v := resp.GetString(dicomnet.TagPatientName); v != "Smith^John^Q" {
		t.Errorf("response PatientName = %q", v)
	}
	sps, ok := resp.GetSequence(dicomnet.TagScheduledProcedureStepSeq)
	if !ok {
		t.Fatal("response has no SPS sequence")
	}
	if v := sps.GetString(dicomnet.TagModality); v != "CT" {
		t.Errorf("response Modality = %q", v)
	}
	if v := sps.GetString(dicomnet.TagSPSStartDate); v != "20240115" {
		t.Errorf("response SPS start date = %q", v)
	}

	// An appointment outside the window must not match.
	other := src.item(18, calDate, 9, 30, 30, "scheduled", "CT Scan", "",
		"MRN0001234", "Smith", "Jane", "", sql.NullTime{}, "", "", "", "", "")
	if other.PatientBirthDate != "" {
		t.Errorf("missing date of birth should be empty, got %q", other.PatientBirthDate)
	}
}

func TestMwlSourceWithoutDatabase(t *testing.T) {
	src := newMwlSource(nil, "FREEMED")
	if _, err := src.Find(context.Background(), nil); err == nil {
		t.Fatal("Find succeeded without a database handle")
	}
}

func TestStartMwlServerDisabledByDefault(t *testing.T) {
	// Mwl.Port defaults to 0, which disables the listener entirely.
	srv, err := startMwlServer(context.Background())
	if err != nil {
		t.Fatalf("startMwlServer: %v", err)
	}
	if srv != nil {
		t.Fatal("the worklist SCP started although mwl.port is 0")
	}
}
