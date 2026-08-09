package ncpdp

import (
	"encoding/xml"
	"fmt"
	"time"
)

// =============================================================================
// NewRx Encoder — Generates NCPDP SCRIPT v2017071 XML
// =============================================================================

// EncodeNewRx generates a complete NCPDP SCRIPT NewRx message as XML bytes.
func EncodeNewRx(rx *NewRx, senderID, receiverID string) ([]byte, error) {
	msg := Message{
		Header: MessageHeader{
			To:        receiverID,
			From:      senderID,
			MessageID: generateMessageID(),
			SentTime:  time.Now().UTC().Format(time.RFC3339),
		},
		Body: MessageBody{
			NewRx: rx,
		},
	}

	// NCPDP SCRIPT uses a specific namespace
	xmlBytes, err := xml.MarshalIndent(msg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("ncpdp: failed to marshal NewRx: %w", err)
	}

	// Add XML declaration
	output := append([]byte(xml.Header), xmlBytes...)
	return output, nil
}

// generateMessageID creates a unique message identifier.
func generateMessageID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// =============================================================================
// PrescriptionInput — data needed to assemble a NewRx from the database
// =============================================================================

// PrescriptionInput holds assembled data from FreeMED database tables
// (prescriptions, patient, physician, pharmacy) ready for NewRx encoding.
type PrescriptionInput struct {
	// Sender info
	SenderNCPDPID string

	// Pharmacy info
	PharmacyNCPDPID string
	PharmacyName     string
	PharmacyNPI      string
	PharmacyPhone    string
	PharmacyFax      string

	// Prescriber info
	PrescriberNPI   string
	PrescriberDEA   string
	PrescriberFName string
	PrescriberLName string
	PrescriberPhone string

	// Patient info
	PatientFName   string
	PatientLName   string
	PatientDOB     string // YYYY-MM-DD
	PatientGender  string // M/F/U
	PatientMRN     string
	PatientPhone   string
	PatientAddr1   string
	PatientAddr2   string
	PatientCity    string
	PatientState   string
	PatientZip     string

	// Medication info
	DrugName      string
	NDCCode       string
	RxNormCode    string
	Dosage        string
	Frequency     string
	SIG           string
	Quantity      string
	QuantityUnit  string
	DaysSupply    string
	Refills       string
	Substitution  string
	WrittenDate   string // YYYY-MM-DD
	DiagnosisCode string
	DiagnosisQual string // ABF=ICD-10, ABK=ICD-9

	// Insurance
	PayerID       string
	PayerName     string
	CardholderID  string
	GroupID       string
	PersonCode    string
}

// AssembleNewRx builds a NewRx struct from database-assembled PrescriptionInput.
func AssembleNewRx(in *PrescriptionInput) *NewRx {
	rx := &NewRx{
		MedicationPrescribed: MedicationPrescribed{
			DrugDescription: in.DrugName,
			Quantity: QuantityVal{
				Value: in.Quantity,
				Unit:  in.QuantityUnit,
			},
			DaysSupply:   in.DaysSupply,
			Directions:   in.SIG,
			Refills:      in.Refills,
			Substitution: in.Substitution,
			WrittenDate:  in.WrittenDate,
		},
		Prescriber: Prescriber{
			NPI:        in.PrescriberNPI,
			DEANumber:  in.PrescriberDEA,
			FirstName:  in.PrescriberFName,
			LastName:   in.PrescriberLName,
		},
		Patient: NCPDPPatient{
			FirstName:   in.PatientFName,
			LastName:    in.PatientLName,
			DateOfBirth: in.PatientDOB,
			Gender:      normalizeGender(in.PatientGender),
			PatientID:   in.PatientMRN,
		},
	}

	// Drug coding
	if in.NDCCode != "" {
		rx.MedicationPrescribed.DrugCoded = &DrugCoded{
			ProductCode:          in.NDCCode,
			ProductCodeQualifier: "ND",
		}
	} else if in.RxNormCode != "" {
		rx.MedicationPrescribed.DrugCoded = &DrugCoded{
			ProductCode:          in.RxNormCode,
			ProductCodeQualifier: "RX",
		}
	}

	// Pharmacy
	if in.PharmacyNCPDPID != "" {
		pharm := &NCPDPPharmacy{
			NCPDPID:   in.PharmacyNCPDPID,
			NPI:       in.PharmacyNPI,
			StoreName: in.PharmacyName,
		}
		if in.PharmacyPhone != "" {
			pharm.Phone = &NCPDPPhone{Number: in.PharmacyPhone}
		}
		if in.PharmacyFax != "" {
			pharm.Fax = &NCPDPPhone{Number: in.PharmacyFax}
		}
		rx.Pharmacy = pharm
	}

	// Patient address
	if in.PatientAddr1 != "" {
		rx.Patient.Address = &NCPDPAddress{
			AddressLine1: in.PatientAddr1,
			AddressLine2: in.PatientAddr2,
			City:         in.PatientCity,
			State:        in.PatientState,
			ZipCode:      in.PatientZip,
		}
	}

	// Patient phone
	if in.PatientPhone != "" {
		rx.Patient.Phone = &NCPDPPhone{Number: in.PatientPhone}
	}

	// Prescriber phone
	if in.PrescriberPhone != "" {
		rx.Prescriber.Phone = &NCPDPPhone{Number: in.PrescriberPhone}
	}

	// Prescriber address
	if in.PatientAddr1 != "" {
		rx.Prescriber.Address = &NCPDPAddress{
			AddressLine1: in.PatientAddr1,
			City:         in.PatientCity,
			State:        in.PatientState,
			ZipCode:      in.PatientZip,
		}
	}

	// Diagnosis
	if in.DiagnosisCode != "" {
		rx.MedicationPrescribed.Diagnosis = &DiagnosisRef{
			Code:      in.DiagnosisCode,
			Qualifier: in.DiagnosisQual,
		}
	}

	// Benefits coordination
	if in.PayerID != "" || in.CardholderID != "" {
		rx.BenefitsCoordination = &BenefitsCoordination{
			PayerID:      in.PayerID,
			PayerName:    in.PayerName,
			CardholderID: in.CardholderID,
			GroupID:      in.GroupID,
			PersonCode:   in.PersonCode,
		}
	}

	// SIG as observation
	rx.Observations = []NCPDPObservation{
		{Type: "SIG", Value: in.SIG},
	}
	if in.Dosage != "" {
		rx.Observations = append(rx.Observations, NCPDPObservation{Type: "Dosage", Value: in.Dosage})
	}
	if in.Frequency != "" {
		rx.Observations = append(rx.Observations, NCPDPObservation{Type: "Frequency", Value: in.Frequency})
	}

	return rx
}

// normalizeGender converts FreeMED gender codes to NCPDP format (M/F/U).
func normalizeGender(gender string) string {
	switch gender {
	case "m", "M", "male", "Male":
		return "M"
	case "f", "F", "female", "Female":
		return "F"
	default:
		return "U"
	}
}
