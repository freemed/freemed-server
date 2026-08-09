package x12_270

import (
	"testing"
)

// sample271Response is a realistic X12 271 eligibility response with
// payer info, subscriber eligibility, and benefit details.
// NM1 segments use 9 elements to pass the parser's length check.
// EB07 amounts use element index 6 (the parser's expected position).
const sample271Response = `ISA*00*          *00*          *ZZ*SENDER         *ZZ*RECEIVER       *240809*1200*U*00401*000000001*0*T*:~
GS*HB*SENDER*RECEIVER*20240809*1200*1*X*004010X092A1~
ST*271*0001~
BHT*0022*11*TRACE001*20240809*1200~
HL*1**20*1~
NM1*PR*2*HEALTH PLAN INC*****PI*PAYERID123~
HL*2*1*21*1~
NM1*1P*2*MEDICAL PRACTICE*****XX*1234567890~
HL*3*2*22*1~
NM1*IL*1*DOE*JOHN****MI*MEMBER001~
REF*6P*GROUP123~
N3*123 MAIN ST~
N4*ANYTOWN*NY*10001~
DMG*D8*19800101*M~
DTP*291*D8*20240101-20241231~
TRN*1*TRACE001~
EB*1**30*Active Coverage~
EB*A**30*Co-Insurance***20.00~
EB*B**30*Co-Payment***25.00~
EB*C**30*Deductible***500.00~
EB*1**48*Hospital Inpatient~
EB*F**30*Limitations~
SE*20*0001~
GE*1*1~
IEA*1*000000001~
`

func TestParse271_BasicResponse(t *testing.T) {
	resp, err := Parse271([]byte(sample271Response))
	if err != nil {
		t.Fatalf("Parse271() error = %v", err)
	}

	if resp == nil {
		t.Fatal("Parse271() returned nil response")
	}

	// Verify payer info
	if resp.PayerName != "HEALTH PLAN INC" {
		t.Errorf("PayerName = %q, want %q", resp.PayerName, "HEALTH PLAN INC")
	}
	if resp.PayerID != "PAYERID123" {
		t.Errorf("PayerID = %q, want %q", resp.PayerID, "PAYERID123")
	}

	// Verify trace number (from TRN, not BHT)
	if resp.TraceNumber != "TRACE001" {
		t.Errorf("TraceNumber = %q, want %q", resp.TraceNumber, "TRACE001")
	}

	// Verify subscriber
	if resp.Subscriber.LastName != "DOE" {
		t.Errorf("Subscriber.LastName = %q, want %q", resp.Subscriber.LastName, "DOE")
	}
	if resp.Subscriber.FirstName != "JOHN" {
		t.Errorf("Subscriber.FirstName = %q, want %q", resp.Subscriber.FirstName, "JOHN")
	}
	if resp.Subscriber.MemberID != "MEMBER001" {
		t.Errorf("Subscriber.MemberID = %q, want %q", resp.Subscriber.MemberID, "MEMBER001")
	}

	// Verify eligibility date from DTP
	if resp.Subscriber.EligibilityDate != "20240101-20241231" {
		t.Errorf("Subscriber.EligibilityDate = %q, want %q", resp.Subscriber.EligibilityDate, "20240101-20241231")
	}

	// Verify no dependent
	if resp.Dependent != nil {
		t.Error("expected nil Dependent for subscriber-only response")
	}

	// Verify benefits count
	if len(resp.Benefits) != 6 {
		t.Errorf("Benefits length = %d, want 6", len(resp.Benefits))
	}
}

func TestParse271_Benefits(t *testing.T) {
	resp, err := Parse271([]byte(sample271Response))
	if err != nil {
		t.Fatalf("Parse271() error = %v", err)
	}

	if len(resp.Benefits) == 0 {
		t.Fatal("expected benefits in response")
	}

	// First benefit: Active coverage
	b0 := resp.Benefits[0]
	if b0.CoverageLevel != "1" {
		t.Errorf("Benefit[0].CoverageLevel = %q, want %q", b0.CoverageLevel, "1")
	}
	if b0.ServiceTypeCode != "30" {
		t.Errorf("Benefit[0].ServiceTypeCode = %q, want %q", b0.ServiceTypeCode, "30")
	}
	if b0.ServiceTypeDesc != "Health Benefit Plan Coverage" {
		t.Errorf("Benefit[0].ServiceTypeDesc = %q, want %q", b0.ServiceTypeDesc, "Health Benefit Plan Coverage")
	}

	// Co-Insurance benefit (amount at element[6] via ***20.00)
	b1 := resp.Benefits[1]
	if b1.CoverageLevel != "A" {
		t.Errorf("Benefit[1].CoverageLevel = %q, want %q", b1.CoverageLevel, "A")
	}
	if b1.Amount != 20.00 {
		t.Errorf("Benefit[1].Amount = %f, want 20.00", b1.Amount)
	}

	// Co-Payment benefit
	b2 := resp.Benefits[2]
	if b2.CoverageLevel != "B" {
		t.Errorf("Benefit[2].CoverageLevel = %q, want %q", b2.CoverageLevel, "B")
	}
	if b2.Amount != 25.00 {
		t.Errorf("Benefit[2].Amount = %f, want 25.00", b2.Amount)
	}

	// Deductible benefit
	b3 := resp.Benefits[3]
	if b3.CoverageLevel != "C" {
		t.Errorf("Benefit[3].CoverageLevel = %q, want %q", b3.CoverageLevel, "C")
	}
	if b3.Amount != 500.00 {
		t.Errorf("Benefit[3].Amount = %f, want 500.00", b3.Amount)
	}

	// Hospital benefit
	b4 := resp.Benefits[4]
	if b4.ServiceTypeCode != "48" {
		t.Errorf("Benefit[4].ServiceTypeCode = %q, want %q", b4.ServiceTypeCode, "48")
	}
	if b4.ServiceTypeDesc != "Hospital — Inpatient" {
		t.Errorf("Benefit[4].ServiceTypeDesc = %q, want %q", b4.ServiceTypeDesc, "Hospital — Inpatient")
	}
}

func TestParse271_EmptyInput(t *testing.T) {
	resp, err := Parse271([]byte{})
	if err != nil {
		t.Fatalf("Parse271() error = %v", err)
	}
	if resp == nil {
		t.Fatal("Parse271() returned nil for empty input")
	}
	// Should return empty response, not nil
	if len(resp.Benefits) != 0 {
		t.Error("expected empty benefits for empty input")
	}
}

func TestParse271_ServiceTypeDescriptions(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"1", "Medical Care"},
		{"2", "Surgical"},
		{"30", "Health Benefit Plan Coverage"},
		{"33", "Chiropractic"},
		{"35", "Dental Care"},
		{"48", "Hospital — Inpatient"},
		{"50", "Hospital — Outpatient"},
		{"88", "Pharmacy"},
		{"AL", "Vision (Optometry)"},
		{"MH", "Mental Health"},
		{"UC", "Urgent Care"},
		{"999", "Service Type 999"},
	}

	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			got := serviceTypeDescription(tc.code)
			if got != tc.want {
				t.Errorf("serviceTypeDescription(%q) = %q, want %q", tc.code, got, tc.want)
			}
		})
	}
}

func TestParse271_WithDependent(t *testing.T) {
	// NM1 segments include all 9 fields to pass parser length check.
	depResponse := `ISA*00*          *00*          *ZZ*SEND           *ZZ*RECV           *240809*1200*U*00401*000000001*0*T*:~
GS*HB*SEND*RECV*20240809*1200*1*X*004010X092A1~
ST*271*0001~
BHT*0022*11*DEPTRACE*20240809*1200~
HL*1**20*1~
NM1*PR*2*HEALTH PLAN*****PI*P999~
HL*2*1*21*1~
NM1*1P*2*PROVIDER*****XX*1111111111~
HL*3*2*22*0~
NM1*IL*1*DOE*JANE****MI*MEM001~
DMG*D8*19850101*F~
DTP*291*D8*20240101-20241231~
EB*1**30*Active Coverage~
HL*4*3*23*0~
NM1*QC*1*DOE*JIMMY****MI*~
INS*Y*18*19***A~
DMG*D8*20100101*M~
DTP*291*D8*20240101-20241231~
EB*1**30*Dependent Coverage~
SE*18*0001~
GE*1*1~
IEA*1*000000001~
`

	resp, err := Parse271([]byte(depResponse))
	if err != nil {
		t.Fatalf("Parse271() error = %v", err)
	}

	// Verify subscriber
	if resp.Subscriber.LastName != "DOE" {
		t.Errorf("Subscriber.LastName = %q, want %q", resp.Subscriber.LastName, "DOE")
	}
	if resp.Subscriber.FirstName != "JANE" {
		t.Errorf("Subscriber.FirstName = %q, want %q", resp.Subscriber.FirstName, "JANE")
	}
	if resp.Subscriber.MemberID != "MEM001" {
		t.Errorf("Subscriber.MemberID = %q, want %q", resp.Subscriber.MemberID, "MEM001")
	}

	// Verify dependent
	if resp.Dependent == nil {
		t.Fatal("expected dependent in response")
	}
	if resp.Dependent.LastName != "DOE" {
		t.Errorf("Dependent.LastName = %q, want %q", resp.Dependent.LastName, "DOE")
	}
	if resp.Dependent.FirstName != "JIMMY" {
		t.Errorf("Dependent.FirstName = %q, want %q", resp.Dependent.FirstName, "JIMMY")
	}
	if resp.Dependent.EligibilityDate != "20240101-20241231" {
		t.Errorf("Dependent.EligibilityDate = %q, want %q", resp.Dependent.EligibilityDate, "20240101-20241231")
	}
}

func TestParse271_Not271Transaction(t *testing.T) {
	// An 837 transaction — the parser should still process, just not match ST check
	not271 := `ISA*00*          *00*          *ZZ*SEND           *ZZ*RECV           *240809*1200*U*00401*000000001*0*T*:~
GS*HC*SEND*RECV*20240809*1200*1*X*004010X098A1~
ST*837*0001~
SE*3*0001~
GE*1*1~
IEA*1*000000001~
`

	resp, err := Parse271([]byte(not271))
	if err != nil {
		t.Fatalf("Parse271() error = %v", err)
	}
	if resp == nil {
		t.Fatal("Parse271() should not return nil for non-271 input")
	}
	// BHT will be parsed but there's no NM1*PR so payer info is empty
	if resp.PayerName != "" {
		t.Error("expected empty payer name for non-271 input")
	}
}
