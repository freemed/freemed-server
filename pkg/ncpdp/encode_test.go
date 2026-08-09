package ncpdp

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestEncodeNewRx_BasicPrescription(t *testing.T) {
	rx := &NewRx{
		MedicationPrescribed: MedicationPrescribed{
			DrugDescription: "Lisinopril 10mg Tablet",
			DrugCoded: &DrugCoded{
				ProductCode:          "00093720101",
				ProductCodeQualifier: "ND",
			},
			Quantity: QuantityVal{
				Value: "30",
				Unit:  "Tablet",
			},
			DaysSupply:   "30",
			Directions:   "Take one tablet by mouth daily",
			Refills:      "0",
			Substitution: "1",
			WrittenDate:  "2024-08-09",
		},
		Prescriber: Prescriber{
			NPI:       "1234567890",
			LastName:  "Smith",
			FirstName: "John",
			DEANumber: "FS1234567",
		},
		Patient: NCPDPPatient{
			LastName:    "Doe",
			FirstName:   "Jane",
			DateOfBirth: "1980-01-01",
			Gender:      "F",
		},
	}

	output, err := EncodeNewRx(rx, "SENDER_ID", "RECEIVER_ID")
	if err != nil {
		t.Fatalf("EncodeNewRx() error = %v", err)
	}

	// Should be non-empty
	if len(output) == 0 {
		t.Fatal("EncodeNewRx() returned empty output")
	}

	outputStr := string(output)

	// Should start with XML declaration
	if !strings.HasPrefix(outputStr, xml.Header) {
		t.Error("output should start with XML declaration")
	}

	// Parse back as XML to verify well-formedness
	var msg Message
	if err := xml.Unmarshal(output, &msg); err != nil {
		t.Fatalf("unable to parse output as XML: %v", err)
	}

	// Verify Message envelope
	if msg.XMLName.Local != "Message" {
		t.Errorf("root element name = %q, want %q", msg.XMLName.Local, "Message")
	}

	// Verify Header
	if msg.Header.To != "RECEIVER_ID" {
		t.Errorf("Header.To = %q, want %q", msg.Header.To, "RECEIVER_ID")
	}
	if msg.Header.From != "SENDER_ID" {
		t.Errorf("Header.From = %q, want %q", msg.Header.From, "SENDER_ID")
	}
	if msg.Header.MessageID == "" {
		t.Error("Header.MessageID should not be empty")
	}
	if msg.Header.SentTime == "" {
		t.Error("Header.SentTime should not be empty")
	}

	// Verify NewRx body
	if msg.Body.NewRx == nil {
		t.Fatal("Body.NewRx should not be nil")
	}

	nrx := msg.Body.NewRx

	// Medication
	if nrx.MedicationPrescribed.DrugDescription != "Lisinopril 10mg Tablet" {
		t.Errorf("DrugDescription = %q, want %q", nrx.MedicationPrescribed.DrugDescription, "Lisinopril 10mg Tablet")
	}
	if nrx.MedicationPrescribed.DrugCoded == nil {
		t.Error("DrugCoded should not be nil")
	} else {
		if nrx.MedicationPrescribed.DrugCoded.ProductCode != "00093720101" {
			t.Errorf("ProductCode = %q, want %q", nrx.MedicationPrescribed.DrugCoded.ProductCode, "00093720101")
		}
		if nrx.MedicationPrescribed.DrugCoded.ProductCodeQualifier != "ND" {
			t.Errorf("ProductCodeQualifier = %q, want %q", nrx.MedicationPrescribed.DrugCoded.ProductCodeQualifier, "ND")
		}
	}
	if nrx.MedicationPrescribed.Quantity.Value != "30" {
		t.Errorf("Quantity.Value = %q, want %q", nrx.MedicationPrescribed.Quantity.Value, "30")
	}
	if nrx.MedicationPrescribed.Directions != "Take one tablet by mouth daily" {
		t.Errorf("Directions = %q", nrx.MedicationPrescribed.Directions)
	}
	if nrx.MedicationPrescribed.Refills != "0" {
		t.Errorf("Refills = %q, want %q", nrx.MedicationPrescribed.Refills, "0")
	}

	// Prescriber
	if nrx.Prescriber.NPI != "1234567890" {
		t.Errorf("Prescriber.NPI = %q, want %q", nrx.Prescriber.NPI, "1234567890")
	}
	if nrx.Prescriber.LastName != "Smith" {
		t.Errorf("Prescriber.LastName = %q, want %q", nrx.Prescriber.LastName, "Smith")
	}

	// Patient
	if nrx.Patient.LastName != "Doe" {
		t.Errorf("Patient.LastName = %q, want %q", nrx.Patient.LastName, "Doe")
	}
	if nrx.Patient.DateOfBirth != "1980-01-01" {
		t.Errorf("Patient.DateOfBirth = %q, want %q", nrx.Patient.DateOfBirth, "1980-01-01")
	}
}

func TestEncodeNewRx_WithPharmacyAndBenefits(t *testing.T) {
	rx := &NewRx{
		MedicationPrescribed: MedicationPrescribed{
			DrugDescription: "Metformin 500mg Tablet",
			Quantity: QuantityVal{
				Value: "60",
				Unit:  "Tablet",
			},
			DaysSupply:  "30",
			Directions:  "Take one tablet twice daily with meals",
			Refills:     "3",
			WrittenDate: "2024-08-09",
		},
		Prescriber: Prescriber{
			NPI:       "9876543210",
			LastName:  "Johnson",
			FirstName: "Sarah",
			DEANumber: "MJ9876543",
		},
		Patient: NCPDPPatient{
			LastName:    "Williams",
			FirstName:   "Robert",
			DateOfBirth: "1965-06-15",
			Gender:      "M",
		},
		Pharmacy: &NCPDPPharmacy{
			NCPDPID:   "PHARM123",
			NPI:       "5556667777",
			StoreName: "Community Pharmacy",
			Phone:     &NCPDPPhone{Number: "555-0100"},
		},
		BenefitsCoordination: &BenefitsCoordination{
			PayerID:      "PAYER001",
			PayerName:    "HealthPlus Insurance",
			CardholderID: "CARD123456",
			GroupID:      "GROUP789",
			PersonCode:   "01",
		},
	}

	output, err := EncodeNewRx(rx, "CLINIC_NCPDP", "PHARM_NCPDP")
	if err != nil {
		t.Fatalf("EncodeNewRx() error = %v", err)
	}

	var msg Message
	if err := xml.Unmarshal(output, &msg); err != nil {
		t.Fatalf("unable to parse output as XML: %v", err)
	}

	nrx := msg.Body.NewRx

	// Pharmacy
	if nrx.Pharmacy == nil {
		t.Fatal("Pharmacy should not be nil")
	}
	if nrx.Pharmacy.NCPDPID != "PHARM123" {
		t.Errorf("Pharmacy.NCPDPID = %q, want %q", nrx.Pharmacy.NCPDPID, "PHARM123")
	}
	if nrx.Pharmacy.StoreName != "Community Pharmacy" {
		t.Errorf("Pharmacy.StoreName = %q, want %q", nrx.Pharmacy.StoreName, "Community Pharmacy")
	}
	if nrx.Pharmacy.Phone == nil {
		t.Error("Pharmacy.Phone should not be nil")
	} else if nrx.Pharmacy.Phone.Number != "555-0100" {
		t.Errorf("Pharmacy.Phone.Number = %q", nrx.Pharmacy.Phone.Number)
	}

	// Benefits
	if nrx.BenefitsCoordination == nil {
		t.Fatal("BenefitsCoordination should not be nil")
	}
	if nrx.BenefitsCoordination.PayerID != "PAYER001" {
		t.Errorf("BenefitsCoordination.PayerID = %q, want %q", nrx.BenefitsCoordination.PayerID, "PAYER001")
	}
	if nrx.BenefitsCoordination.CardholderID != "CARD123456" {
		t.Errorf("BenefitsCoordination.CardholderID = %q, want %q", nrx.BenefitsCoordination.CardholderID, "CARD123456")
	}
}

func TestEncodeNewRx_NCPDPNamespace(t *testing.T) {
	rx := &NewRx{
		MedicationPrescribed: MedicationPrescribed{
			DrugDescription: "Amoxicillin 500mg",
			Quantity: QuantityVal{
				Value: "21",
				Unit:  "Capsule",
			},
			Directions:  "Take one capsule three times daily",
			WrittenDate: "2024-08-09",
		},
		Prescriber: Prescriber{
			NPI:       "1111111111",
			LastName:  "Brown",
			FirstName: "Emily",
		},
		Patient: NCPDPPatient{
			LastName:    "Davis",
			FirstName:   "Michael",
			DateOfBirth: "1990-03-22",
			Gender:      "M",
		},
	}

	output, err := EncodeNewRx(rx, "SEND", "RECV")
	if err != nil {
		t.Fatalf("EncodeNewRx() error = %v", err)
	}

	outputStr := string(output)

	// Verify XML well-formed by parsing
	var msg Message
	if err := xml.Unmarshal(output, &msg); err != nil {
		t.Fatalf("XML parse failed: %v", err)
	}

	// Verify header has required fields
	if msg.Header.MessageID == "" {
		t.Error("MessageID missing")
	}
	if msg.Header.SentTime == "" {
		t.Error("SentTime missing")
	}

	// Verify body has NewRx
	if msg.Body.NewRx == nil {
		t.Fatal("NewRx missing in body")
	}

	// Verify the XML contains recognisable structure markers
	if !strings.Contains(outputStr, "<Message>") {
		t.Error("missing <Message> root element")
	}
	if !strings.Contains(outputStr, "<Header>") {
		t.Error("missing <Header> element")
	}
	if !strings.Contains(outputStr, "<Body>") {
		t.Error("missing <Body> element")
	}
	if !strings.Contains(outputStr, "<NewRx>") {
		t.Error("missing <NewRx> element")
	}
	if !strings.Contains(outputStr, "<MedicationPrescribed>") {
		t.Error("missing <MedicationPrescribed> element")
	}
	if !strings.Contains(outputStr, "<Prescriber>") {
		t.Error("missing <Prescriber> element")
	}
	if !strings.Contains(outputStr, "<Patient>") {
		t.Error("missing <Patient> element")
	}
	if !strings.Contains(outputStr, "<DrugDescription>") {
		t.Error("missing <DrugDescription> element")
	}
}

func TestEncodeNewRx_EmptyFieldsOmitted(t *testing.T) {
	// Test that optional fields with empty values are properly omitted from output
	rx := &NewRx{
		MedicationPrescribed: MedicationPrescribed{
			DrugDescription: "Test Drug",
			Quantity: QuantityVal{
				Value: "10",
				Unit:  "",
			},
			Directions:  "As directed",
			WrittenDate: "2024-08-09",
		},
		Prescriber: Prescriber{
			NPI:       "0000000000",
			LastName:  "Test",
			FirstName: "Doctor",
		},
		Patient: NCPDPPatient{
			LastName:    "Patient",
			FirstName:   "Test",
			DateOfBirth: "2000-01-01",
			Gender:      "M",
		},
	}

	output, err := EncodeNewRx(rx, "S", "R")
	if err != nil {
		t.Fatalf("EncodeNewRx() error = %v", err)
	}

	var msg Message
	if err := xml.Unmarshal(output, &msg); err != nil {
		t.Fatalf("XML parse failed: %v", err)
	}

	nrx := msg.Body.NewRx

	// Optional fields that were not set should be nil/empty
	if nrx.Pharmacy != nil {
		t.Error("Pharmacy should be nil when not provided")
	}
	if nrx.BenefitsCoordination != nil {
		t.Error("BenefitsCoordination should be nil when not provided")
	}
	if nrx.MedicationPrescribed.DrugCoded != nil {
		t.Error("DrugCoded should be nil when not provided")
	}
	if nrx.MedicationPrescribed.Diagnosis != nil {
		t.Error("Diagnosis should be nil when not provided")
	}
}

func TestEncodeNewRx_DifferentHeaderIDs(t *testing.T) {
	rx := &NewRx{
		MedicationPrescribed: MedicationPrescribed{
			DrugDescription: "Aspirin 81mg",
			Quantity: QuantityVal{
				Value: "100",
				Unit:  "Tablet",
			},
			Directions:  "Take one daily",
			WrittenDate: "2024-08-09",
		},
		Prescriber: Prescriber{
			NPI:       "2222222222",
			LastName:  "Lee",
			FirstName: "David",
		},
		Patient: NCPDPPatient{
			LastName:    "Kim",
			FirstName:   "Susan",
			DateOfBirth: "1975-12-01",
			Gender:      "F",
		},
	}

	// Test with different sender/receiver IDs
	output, err := EncodeNewRx(rx, "SENDER_ABC", "RECEIVER_XYZ")
	if err != nil {
		t.Fatalf("EncodeNewRx() error = %v", err)
	}

	var msg Message
	if err := xml.Unmarshal(output, &msg); err != nil {
		t.Fatalf("XML parse failed: %v", err)
	}

	if msg.Header.To != "RECEIVER_XYZ" {
		t.Errorf("Header.To = %q, want %q", msg.Header.To, "RECEIVER_XYZ")
	}
	if msg.Header.From != "SENDER_ABC" {
		t.Errorf("Header.From = %q, want %q", msg.Header.From, "SENDER_ABC")
	}
}
