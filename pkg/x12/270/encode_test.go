package x12_270

import (
	"strings"
	"testing"
)

func TestEncode270_BasicEligibilityInquiry(t *testing.T) {
	inq := &EligibilityInquiry270{
		SenderID:        "SENDER123",
		ReceiverID:      "RECEIVER456",
		TransactionDate: "20260809",
		GSDate:          "20260809",
		GSTime:          "1200",
		GroupControlNo:  "1",
		TransactionNo:   "0001",
		InfoSource: InfoSource{
			EntityID:     "PAYERID",
			EntityIDQual: "PI",
			Name:         "TEST PAYER",
		},
		InfoReceiver: InfoReceiver{
			EntityID:     "1234567890",
			EntityIDQual: "XX",
			Name:         "TEST PROVIDER",
		},
		Subscriber: SubscriberInfo{
			LastName:    "DOE",
			FirstName:   "JOHN",
			MiddleName:  "Q",
			IDQualifier: "MI",
			MemberID:    "MEMBER001",
			GroupNumber: "GROUP123",
			DOB:         "19800101",
			Gender:      "M",
			Address1:    "123 MAIN ST",
			City:        "ANYTOWN",
			State:       "NY",
			Zip:         "10001",
		},
		ServiceTypes: []string{"30"},
	}

	output, err := Encode270(inq)
	if err != nil {
		t.Fatalf("Encode270() error = %v", err)
	}

	outputStr := string(output)

	// Verify ISA segment exists and has correct structure
	if !strings.Contains(outputStr, "ISA*") {
		t.Error("missing ISA segment")
	}
	if !strings.Contains(outputStr, "SENDER123") {
		t.Error("ISA missing sender ID")
	}
	if !strings.Contains(outputStr, "RECEIVER456") {
		t.Error("ISA missing receiver ID")
	}

	// Verify GS segment
	if !strings.Contains(outputStr, "GS*HS*") {
		t.Error("missing GS segment with HS functional ID")
	}

	// Verify ST segment
	if !strings.Contains(outputStr, "ST*270*") {
		t.Error("missing ST segment with 270 transaction ID")
	}

	// Verify BHT segment
	if !strings.Contains(outputStr, "BHT*0022*13*") {
		t.Error("missing BHT segment")
	}

	// Verify HL segments for info source (2000A) and receiver (2000B)
	if !strings.Contains(outputStr, "HL*1**20*1") {
		t.Error("missing HL segment for Information Source (2000A)")
	}
	if !strings.Contains(outputStr, "HL*2*1*21*1") {
		t.Error("missing HL segment for Information Receiver (2000B)")
	}

	// Verify NM1 for payer (PR qualifier)
	if !strings.Contains(outputStr, "NM1*PR*2*TEST PAYER*****PI*PAYERID") {
		t.Error("missing or incorrect NM1 for payer (PR)")
	}

	// Verify NM1 for provider (1P qualifier)
	if !strings.Contains(outputStr, "NM1*1P*2*TEST PROVIDER*****XX*1234567890") {
		t.Error("missing or incorrect NM1 for provider (1P)")
	}

	// Verify REF for provider tax ID (TJ) since EntityIDQual is XX
	if !strings.Contains(outputStr, "REF*TJ*1234567890") {
		t.Error("missing REF*TJ for provider tax ID")
	}

	// Verify subscriber HL segment
	if !strings.Contains(outputStr, "HL*3*2*22*1") {
		t.Error("missing subscriber HL segment")
	}

	// Verify NM1 for subscriber (IL qualifier)
	if !strings.Contains(outputStr, "NM1*IL*1*DOE*JOHN*Q***MI*MEMBER001") {
		t.Error("missing or incorrect NM1 for subscriber")
	}

	// Verify REF for group number
	if !strings.Contains(outputStr, "REF*6P*GROUP123") {
		t.Error("missing REF for group number")
	}

	// Verify N3/N4 address
	if !strings.Contains(outputStr, "N3*123 MAIN ST") {
		t.Error("missing N3 address segment")
	}
	if !strings.Contains(outputStr, "N4*ANYTOWN*NY*10001") {
		t.Error("missing N4 city/state/zip segment")
	}

	// Verify DMG demographic
	if !strings.Contains(outputStr, "DMG*D8*19800101*M") {
		t.Error("missing or incorrect DMG segment")
	}

	// Verify EQ eligibility query
	if !strings.Contains(outputStr, "EQ*30") {
		t.Error("missing EQ segment with service type 30")
	}

	// Verify trailers
	if !strings.Contains(outputStr, "SE*") {
		t.Error("missing SE trailer")
	}
	if !strings.Contains(outputStr, "GE*1*1") {
		t.Error("missing GE trailer")
	}
	if !strings.Contains(outputStr, "IEA*1*") {
		t.Error("missing IEA trailer")
	}

	// Verify segment terminator is used consistently
	if !strings.Contains(outputStr, "~") {
		t.Error("missing segment terminators (~)")
	}

	// Verify no raw newlines in output
	if strings.Contains(outputStr, "\n") {
		t.Error("output should not contain newlines")
	}
}

func TestEncode270_WithDependent(t *testing.T) {
	inq := &EligibilityInquiry270{
		SenderID:        "SEND",
		ReceiverID:      "RECV",
		TransactionDate: "20260809",
		GSDate:          "20260809",
		GSTime:          "1200",
		GroupControlNo:  "1",
		TransactionNo:   "0001",
		InfoSource: InfoSource{
			EntityID:     "PAYER1",
			EntityIDQual: "PI",
			Name:         "HEALTH PLAN",
		},
		InfoReceiver: InfoReceiver{
			EntityID:     "9876543210",
			EntityIDQual: "FI",
			Name:         "DR SMITH",
		},
		Subscriber: SubscriberInfo{
			LastName:    "DOE",
			FirstName:   "JANE",
			IDQualifier: "MI",
			MemberID:    "MEM999",
			DOB:         "19850101",
			Gender:      "F",
		},
		Dependent: &DependentInfo{
			LastName:     "DOE",
			FirstName:    "JIMMY",
			DOB:          "20100101",
			Gender:       "M",
			Relationship: "19",
		},
		ServiceTypes: []string{"30", "48"},
	}

	output, err := Encode270(inq)
	if err != nil {
		t.Fatalf("Encode270() error = %v", err)
	}

	outputStr := string(output)

	// Dependent HL segment should have code 23
	if !strings.Contains(outputStr, "HL*4*3*23*0") {
		t.Error("missing dependent HL segment (2000D)")
	}

	// Dependent NM1 with QC qualifier
	if !strings.Contains(outputStr, "NM1*QC*1*DOE*JIMMY") {
		t.Error("missing or incorrect NM1 for dependent (QC)")
	}

	// INS relationship segment
	if !strings.Contains(outputStr, "INS*Y*18*19***A") {
		t.Error("missing or incorrect INS segment for dependent relationship")
	}

	// Dependent DMG
	if !strings.Contains(outputStr, "DMG*D8*20100101*M") {
		t.Error("missing or incorrect DMG segment for dependent")
	}

	// Subscriber HL should have HL04=0 (has dependents)
	if !strings.Contains(outputStr, "HL*3*2*22*0") {
		t.Error("subscriber HL should indicate has dependents (HL04=0)")
	}

	// EQ with multiple service types
	if !strings.Contains(outputStr, "EQ*30*48") {
		t.Error("missing EQ segment with multiple service types")
	}

	// Verify all 4 HL segments exist
	hlCount := strings.Count(outputStr, "HL*")
	if hlCount < 4 {
		t.Errorf("expected at least 4 HL segments, got %d", hlCount)
	}
}

func TestEncode270_MinimalInquiry(t *testing.T) {
	inq := &EligibilityInquiry270{
		SenderID:        "A",
		ReceiverID:      "B",
		TransactionDate: "20260809",
		GSDate:          "20260809",
		GSTime:          "0000",
		GroupControlNo:  "1",
		TransactionNo:   "0001",
		InfoSource: InfoSource{
			EntityID:     "X",
			EntityIDQual: "PI",
			Name:         "P",
		},
		InfoReceiver: InfoReceiver{
			EntityID:     "Y",
			EntityIDQual: "XX",
			Name:         "R",
		},
		Subscriber: SubscriberInfo{
			LastName:    "S",
			FirstName:   "F",
			IDQualifier: "MI",
			MemberID:    "M",
			DOB:         "19900101",
			Gender:      "U",
		},
		ServiceTypes: []string{"30"},
	}

	output, err := Encode270(inq)
	if err != nil {
		t.Fatalf("Encode270() error = %v", err)
	}

	outputStr := string(output)

	// Must start with ISA
	if !strings.HasPrefix(outputStr, "ISA*") {
		t.Error("output must start with ISA segment")
	}

	// Must end with IEA
	if !strings.Contains(outputStr, "IEA*1*") {
		t.Error("output must end with IEA trailer")
	}

	// padRight pads to exactly 15 chars, so "A" becomes "A" + 14 spaces
	expectedSender := "A" + strings.Repeat(" ", 14)
	if !strings.Contains(outputStr, expectedSender) {
		t.Errorf("ISA sender padded incorrectly; expected %q within output", expectedSender)
	}
	expectedReceiver := "B" + strings.Repeat(" ", 14)
	if !strings.Contains(outputStr, expectedReceiver) {
		t.Errorf("ISA receiver padded incorrectly; expected %q within output", expectedReceiver)
	}
}

func TestEncode270_NilDependent(t *testing.T) {
	inq := &EligibilityInquiry270{
		SenderID:        "SEND",
		ReceiverID:      "RECV",
		TransactionDate: "20260809",
		GSDate:          "20260809",
		GSTime:          "1200",
		GroupControlNo:  "1",
		TransactionNo:   "0001",
		InfoSource: InfoSource{
			EntityID:     "PAY",
			EntityIDQual: "PI",
			Name:         "PAYER",
		},
		InfoReceiver: InfoReceiver{
			EntityID:     "123",
			EntityIDQual: "FI",
			Name:         "DOC",
		},
		Subscriber: SubscriberInfo{
			LastName:    "SMITH",
			FirstName:   "BOB",
			IDQualifier: "MI",
			MemberID:    "M123",
			DOB:         "19700101",
			Gender:      "M",
		},
		ServiceTypes: []string{"30"},
	}

	output, err := Encode270(inq)
	if err != nil {
		t.Fatalf("Encode270() error = %v", err)
	}

	outputStr := string(output)

	// When no dependent, there should be exactly 3 HL segments
	hlCount := strings.Count(outputStr, "HL*")
	if hlCount != 3 {
		t.Errorf("expected 3 HL segments (no dependent), got %d", hlCount)
	}

	// No QC (dependent) NM1 should appear
	if strings.Contains(outputStr, "NM1*QC") {
		t.Error("should not have dependent NM1 when Dependent is nil")
	}

	// No REF*TJ when EntityIDQual is not XX
	if strings.Contains(outputStr, "REF*TJ") {
		t.Error("should not have REF*TJ when EntityIDQual is FI, not XX")
	}
}
