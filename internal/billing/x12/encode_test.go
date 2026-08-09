package x12

import (
	"strings"
	"testing"
)

// TestSegmentString verifies basic Segment serialization.
func TestSegmentString(t *testing.T) {
	s := NewSegment("NM1", "85", "2", "FREEMED CLINIC", "", "", "", "", "", "XX", "1234567890")
	got := s.String()
	want := "NM1*85*2*FREEMED CLINIC******XX*1234567890~"
	if got != want {
		t.Errorf("Segment.String() = %q, want %q", got, want)
	}
}

// TestFormatDollars verifies monetary formatting.
func TestFormatDollars(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{100.00, "10000"},
		{0.00, "0"},
		{0.50, "50"},
		{1.25, "125"},
		{1500.99, "150099"},
		{0.01, "1"},
		{12345.67, "1234567"},
	}
	for _, tt := range tests {
		got := formatDollars(tt.input)
		if got != tt.want {
			t.Errorf("formatDollars(%f) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// TestBuildCLM verifies CLM segment construction.
func TestBuildCLM(t *testing.T) {
	seg := BuildCLM("CLAIM123", 150.00, "11", "A", "B", "Y")
	got := seg.String()
	// CLM*CLAIM123*15000***11:A:B:Y*Y*A*Y*Y*P~
	want := "CLM*CLAIM123*15000***11:A:B:Y*Y*A*Y*Y*P~"
	if got != want {
		t.Errorf("BuildCLM() = %q, want %q", got, want)
	}
}

// TestBuildSV1 verifies SV1 segment construction.
func TestBuildSV1(t *testing.T) {
	seg := BuildSV1("99213", []string{"25"}, 95.00, 1, []int{1, 2})
	got := seg.String()
	// SV1*HC:99213:25*9500*UN*1***1:2~
	want := "SV1*HC:99213:25*9500*UN*1***1:2~"
	if got != want {
		t.Errorf("BuildSV1() = %q, want %q", got, want)
	}
}

// TestBuildSV1NoModifiers verifies SV1 without modifiers.
func TestBuildSV1NoModifiers(t *testing.T) {
	seg := BuildSV1("99214", nil, 120.00, 1, []int{1})
	got := seg.String()
	// SV1*HC:99214*12000*UN*1***1~
	want := "SV1*HC:99214*12000*UN*1***1~"
	if got != want {
		t.Errorf("BuildSV1() = %q, want %q", got, want)
	}
}

// TestBuildSV1MultipleModifiers verifies SV1 with all modifiers.
func TestBuildSV1MultipleModifiers(t *testing.T) {
	seg := BuildSV1("97110", []string{"GP", "59", "LT"}, 75.00, 2, []int{1})
	got := seg.String()
	// SV1*HC:97110:GP:59:LT*7500*UN*2***1~
	want := "SV1*HC:97110:GP:59:LT*7500*UN*2***1~"
	if got != want {
		t.Errorf("BuildSV1() = %q, want %q", got, want)
	}
}

// TestBuildHI verifies HI segment construction for diagnoses.
func TestBuildHI(t *testing.T) {
	seg := BuildHI([]string{"J45.909", "J30.9"})
	got := seg.String()
	want := "HI*BK:J45.909*BF:J30.9~"
	if got != want {
		t.Errorf("BuildHI() = %q, want %q", got, want)
	}
}

// TestBuildDTP verifies DTP segment construction.
func TestBuildDTP(t *testing.T) {
	seg := BuildDTP("472", "D8", "20240101")
	got := seg.String()
	want := "DTP*472*D8*20240101~"
	if got != want {
		t.Errorf("BuildDTP() = %q, want %q", got, want)
	}
}

// TestEncode837Professional verifies a complete minimal 837P encoding.
func TestEncode837Professional(t *testing.T) {
	claim := &Claim837P{
		SubmitterName: "FREEMED EMR",
		SubmitterID:   "SUBMIT123",
		ReceiverName:  "TEST PAYER",
		ReceiverID:    "PAYER456",
		BillingProvider: ProviderInfo{
			Name:    "FREEMED CLINIC",
			NPI:     "1234567890",
			TaxID:   "123456789",
			Address1: "123 MAIN ST",
			City:    "HARTFORD",
			State:   "CT",
			Zip:     "06101",
			Phone:   "8605551212",
		},
		Subscriber: SubscriberInfo{
			LastName:  "DOE",
			FirstName: "JOHN",
			MemberID:  "MEMBER001",
			SSN:       "123456789",
			DOB:       "19800101",
			PayerName: "TEST PAYER",
			PayerID:   "PAYER456",
		},
		Patient: PatientInfo{
			LastName:                 "DOE",
			FirstName:                "JOHN",
			DOB:                      "19800101",
			RelationshipToSubscriber: "18", // self
			SSN:                      "123456789",
		},
		Claim: ClaimInfo{
			ClaimID:        "CLAIM001",
			TotalCharges:   245.00,
			PlaceOfService: "11",
			DiagnosisCodes: []string{"J45.909", "J30.9"},
			StatementFrom:  "20240101",
			StatementTo:    "20240101",
		},
		ServiceLines: []ServiceLineInfo{
			{
				LineNumber:        1,
				ProcedureCode:     "99213",
				Modifier1:         "25",
				Charge:            150.00,
				Units:             1,
				DiagnosisPointers: []int{1, 2},
				ServiceDate:       "20240101",
			},
			{
				LineNumber:        2,
				ProcedureCode:     "95004",
				Charge:            95.00,
				Units:             1,
				DiagnosisPointers: []int{1},
				ServiceDate:       "20240101",
			},
		},
		RenderingProvider: &ProviderRefInfo{
			LastName:  "SMITH",
			FirstName: "JANE",
			NPI:       "9876543210",
			Taxonomy:  "207Q00000X",
		},
	}

	output, err := Encode837Professional(claim)
	if err != nil {
		t.Fatalf("Encode837Professional() error: %v", err)
	}

	got := string(output)

	// Verify required envelope segments are present
	requiredSegments := []string{
		"ISA*",
		"GS*HC*",
		"ST*837*",
		"BHT*0019*",
		"NM1*41*", // submitter
		"NM1*40*", // receiver
		"HL*1**20*",  // billing provider HL
		"PRV*BI*",
		"NM1*85*", // billing provider name
		"HL*2*1*22*", // subscriber HL
		"SBR*P*",
		"NM1*IL*", // subscriber name
		"NM1*PR*", // payer name
		"CLM*CLAIM001*",
		"HI*BK:",
		"NM1*82*", // rendering provider
		"LX*1~",
		"SV1*HC:99213:25",
		"DTP*472*",
		"LX*2~",
		"SV1*HC:95004",
		"SE*",
		"GE*",
		"IEA*",
	}

	for _, seg := range requiredSegments {
		if !strings.Contains(got, seg) {
			t.Errorf("missing expected segment prefix %q in output:\n%s", seg, got)
		}
	}

	// Verify no empty tags (Segment{} should not appear in output)
	if strings.Contains(got, "Segment{}") {
		t.Error("output contains empty Segment{} placeholder")
	}

	// Verify segment terminators are present
	if !strings.Contains(got, "~") {
		t.Error("output missing segment terminator '~'")
	}
}

// TestEncode837ProfessionalWithPatient verifies encoding when patient ≠ subscriber.
func TestEncode837ProfessionalWithPatient(t *testing.T) {
	claim := &Claim837P{
		SubmitterName: "FREEMED EMR",
		SubmitterID:   "SUBMIT123",
		ReceiverName:  "TEST PAYER",
		ReceiverID:    "PAYER456",
		BillingProvider: ProviderInfo{
			Name:    "FREEMED CLINIC",
			NPI:     "1234567890",
			TaxID:   "123456789",
			Address1: "123 MAIN ST",
			City:    "HARTFORD",
			State:   "CT",
			Zip:     "06101",
		},
		Subscriber: SubscriberInfo{
			LastName:  "DOE",
			FirstName: "JOHN",
			MemberID:  "MEMBER001",
			PayerName: "TEST PAYER",
			PayerID:   "PAYER456",
		},
		Patient: PatientInfo{
			LastName:                 "DOE",
			FirstName:                "JANE",
			DOB:                      "20150101",
			RelationshipToSubscriber: "19", // child
			SSN:                      "987654321",
		},
		Claim: ClaimInfo{
			ClaimID:        "CLAIM002",
			TotalCharges:   150.00,
			PlaceOfService: "11",
			DiagnosisCodes: []string{"J45.909"},
			StatementFrom:  "20240101",
			StatementTo:    "20240101",
		},
		ServiceLines: []ServiceLineInfo{
			{
				LineNumber:        1,
				ProcedureCode:     "99213",
				Charge:            150.00,
				Units:             1,
				DiagnosisPointers: []int{1},
				ServiceDate:       "20240101",
			},
		},
	}

	output, err := Encode837Professional(claim)
	if err != nil {
		t.Fatalf("Encode837Professional() error: %v", err)
	}

	got := string(output)

	// Should contain patient HL (2000C)
	if !strings.Contains(got, "HL*3*2*23*") {
		t.Error("missing patient HL segment (2000C)")
	}
	if !strings.Contains(got, "NM1*QC*") {
		t.Error("missing patient NM1 segment (2010CA)")
	}
	if !strings.Contains(got, "JANE") {
		t.Error("missing patient first name")
	}

	// Subscriber HL04 should be "1" (has children)
	if !strings.Contains(got, "HL*2*1*22*1~") {
		t.Error("subscriber HL04 should be '1' when patient ≠ subscriber")
	}
}

// TestValidateClaim tests the validation logic.
func TestValidateClaim(t *testing.T) {
	tests := []struct {
		name    string
		claim   *Claim837P
		wantErr bool
	}{
		{
			name:    "nil claim",
			claim:   nil,
			wantErr: true,
		},
		{
			name: "minimal valid claim",
			claim: &Claim837P{
				SubmitterID: "SUB1",
				ReceiverID:  "REC1",
				BillingProvider: ProviderInfo{
					NPI: "1234567890",
				},
				Subscriber: SubscriberInfo{
					MemberID: "MEM1",
				},
				Patient: PatientInfo{
					RelationshipToSubscriber: "18",
				},
				Claim: ClaimInfo{
					ClaimID:        "CLM1",
					PlaceOfService: "11",
					StatementFrom:  "20240101",
					StatementTo:    "20240101",
				},
				ServiceLines: []ServiceLineInfo{
					{
						ProcedureCode:     "99213",
						Charge:            100.00,
						Units:             1,
						DiagnosisPointers: []int{1},
						ServiceDate:       "20240101",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing submitter ID",
			claim: &Claim837P{
				ReceiverID: "REC1",
				BillingProvider: ProviderInfo{
					NPI: "1234567890",
				},
				Subscriber: SubscriberInfo{
					MemberID: "MEM1",
				},
				Patient: PatientInfo{
					RelationshipToSubscriber: "18",
				},
				Claim: ClaimInfo{
					ClaimID:        "CLM1",
					PlaceOfService: "11",
					StatementFrom:  "20240101",
					StatementTo:    "20240101",
				},
				ServiceLines: []ServiceLineInfo{
					{
						ProcedureCode:     "99213",
						Charge:            100.00,
						Units:             1,
						ServiceDate:       "20240101",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid date format",
			claim: &Claim837P{
				SubmitterID: "SUB1",
				ReceiverID:  "REC1",
				BillingProvider: ProviderInfo{
					NPI: "1234567890",
				},
				Subscriber: SubscriberInfo{
					MemberID: "MEM1",
				},
				Patient: PatientInfo{
					RelationshipToSubscriber: "18",
				},
				Claim: ClaimInfo{
					ClaimID:        "CLM1",
					PlaceOfService: "11",
					StatementFrom:  "01/01/2024", // invalid format
					StatementTo:    "20240101",
				},
				ServiceLines: []ServiceLineInfo{
					{
						ProcedureCode:     "99213",
						Charge:            100.00,
						Units:             1,
						ServiceDate:       "20240101",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "diagnosis pointer out of range",
			claim: &Claim837P{
				SubmitterID: "SUB1",
				ReceiverID:  "REC1",
				BillingProvider: ProviderInfo{
					NPI: "1234567890",
				},
				Subscriber: SubscriberInfo{
					MemberID: "MEM1",
				},
				Patient: PatientInfo{
					RelationshipToSubscriber: "18",
				},
				Claim: ClaimInfo{
					ClaimID:        "CLM1",
					PlaceOfService: "11",
					DiagnosisCodes: []string{"J45.909"}, // only 1 diagnosis
					StatementFrom:  "20240101",
					StatementTo:    "20240101",
				},
				ServiceLines: []ServiceLineInfo{
					{
						ProcedureCode:     "99213",
						Charge:            100.00,
						Units:             1,
						DiagnosisPointers: []int{3}, // out of range
						ServiceDate:       "20240101",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateClaim(tt.claim)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateClaim() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

// TestEncodeWithMultipleServiceLines verifies encoding of a claim with
// three service lines, each with different modifiers and dates.
func TestEncodeWithMultipleServiceLines(t *testing.T) {
	claim := &Claim837P{
		SubmitterName: "FREEMED EMR",
		SubmitterID:   "SUBMIT123",
		ReceiverName:  "TEST PAYER",
		ReceiverID:    "PAYER456",
		BillingProvider: ProviderInfo{
			Name:    "FREEMED CLINIC",
			NPI:     "1234567890",
			TaxID:   "123456789",
			Address1: "123 MAIN ST",
			City:    "HARTFORD",
			State:   "CT",
			Zip:     "06101",
		},
		Subscriber: SubscriberInfo{
			LastName:  "DOE",
			FirstName: "JOHN",
			MemberID:  "MEMBER001",
			PayerName: "TEST PAYER",
			PayerID:   "PAYER456",
		},
		Patient: PatientInfo{
			LastName:                 "DOE",
			FirstName:                "JOHN",
			RelationshipToSubscriber: "18",
		},
		Claim: ClaimInfo{
			ClaimID:        "CLAIM003",
			TotalCharges:   500.00,
			PlaceOfService: "11",
			DiagnosisCodes: []string{"J45.909", "J30.9", "E11.9"},
			StatementFrom:  "20240201",
			StatementTo:    "20240201",
		},
		ServiceLines: []ServiceLineInfo{
			{LineNumber: 1, ProcedureCode: "99214", Modifier1: "25", Charge: 200.00, Units: 1, DiagnosisPointers: []int{1, 2}, ServiceDate: "20240201"},
			{LineNumber: 2, ProcedureCode: "97110", Modifier1: "GP", Charge: 150.00, Units: 2, DiagnosisPointers: []int{1}, ServiceDate: "20240201"},
			{LineNumber: 3, ProcedureCode: "85025", Charge: 150.00, Units: 1, DiagnosisPointers: []int{3}, ServiceDate: "20240201"},
		},
	}

	output, err := Encode837Professional(claim)
	if err != nil {
		t.Fatalf("Encode837Professional error: %v", err)
	}

	got := string(output)

	// Verify all three LX segments
	if !strings.Contains(got, "LX*1~") {
		t.Error("missing LX segment for line 1")
	}
	if !strings.Contains(got, "LX*2~") {
		t.Error("missing LX segment for line 2")
	}
	if !strings.Contains(got, "LX*3~") {
		t.Error("missing LX segment for line 3")
	}

	// Verify all three SV1 segments
	if !strings.Contains(got, "SV1*HC:99214:25") {
		t.Error("missing SV1 for procedure 99214")
	}
	if !strings.Contains(got, "SV1*HC:97110:GP") {
		t.Error("missing SV1 for procedure 97110")
	}
	if !strings.Contains(got, "SV1*HC:85025") {
		t.Error("missing SV1 for procedure 85025")
	}

	// Verify DTP segments for each service line
	dtpCount := 0
	for i := 0; i < len(got)-6; i++ {
		if got[i:i+6] == "DTP*47" {
			dtpCount++
		}
	}
	if dtpCount < 3 {
		t.Errorf("expected at least 3 DTP segments, got %d", dtpCount)
	}
}

// TestEncodeWithAllModifiers verifies SV1 encoding with all four modifier slots.
func TestEncodeWithAllModifiers(t *testing.T) {
	claim := &Claim837P{
		SubmitterName: "FREEMED EMR",
		SubmitterID:   "SUBMIT123",
		ReceiverName:  "TEST PAYER",
		ReceiverID:    "PAYER456",
		BillingProvider: ProviderInfo{
			Name:    "FREEMED CLINIC",
			NPI:     "1234567890",
			TaxID:   "123456789",
			Address1: "123 MAIN ST",
			City:    "HARTFORD",
			State:   "CT",
			Zip:     "06101",
		},
		Subscriber: SubscriberInfo{
			LastName:  "DOE",
			FirstName: "JOHN",
			MemberID:  "MEMBER001",
			PayerName: "TEST PAYER",
			PayerID:   "PAYER456",
		},
		Patient: PatientInfo{
			LastName:                 "DOE",
			FirstName:                "JOHN",
			RelationshipToSubscriber: "18",
		},
		Claim: ClaimInfo{
			ClaimID:        "CLAIM004",
			TotalCharges:   350.00,
			PlaceOfService: "11",
			DiagnosisCodes: []string{"J45.909"},
			StatementFrom:  "20240301",
			StatementTo:    "20240301",
		},
		ServiceLines: []ServiceLineInfo{
			{
				LineNumber:        1,
				ProcedureCode:     "99215",
				Modifier1:         "25",
				Modifier2:         "59",
				Modifier3:         "LT",
				Modifier4:         "XE",
				Charge:            350.00,
				Units:             1,
				DiagnosisPointers: []int{1},
				ServiceDate:       "20240301",
			},
		},
	}

	output, err := Encode837Professional(claim)
	if err != nil {
		t.Fatalf("Encode837Professional error: %v", err)
	}

	got := string(output)

	// SV1 with all four modifiers
	if !strings.Contains(got, "SV1*HC:99215:25:59:LT:XE") {
		t.Errorf("expected SV1 with all modifiers, got: %s", got)
	}
}

// TestEncodeWithReferringProvider verifies encoding includes referring provider
// in loop 2310A.
func TestEncodeWithReferringProvider(t *testing.T) {
	claim := &Claim837P{
		SubmitterName: "FREEMED EMR",
		SubmitterID:   "SUBMIT123",
		ReceiverName:  "TEST PAYER",
		ReceiverID:    "PAYER456",
		BillingProvider: ProviderInfo{
			Name:    "FREEMED CLINIC",
			NPI:     "1234567890",
			TaxID:   "123456789",
			Address1: "123 MAIN ST",
			City:    "HARTFORD",
			State:   "CT",
			Zip:     "06101",
		},
		Subscriber: SubscriberInfo{
			LastName:  "DOE",
			FirstName: "JOHN",
			MemberID:  "MEMBER001",
			PayerName: "TEST PAYER",
			PayerID:   "PAYER456",
		},
		Patient: PatientInfo{
			LastName:                 "DOE",
			FirstName:                "JOHN",
			RelationshipToSubscriber: "18",
		},
		ReferringProvider: &ProviderRefInfo{
			LastName:  "WILSON",
			FirstName: "ROBERT",
			NPI:       "1112223334",
			Taxonomy:  "207Q00000X",
		},
		Claim: ClaimInfo{
			ClaimID:        "CLAIM005",
			TotalCharges:   200.00,
			PlaceOfService: "11",
			DiagnosisCodes: []string{"J45.909"},
			StatementFrom:  "20240401",
			StatementTo:    "20240401",
		},
		ServiceLines: []ServiceLineInfo{
			{LineNumber: 1, ProcedureCode: "99213", Charge: 200.00, Units: 1, DiagnosisPointers: []int{1}, ServiceDate: "20240401"},
		},
	}

	output, err := Encode837Professional(claim)
	if err != nil {
		t.Fatalf("Encode837Professional error: %v", err)
	}

	got := string(output)

	// Verify referring provider NM1 segment (DN qualifier)
	if !strings.Contains(got, "NM1*DN*1*WILSON*ROBERT****XX*1112223334") {
		t.Error("missing referring provider NM1 segment")
	}

	// Verify referring provider REF
	if !strings.Contains(got, "REF*1C*1112223334") {
		t.Error("missing referring provider REF segment")
	}
}

// TestPadOrTrunc verifies the padding helper.
func TestPadOrTrunc(t *testing.T) {
	tests := []struct {
		input  string
		length int
		want   string
	}{
		{"ABC", 5, "ABC  "},
		{"ABCDEF", 3, "ABC"},
		{"X", 15, "X              "}, // 1 + 14 spaces = 15
	}
	for _, tt := range tests {
		got := padOrTrunc(tt.input, tt.length)
		if got != tt.want {
			t.Errorf("padOrTrunc(%q, %d) = %q, want %q", tt.input, tt.length, got, tt.want)
		}
	}
}
