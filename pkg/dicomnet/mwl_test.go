package dicomnet

import (
	"testing"
)

// standardIdentifier returns a realistic Modality Worklist C-FIND identifier:
// a matching modality, a wildcarded patient name, a date range, and a set of
// return keys.
func standardIdentifier() DataSet {
	sps := DataSet{
		NewStringElement(TagModality, "CS", "CT"),
		NewStringElement(TagScheduledStationAETitle, "AE", ""),
		NewStringElement(TagSPSStartDate, "DA", "20240101-20240131"),
		NewStringElement(TagSPSStartTime, "TM", ""),
		NewStringElement(TagScheduledPerformingPhysName, "PN", ""),
		NewStringElement(TagSPSID, "SH", ""),
		NewStringElement(TagSPSDescription, "LO", ""),
	}
	return DataSet{
		NewStringElement(TagPatientName, "PN", "SMITH*"),
		NewStringElement(TagPatientID, "LO", ""),
		NewStringElement(TagAccessionNumber, "SH", ""),
		NewStringElement(TagStudyInstanceUID, "UI", ""),
		NewStringElement(TagRequestedProcedureID, "SH", ""),
		NewSequenceElement(TagScheduledProcedureStepSeq, sps),
	}
}

func sampleItem() MWLItem {
	return MWLItem{
		PatientName:                       "SMITH^JOHN",
		PatientID:                         "MRN0001234",
		PatientBirthDate:                  "19700101",
		PatientSex:                        "M",
		AccessionNumber:                   "ACC1001",
		StudyInstanceUID:                  "2.25.1234567890",
		StudyDescription:                  "CT CHEST W/O",
		InstitutionName:                   "FREEMED CLINIC",
		ReferringPhysicianName:            "HOUSE^GREGORY",
		RequestedProcedureID:              "RP1001",
		RequestedProcedureDescription:     "CT CHEST",
		Modality:                          "CT",
		ScheduledStationAETitle:           "CT1",
		ScheduledProcedureStepStartDate:   "20240115",
		ScheduledProcedureStepStartTime:   "090000",
		ScheduledProcedureStepEndDate:     "20240115",
		ScheduledProcedureStepEndTime:     "091500",
		ScheduledPerformingPhysicianName:  "WILSON^JAMES",
		ScheduledProcedureStepID:          "SPS1001",
		ScheduledProcedureStepDescription: "CT CHEST W/O CONTRAST",
		ScheduledProcedureStepLocation:    "CT ROOM 1",
		ScheduledProcedureStepStatus:      "SCHEDULED",
		RequestedContrastAgent:            "",
	}
}

// ---------------------------------------------------------------------------
// Query parsing
// ---------------------------------------------------------------------------

func TestParseMWLQuery(t *testing.T) {
	q, err := ParseMWLQuery(standardIdentifier())
	if err != nil {
		t.Fatalf("ParseMWLQuery: %v", err)
	}
	if v, ok := q.MatchingValue(TagPatientName); !ok || v != "SMITH*" {
		t.Errorf("PatientName matching value = %q/%v", v, ok)
	}
	if _, ok := q.MatchingValue(TagPatientID); ok {
		t.Error("zero-length PatientID must not be a matching key")
	}
	if v, ok := q.SPSMatchingValue(TagModality); !ok || v != "CT" {
		t.Errorf("Modality matching value = %q/%v", v, ok)
	}
	if _, ok := q.SPSMatchingValue(TagSPSStartTime); ok {
		t.Error("zero-length start time must not be a matching key")
	}

	dr, err := q.DateRange()
	if err != nil {
		t.Fatalf("DateRange: %v", err)
	}
	if dr.Universal || dr.Start != "20240101" || dr.End != "20240131" {
		t.Errorf("date range = %+v", dr)
	}
	tr, err := q.TimeRange()
	if err != nil {
		t.Fatalf("TimeRange: %v", err)
	}
	if !tr.Universal {
		t.Errorf("empty time value should be universal matching, got %+v", tr)
	}

	// Every identifier attribute must survive as a key, with the return keys
	// marked as such.
	var sawReturnKey, sawModuleKey bool
	for _, k := range q.Keys {
		if k.Tag == TagPatientID {
			if !k.IsReturnKey() {
				t.Error("PatientID should be a return key")
			}
			sawReturnKey = true
		}
		if k.Tag == TagStudyInstanceUID && !k.InSPS && !k.IsReturnKey() {
			t.Error("StudyInstanceUID should be a return key")
		}
		if k.Tag == TagModality {
			sawModuleKey = true
			if !k.InSPS {
				t.Error("Modality key should be flagged as belonging to the SPS sequence")
			}
		}
	}
	if !sawReturnKey || !sawModuleKey {
		t.Fatalf("keys incomplete: %+v", q.Keys)
	}
	if len(q.UnsupportedMatchingKeys) != 0 {
		t.Errorf("unexpected unsupported matching keys: %v", q.UnsupportedMatchingKeys)
	}

	// Round trip through the wire format so the parser and encoder agree.
	for _, explicit := range []bool{true, false} {
		raw, err := EncodeDataSet(standardIdentifier(), explicit)
		if err != nil {
			t.Fatalf("explicit=%v EncodeDataSet: %v", explicit, err)
		}
		ds, err := DecodeDataSet(raw, explicit)
		if err != nil {
			t.Fatalf("explicit=%v DecodeDataSet: %v", explicit, err)
		}
		q2, err := ParseMWLQuery(ds)
		if err != nil {
			t.Fatalf("explicit=%v ParseMWLQuery: %v", explicit, err)
		}
		if v, _ := q2.MatchingValue(TagPatientName); v != "SMITH*" {
			t.Errorf("explicit=%v: PatientName = %q", explicit, v)
		}
		if dr2, err := q2.DateRange(); err != nil || dr2.Start != "20240101" || dr2.End != "20240131" {
			t.Errorf("explicit=%v: date range = %+v (%v)", explicit, dr2, err)
		}
		if v, _ := q2.SPSMatchingValue(TagModality); v != "CT" {
			t.Errorf("explicit=%v: Modality = %q", explicit, v)
		}
	}
}

func TestParseMWLQueryEmptyAndInvalid(t *testing.T) {
	q, err := ParseMWLQuery(nil)
	if err != nil {
		t.Fatalf("empty identifier: %v", err)
	}
	if !q.MatchItem(sampleItem()) {
		t.Error("an empty identifier must match every item")
	}

	// A tag value of zero is not a valid data set element.
	if _, err := ParseMWLQuery(DataSet{{Tag: 0, VR: "UN"}}); err == nil {
		t.Error("ParseMWLQuery accepted a zero tag")
	}
	// Item tags cannot appear in an identifier.
	if _, err := ParseMWLQuery(DataSet{{Tag: TagItem, VR: "UN"}}); err == nil {
		t.Error("ParseMWLQuery accepted an item tag")
	}
	// A sequence element with no items is legal but must not produce SPS keys.
	q, err = ParseMWLQuery(DataSet{NewSequenceElement(TagScheduledProcedureStepSeq)})
	if err != nil {
		t.Fatalf("empty SPS sequence: %v", err)
	}
	if len(q.Keys) != 1 {
		t.Errorf("keys = %d, want 1", len(q.Keys))
	}
}

func TestUnsupportedMatchingKeysAreReportedNotFiltered(t *testing.T) {
	ds := DataSet{
		NewStringElement(TagRequestedProcedurePriority, "SH", "HIGH"),
		NewStringElement(TagPatientID, "LO", "MRN0001234"),
	}
	q, err := ParseMWLQuery(ds)
	if err != nil {
		t.Fatalf("ParseMWLQuery: %v", err)
	}
	if len(q.UnsupportedMatchingKeys) != 1 || q.UnsupportedMatchingKeys[0] != TagRequestedProcedurePriority {
		t.Fatalf("unsupported matching keys = %v", q.UnsupportedMatchingKeys)
	}
	// The unsupported key must not filter anything out.
	if !q.MatchItem(sampleItem()) {
		t.Error("an unsupported matching key filtered out a matching item")
	}
	if q.MatchItem(MWLItem{PatientID: "OTHER"}) {
		t.Error("a supported matching key was ignored")
	}
}

// ---------------------------------------------------------------------------
// Wildcard matching
// ---------------------------------------------------------------------------

func TestMatchWildcard(t *testing.T) {
	tests := []struct {
		pattern string
		value   string
		want    bool
	}{
		{"", "", true},
		{"", "X", false},
		{"*", "", true},
		{"*", "ANYTHING", true},
		{"SMITH^JOHN", "SMITH^JOHN", true},
		{"SMITH^JOHN", "SMITH^JANE", false},
		{"smith^john", "SMITH^JOHN", true}, // case insensitive
		{"SMITH*", "SMITH^JOHN", true},
		{"SMITH*", "SMITHSON", true},
		{"SMITH*", "SMYTH", false},
		{"*JOHN", "SMITH^JOHN", true},
		{"*^JOHN", "SMITH^JOHN", true},
		{"SM*JOHN", "SMITH^JOHN", true},
		{"SM?TH", "SMITH", true},
		{"SM?TH", "SMTH", false}, // ? matches exactly one character
		{"SM?TH", "SMITH^JOHN", false},
		{"?", "", false},
		{"**", "ANYTHING", true},
		{"*A*", "BAB", true},
		// A literal '*' stored in the attribute value.
		{"STAR*NAME", "STAR*NAME", true},
		{"STAR*", "STAR*NAME", true},
		{"*STAR*NAME", "STAR*NAME", true},
		{"STAR?NAME", "STAR*NAME", true}, // ? matches the literal '*'
		// A literal '?' stored in the attribute value.
		{"A?C", "A?C", true},
		{"A*C", "A?C", true},
		{"WHO?^NAME", "WHO?^NAME", true},
		{"W?O?^NAME", "WHO?^NAME", true},
		// No regular expression metacharacters are honoured.
		{"A.B", "AXB", false},
		{"A.B", "A.B", true},
		{"A+", "AAA", false},
		{"O'BRIEN", "O'BRIEN", true},
		{"*BRIEN*", "RUTH O'BRIEN", true},
		{"MÜLLER*", "MÜLLER^JÖRG", true},
	}
	for _, tc := range tests {
		if got := MatchWildcard(tc.pattern, tc.value); got != tc.want {
			t.Errorf("MatchWildcard(%q, %q) = %v, want %v", tc.pattern, tc.value, got, tc.want)
		}
	}
}

func TestMatchWildcardHostilePatterns(t *testing.T) {
	// Patterns full of wildcards must terminate quickly rather than blow up.
	long := ""
	for i := 0; i < 200; i++ {
		long += "*a"
	}
	if MatchWildcard(long, "bababababababababababab") {
		// The exact answer is not important; termination without panic is.
		t.Log("long pattern matched (fine)")
	}
	if MatchWildcard("****", "") != true {
		t.Error("trailing wildcards must match an empty value")
	}
}

func TestMatchMultiValue(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		value   string
		want    bool
	}{
		{"single against single", "SMITH^JOHN", "SMITH^JOHN", true},
		{"single against first component", "SMITH^JOHN", "SMITH^JOHN\\DOE^JANE", true},
		{"single against second component", "DOE^JANE", "SMITH^JOHN\\DOE^JANE", true},
		{"single against no component", "JONES^BOB", "SMITH^JOHN\\DOE^JANE", false},
		{"wildcard against multi-valued", "SMITH*", "SMITH^JOHN\\DOE^JANE", true},
		{"multi-valued pattern, both present", "SMITH^JOHN\\DOE^JANE", "SMITH^JOHN\\DOE^JANE", true},
		{"multi-valued pattern, order independent", "DOE^JANE\\SMITH^JOHN", "SMITH^JOHN\\DOE^JANE", true},
		{"multi-valued pattern, one missing", "SMITH^JOHN\\JONES^BOB", "SMITH^JOHN\\DOE^JANE", false},
		{"empty pattern is universal", "", "", true},
		{"empty pattern against a value", "", "ANY", true},
		{"wildcard against empty value", "*", "", true},
		{"non-empty pattern against empty value", "A", "", false},
		{"backslash-only pattern", "\\", "A", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := MatchMultiValue(tc.pattern, tc.value); got != tc.want {
				t.Errorf("MatchMultiValue(%q, %q) = %v, want %v", tc.pattern, tc.value, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Date and time ranges
// ---------------------------------------------------------------------------

func TestDateRangeMatching(t *testing.T) {
	tests := []struct {
		spec      string
		date      string
		want      bool
		wantErr   bool
		universal bool
	}{
		{spec: "", date: "20240115", want: true, universal: true},
		{spec: "*", date: "20240115", want: true, universal: true},
		{spec: "*", date: "", want: true, universal: true},
		{spec: "20240115", date: "20240115", want: true},
		{spec: "20240115", date: "20240116", want: false},
		{spec: "20240101-20240131", date: "20240101", want: true},
		{spec: "20240101-20240131", date: "20240131", want: true},
		{spec: "20240101-20240131", date: "20240115", want: true},
		{spec: "20240101-20240131", date: "20231231", want: false},
		{spec: "20240101-20240131", date: "20240201", want: false},
		{spec: "20240101-20240131", date: "", want: false},
		{spec: "20240101-", date: "20250101", want: true},
		{spec: "20240101-", date: "20231231", want: false},
		{spec: "-20240131", date: "20200101", want: true},
		{spec: "-20240131", date: "20240201", want: false},
		{spec: "-", date: "20240201", want: true, universal: true},
		{spec: "2024", date: "20240101", wantErr: true},
		{spec: "2024010a", date: "20240101", wantErr: true},
		{spec: "20240101-2024", date: "20240101", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.spec+"/"+tc.date, func(t *testing.T) {
			r, err := ParseDateRange(tc.spec)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseDateRange(%q) accepted an invalid range", tc.spec)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseDateRange(%q): %v", tc.spec, err)
			}
			if r.Universal != tc.universal {
				t.Errorf("Universal = %v, want %v", r.Universal, tc.universal)
			}
			if got := r.Matches(tc.date); got != tc.want {
				t.Errorf("Matches(%q) = %v, want %v", tc.date, got, tc.want)
			}
		})
	}
}

func TestTimeRangeMatching(t *testing.T) {
	tests := []struct {
		spec    string
		value   string
		want    bool
		wantErr bool
	}{
		{spec: "", value: "090000", want: true},
		{spec: "090000", value: "090000", want: true},
		{spec: "090000", value: "090001", want: false},
		{spec: "0900", value: "0900", want: true},
		{spec: "0900-1700", value: "0900", want: true},
		{spec: "0900-1700", value: "1200", want: true},
		{spec: "0900-1700", value: "1700", want: true},
		{spec: "0900-1700", value: "0800", want: false},
		{spec: "0900-1700", value: "", want: false},
		{spec: "09", value: "090000", want: true},
		{spec: "900", value: "090000", wantErr: true},
		{spec: "09:00", value: "090000", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.spec+"/"+tc.value, func(t *testing.T) {
			r, err := ParseTimeRange(tc.spec)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseTimeRange(%q) accepted an invalid range", tc.spec)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseTimeRange(%q): %v", tc.spec, err)
			}
			if got := r.Matches(tc.value); got != tc.want {
				t.Errorf("Matches(%q) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}

func TestQueryDateRangePropagatesErrors(t *testing.T) {
	q, err := ParseMWLQuery(DataSet{
		NewSequenceElement(TagScheduledProcedureStepSeq, DataSet{
			NewStringElement(TagSPSStartDate, "DA", "not-a-date"),
		}),
	})
	if err != nil {
		t.Fatalf("ParseMWLQuery: %v", err)
	}
	if _, err := q.DateRange(); err == nil {
		t.Fatal("DateRange accepted an invalid date range")
	}
	if q.MatchItem(sampleItem()) {
		t.Error("an unparseable date range must not match items")
	}
}

// ---------------------------------------------------------------------------
// Matching and response construction
// ---------------------------------------------------------------------------

func TestQueryMatchItem(t *testing.T) {
	q, err := ParseMWLQuery(standardIdentifier())
	if err != nil {
		t.Fatalf("ParseMWLQuery: %v", err)
	}
	if !q.MatchItem(sampleItem()) {
		t.Error("the sample item should match the standard identifier")
	}

	// Wrong modality.
	item := sampleItem()
	item.Modality = "MR"
	if q.MatchItem(item) {
		t.Error("item with the wrong modality matched")
	}

	// Wrong patient (wildcard prefix).
	item = sampleItem()
	item.PatientName = "JONES^MARY"
	if q.MatchItem(item) {
		t.Error("item with the wrong patient name matched")
	}

	// Outside the date range.
	item = sampleItem()
	item.ScheduledProcedureStepStartDate = "20240215"
	if q.MatchItem(item) {
		t.Error("item outside the requested date range matched")
	}

	// A bare single date only matches that day.
	single, err := ParseMWLQuery(DataSet{
		NewSequenceElement(TagScheduledProcedureStepSeq, DataSet{
			NewStringElement(TagSPSStartDate, "DA", "20240115"),
		}),
	})
	if err != nil {
		t.Fatalf("ParseMWLQuery: %v", err)
	}
	if !single.MatchItem(sampleItem()) {
		t.Error("bare single date did not match the same day")
	}
	item = sampleItem()
	item.ScheduledProcedureStepStartDate = "20240116"
	if single.MatchItem(item) {
		t.Error("bare single date matched a different day")
	}

	// A nil query matches everything (defensive behaviour).
	var nilQuery *MWLQuery
	if !nilQuery.MatchItem(sampleItem()) {
		t.Error("nil query must match every item")
	}
}

func TestBuildResponseIdentifierContainsOnlyRequestedKeys(t *testing.T) {
	q, err := ParseMWLQuery(standardIdentifier())
	if err != nil {
		t.Fatalf("ParseMWLQuery: %v", err)
	}
	ds, err := q.BuildResponseIdentifier(sampleItem())
	if err != nil {
		t.Fatalf("BuildResponseIdentifier: %v", err)
	}

	// Five top-level attributes plus the SPS sequence were requested.
	if len(ds) != 6 {
		t.Fatalf("response has %d top-level elements, want 6: %+v", len(ds), ds)
	}
	if v := ds.GetString(TagPatientName); v != "SMITH^JOHN" {
		t.Errorf("PatientName = %q, want the item's value", v)
	}
	if v := ds.GetString(TagPatientID); v != "MRN0001234" {
		t.Errorf("PatientID = %q", v)
	}
	if v := ds.GetString(TagAccessionNumber); v != "ACC1001" {
		t.Errorf("AccessionNumber = %q", v)
	}
	if v := ds.GetString(TagStudyInstanceUID); v != "2.25.1234567890" {
		t.Errorf("StudyInstanceUID = %q", v)
	}
	if v := ds.GetString(TagRequestedProcedureID); v != "RP1001" {
		t.Errorf("RequestedProcedureID = %q", v)
	}
	// Attributes that were never requested must be absent.
	for _, tag := range []uint32{TagPatientSex, TagPatientBirthDate, TagStudyDescription, TagInstitutionName} {
		if _, ok := ds.Get(tag); ok {
			t.Errorf("attribute 0x%08X was returned although it was not requested", tag)
		}
	}

	sps, ok := ds.GetSequence(TagScheduledProcedureStepSeq)
	if !ok {
		t.Fatal("ScheduledProcedureStepSequence missing from the response")
	}
	if len(sps) != 7 {
		t.Fatalf("SPS item has %d elements, want the 7 requested: %+v", len(sps), sps)
	}
	if v := sps.GetString(TagModality); v != "CT" {
		t.Errorf("SPS Modality = %q", v)
	}
	if v := sps.GetString(TagScheduledStationAETitle); v != "CT1" {
		t.Errorf("SPS ScheduledStationAETitle = %q", v)
	}
	if v := sps.GetString(TagSPSStartDate); v != "20240115" {
		t.Errorf("SPS start date = %q", v)
	}
	if v := sps.GetString(TagScheduledPerformingPhysName); v != "WILSON^JAMES" {
		t.Errorf("SPS performing physician = %q", v)
	}
	if v := sps.GetString(TagSPSID); v != "SPS1001" {
		t.Errorf("SPS ID = %q", v)
	}
	if _, ok := sps.Get(TagRequestedContrastAgent); ok {
		t.Error("SPS contrast agent returned although it was not requested")
	}

	// The response identifier must encode and decode cleanly in both transfer
	// syntaxes, preserving values.
	for _, explicit := range []bool{true, false} {
		raw, err := EncodeDataSet(ds, explicit)
		if err != nil {
			t.Fatalf("explicit=%v EncodeDataSet: %v", explicit, err)
		}
		got, err := DecodeDataSet(raw, explicit)
		if err != nil {
			t.Fatalf("explicit=%v DecodeDataSet: %v", explicit, err)
		}
		if v := got.GetString(TagPatientName); v != "SMITH^JOHN" {
			t.Errorf("explicit=%v: decoded PatientName = %q", explicit, v)
		}
		inner, ok := got.GetSequence(TagScheduledProcedureStepSeq)
		if !ok {
			t.Fatalf("explicit=%v: decoded sequence missing", explicit)
		}
		if v := inner.GetString(TagSPSStartDate); v != "20240115" {
			t.Errorf("explicit=%v: decoded SPS start date = %q", explicit, v)
		}
		if v := inner.GetString(TagSPSID); v != "SPS1001" {
			t.Errorf("explicit=%v: decoded SPS ID = %q", explicit, v)
		}
	}
}

func TestBuildResponseIdentifierZeroLengthForUnsupportedAttributes(t *testing.T) {
	q, err := ParseMWLQuery(DataSet{
		NewStringElement(TagPatientName, "PN", ""),
		NewStringElement(TagRequestedProcedurePriority, "SH", ""),
		NewStringElement(TagStudyInstanceUID, "UI", ""),
	})
	if err != nil {
		t.Fatalf("ParseMWLQuery: %v", err)
	}
	ds, err := q.BuildResponseIdentifier(sampleItem())
	if err != nil {
		t.Fatalf("BuildResponseIdentifier: %v", err)
	}
	if len(ds) != 3 {
		t.Fatalf("elements = %d, want 3", len(ds))
	}
	e, ok := ds.Get(TagRequestedProcedurePriority)
	if !ok {
		t.Fatal("RequestedProcedurePriority not returned")
	}
	if len(e.Value) != 0 {
		t.Errorf("unsupported attribute returned %q, want zero length", e.String())
	}
	if e.VR != "SH" {
		t.Errorf("unsupported attribute VR = %q, want SH", e.VR)
	}
}

func TestBuildResponseIdentifierNilQuery(t *testing.T) {
	var q *MWLQuery
	if _, err := q.BuildResponseIdentifier(sampleItem()); err == nil {
		t.Fatal("BuildResponseIdentifier accepted a nil query")
	}
}

func TestBuildResponseIdentifierSequenceShape(t *testing.T) {
	// A query asking only for the SPS sequence must return only the sequence.
	q, err := ParseMWLQuery(DataSet{
		NewSequenceElement(TagScheduledProcedureStepSeq, DataSet{
			NewStringElement(TagModality, "CS", ""),
		}),
	})
	if err != nil {
		t.Fatalf("ParseMWLQuery: %v", err)
	}
	ds, err := q.BuildResponseIdentifier(sampleItem())
	if err != nil {
		t.Fatalf("BuildResponseIdentifier: %v", err)
	}
	if len(ds) != 1 || ds[0].Tag != TagScheduledProcedureStepSeq {
		t.Fatalf("response = %+v", ds)
	}
	if len(ds[0].Items) != 1 || len(ds[0].Items[0]) != 1 {
		t.Fatalf("SPS item = %+v", ds[0].Items)
	}
	if v := ds[0].Items[0][0].String(); v != "CT" {
		t.Errorf("Modality = %q", v)
	}
}
