package x12_835

import (
	"testing"
)

const minimal835 = `ISA*00*          *00*          *ZZ*SUBMITTER      *ZZ*RECEIVER       *250101*1200*^*00501*000000001*0*P*:~
GS*HP*SUBMITTER*RECEIVER*20250101*1200*1*X*005010X221A1~
ST*835*0001~
BPR*I*1250.00*C*CHK**20250101*CHK123456*PAYERID01**PAYER NAME INC~
TRN*1*CHK123456*PAYERID01~
DTM*405*20250101~
N1*PR*PAYER NAME INC~
N3*123 PAYER ST~
N4*HARTFORD*CT*06101~
REF*EV*PAYERID01~
LX*1~
CLP*VC001*1*500.00*450.00*50.00*12*MC-2025-001~
CAS*CO*45*50.00~
NM1*QC*1*DOE*JOHN****MI*MEMBER001~
NM1*82*1*SMITH*JANE****XX*1234567890~
SVC*HC:99213*500.00*450.00~
CAS*PR*3*50.00~
DTM*472*20250101~
CLP*VC002*1*300.00*300.00*0.00*12*MC-2025-002~
NM1*QC*1*DOE*JANE****MI*MEMBER002~
SVC*HC:99214*300.00*300.00~
DTM*472*20250101~
SE*28*0001~
GE*1*1~
IEA*1*000000001~`

func TestParse835(t *testing.T) {
	era, err := Parse835([]byte(minimal835))
	if err != nil {
		t.Fatalf("Parse835 error: %v", err)
	}

	if era.PayerName != "PAYER NAME INC" {
		t.Errorf("expected PayerName 'PAYER NAME INC', got %q", era.PayerName)
	}
	if era.PayerID != "PAYERID01" {
		t.Errorf("expected PayerID 'PAYERID01', got %q", era.PayerID)
	}
	if era.PaymentAmount != 1250.0 {
		t.Errorf("expected PaymentAmount 1250.0, got %f", era.PaymentAmount)
	}
	if era.PaymentMethod != "CHK" {
		t.Errorf("expected PaymentMethod 'CHK', got %q", era.PaymentMethod)
	}
	if era.TraceNumber != "CHK123456" {
		t.Errorf("expected TraceNumber 'CHK123456', got %q", era.TraceNumber)
	}
	if era.PaymentDate != "20250101" {
		t.Errorf("expected PaymentDate '20250101', got %q", era.PaymentDate)
	}

	if len(era.Claims) != 2 {
		t.Fatalf("expected 2 claims, got %d", len(era.Claims))
	}

	// First claim
	c1 := era.Claims[0]
	if c1.PatientControlNumber != "VC001" {
		t.Errorf("claim 1: expected PatientControlNumber 'VC001', got %q", c1.PatientControlNumber)
	}
	if c1.ClaimStatusCode != "1" {
		t.Errorf("claim 1: expected ClaimStatusCode '1', got %q", c1.ClaimStatusCode)
	}
	if c1.ClaimCharge != 500.0 {
		t.Errorf("claim 1: expected ClaimCharge 500.0, got %f", c1.ClaimCharge)
	}
	if c1.ClaimPayment != 450.0 {
		t.Errorf("claim 1: expected ClaimPayment 450.0, got %f", c1.ClaimPayment)
	}
	if c1.PatientResponsibility != 50.0 {
		t.Errorf("claim 1: expected PatientResponsibility 50.0, got %f", c1.PatientResponsibility)
	}
	if c1.PayerClaimID != "MC-2025-001" {
		t.Errorf("claim 1: expected PayerClaimID 'MC-2025-001', got %q", c1.PayerClaimID)
	}
	if len(c1.CASAdjustments) != 1 {
		t.Errorf("claim 1: expected 1 CAS adjustment (claim-level), got %d", len(c1.CASAdjustments))
	}
	if len(c1.ServiceLines) != 1 {
		t.Errorf("claim 1: expected 1 service line, got %d", len(c1.ServiceLines))
	}

	sl1 := c1.ServiceLines[0]
	if sl1.ProcedureCode != "99213" {
		t.Errorf("claim 1: expected ProcedureCode '99213', got %q", sl1.ProcedureCode)
	}
	if sl1.LineCharge != 500.0 {
		t.Errorf("claim 1: expected LineCharge 500.0, got %f", sl1.LineCharge)
	}
	if sl1.LinePayment != 450.0 {
		t.Errorf("claim 1: expected LinePayment 450.0, got %f", sl1.LinePayment)
	}
	if len(sl1.CASAdjustments) != 1 {
		t.Errorf("claim 1 service line: expected 1 CAS, got %d", len(sl1.CASAdjustments))
	}

	// Second claim
	c2 := era.Claims[1]
	if c2.PatientControlNumber != "VC002" {
		t.Errorf("claim 2: expected PatientControlNumber 'VC002', got %q", c2.PatientControlNumber)
	}
	if c2.ClaimPayment != 300.0 {
		t.Errorf("claim 2: expected ClaimPayment 300.0, got %f", c2.ClaimPayment)
	}
	if len(c2.ServiceLines) != 1 {
		t.Errorf("claim 2: expected 1 service line, got %d", len(c2.ServiceLines))
	}

	sl2 := c2.ServiceLines[0]
	if sl2.ProcedureCode != "99214" {
		t.Errorf("claim 2: expected ProcedureCode '99214', got %q", sl2.ProcedureCode)
	}
}

func TestParse835_EmptyFile(t *testing.T) {
	_, err := Parse835([]byte{})
	if err == nil {
		t.Error("expected error for empty file")
	}
}

func TestParse835_NoClaims(t *testing.T) {
	noClaims := `ISA*00*          *00*          *ZZ*SUB             *ZZ*REC              *250101*1200*^*00501*0001*0*P*:~
GS*HP*SUB*REC*20250101*1200*1*X*005010X221A1~
ST*835*0001~
BPR*I*0.00*C*NON**20250101**PAYERID**PAYERNAME~
TRN*1**PAYERID~
SE*6*0001~
GE*1*1~
IEA*1*0001~`

	era, err := Parse835([]byte(noClaims))
	if err != nil {
		t.Fatalf("Parse835 error: %v", err)
	}
	if era == nil {
		t.Fatal("expected non-nil ERA")
	}
	if len(era.Claims) != 0 {
		t.Errorf("expected 0 claims, got %d", len(era.Claims))
	}
	if era.PayerName != "PAYERNAME" {
		t.Errorf("expected PayerName 'PAYERNAME', got %q", era.PayerName)
	}
}

func TestParse835_PLB(t *testing.T) {
	plb835 := `ISA*00*          *00*          *ZZ*SUB             *ZZ*REC              *250101*1200*^*00501*0001*0*P*:~
GS*HP*SUB*REC*20250101*1200*1*X*005010X221A1~
ST*835*0001~
BPR*I*100.00*C*CHK**20250101*TRACE001*PID01**PAYER~
TRN*1*TRACE001*PID01~
CLP*VC001*1*100.00*100.00*0.00**MC-001~
SVC*HC:99213*100.00*100.00~
DTM*472*20250101~
PLB*PROV123*20250101*WO*50.00~
PLB*PROV456*20250101*FB*-25.00~
SE*10*0001~
GE*1*1~
IEA*1*0001~`

	era, err := Parse835([]byte(plb835))
	if err != nil {
		t.Fatalf("Parse835 error: %v", err)
	}
	if len(era.ProviderAdjustments) != 2 {
		t.Fatalf("expected 2 PLB adjustments, got %d", len(era.ProviderAdjustments))
	}
	if era.ProviderAdjustments[0].ProviderID != "PROV123" {
		t.Errorf("expected PLB ProviderID 'PROV123', got %q", era.ProviderAdjustments[0].ProviderID)
	}
	if era.ProviderAdjustments[0].Amount != 50.0 {
		t.Errorf("expected PLB Amount 50.0, got %f", era.ProviderAdjustments[0].Amount)
	}
	if era.ProviderAdjustments[1].ProviderID != "PROV456" {
		t.Errorf("expected PLB ProviderID 'PROV456', got %q", era.ProviderAdjustments[1].ProviderID)
	}
	if era.ProviderAdjustments[1].Amount != -25.0 {
		t.Errorf("expected PLB Amount -25.0, got %f", era.ProviderAdjustments[1].Amount)
	}
}

func TestParse835WithMultipleCAS(t *testing.T) {
	multipleCAS := `ISA*00*          *00*          *ZZ*SUB             *ZZ*REC              *250201*1200*^*00501*0001*0*P*:~
GS*HP*SUB*REC*20250201*1200*1*X*005010X221A1~
ST*835*0001~
BPR*I*750.00*C*CHK**20250201*CHK789*PID01**PAYER~
TRN*1*CHK789*PID01~
N1*PR*PAYER~
N3*123 MAIN~
N4*HARTFORD*CT*06101~
REF*EV*PID01~
CLP*VC003*1*1000.00*750.00*250.00**MC-003~
CAS*CO*45*150.00~
CAS*CO*96*50.00~
CAS*OA*109*50.00~
NM1*QC*1*SMITH*BOB****MI*MEM003~
SVC*HC:99214:25*1000.00*750.00~
CAS*PR*3*100.00~
CAS*CO*45*100.00~
CAS*CO*96*50.00~
DTM*472*20250201~
SE*17*0001~
GE*1*1~
IEA*1*0001~`

	era, err := Parse835([]byte(multipleCAS))
	if err != nil {
		t.Fatalf("Parse835 error: %v", err)
	}

	if len(era.Claims) != 1 {
		t.Fatalf("expected 1 claim, got %d", len(era.Claims))
	}

	c := era.Claims[0]

	// Claim-level CAS: should have 3 adjustments
	if len(c.CASAdjustments) != 3 {
		t.Errorf("expected 3 claim-level CAS adjustments, got %d: %+v", len(c.CASAdjustments), c.CASAdjustments)
	}

	// Verify each claim-level CAS
	if len(c.CASAdjustments) >= 1 {
		if c.CASAdjustments[0].GroupCode != "CO" {
			t.Errorf("CAS[0] GroupCode = %q, want %q", c.CASAdjustments[0].GroupCode, "CO")
		}
		if c.CASAdjustments[0].ReasonCode != "45" {
			t.Errorf("CAS[0] ReasonCode = %q, want %q", c.CASAdjustments[0].ReasonCode, "45")
		}
		if c.CASAdjustments[0].Amount != 150.0 {
			t.Errorf("CAS[0] Amount = %f, want %f", c.CASAdjustments[0].Amount, 150.0)
		}
	}
	if len(c.CASAdjustments) >= 2 {
		if c.CASAdjustments[1].GroupCode != "CO" {
			t.Errorf("CAS[1] GroupCode = %q, want %q", c.CASAdjustments[1].GroupCode, "CO")
		}
		if c.CASAdjustments[1].ReasonCode != "96" {
			t.Errorf("CAS[1] ReasonCode = %q, want %q", c.CASAdjustments[1].ReasonCode, "96")
		}
	}
	if len(c.CASAdjustments) >= 3 {
		if c.CASAdjustments[2].GroupCode != "OA" {
			t.Errorf("CAS[2] GroupCode = %q, want %q", c.CASAdjustments[2].GroupCode, "OA")
		}
		if c.CASAdjustments[2].ReasonCode != "109" {
			t.Errorf("CAS[2] ReasonCode = %q, want %q", c.CASAdjustments[2].ReasonCode, "109")
		}
	}

	// Service-line level CAS: should have 3 adjustments
	if len(c.ServiceLines) != 1 {
		t.Fatalf("expected 1 service line, got %d", len(c.ServiceLines))
	}
	sl := c.ServiceLines[0]
	if len(sl.CASAdjustments) != 3 {
		t.Errorf("expected 3 service-line CAS adjustments, got %d: %+v", len(sl.CASAdjustments), sl.CASAdjustments)
	}

	if len(sl.CASAdjustments) >= 1 {
		if sl.CASAdjustments[0].GroupCode != "PR" {
			t.Errorf("line CAS[0] GroupCode = %q, want %q", sl.CASAdjustments[0].GroupCode, "PR")
		}
		if sl.CASAdjustments[0].ReasonCode != "3" {
			t.Errorf("line CAS[0] ReasonCode = %q, want %q", sl.CASAdjustments[0].ReasonCode, "3")
		}
		if sl.CASAdjustments[0].Amount != 100.0 {
			t.Errorf("line CAS[0] Amount = %f, want %f", sl.CASAdjustments[0].Amount, 100.0)
		}
	}
}

func TestNormalize835Date(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"20250101", "20250101"},
		{"250101", "20250101"},
		{"  20250101  ", "20250101"},
	}
	for _, tt := range tests {
		result := normalize835Date(tt.input)
		if result != tt.expected {
			t.Errorf("normalize835Date(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}
