// Package ncpdp provides NCPDP SCRIPT v2017071 message types and encoding
// for electronic prescribing (e-Prescribing / eRx).
//
// NCPDP SCRIPT is the standard for electronic transmission of prescriptions,
// renewals, medication history, and related pharmacy communications.
// This package implements the NewRx (new prescription) message type.
package ncpdp

import "encoding/xml"

// =============================================================================
// NCPDP SCRIPT v2017071 Message Envelope
// =============================================================================

// Message is the root NCPDP SCRIPT message envelope.
type Message struct {
	XMLName xml.Name       `xml:"Message"`
	Header  MessageHeader  `xml:"Header"`
	Body    MessageBody    `xml:"Body"`
}

// MessageHeader contains transport-level metadata.
type MessageHeader struct {
	To         string `xml:"To"`         // NCPDPID of receiver (pharmacy)
	From       string `xml:"From"`       // NCPDPID of sender (prescriber system)
	MessageID  string `xml:"MessageID"`  // Unique message identifier
	SentTime   string `xml:"SentTime"`   // ISO 8601 timestamp
	RelatesToMessageID string `xml:"RelatesToMessageID,omitempty"` // For responses
}

// MessageBody wraps the transaction type.
type MessageBody struct {
	NewRx          *NewRx          `xml:"NewRx,omitempty"`
	RenewalRequest *RenewalRequest `xml:"RenewalRequest,omitempty"`
	CancelRx       *CancelRx       `xml:"CancelRx,omitempty"`
	Status         *Status         `xml:"Status,omitempty"`
	Error          *NCPDPError     `xml:"Error,omitempty"`
}

// =============================================================================
// NewRx — New Prescription
// =============================================================================

// NewRx represents a new prescription request (NewRx transaction).
type NewRx struct {
	XMLName               xml.Name               `xml:"NewRx"`
	MedicationPrescribed  MedicationPrescribed   `xml:"MedicationPrescribed"`
	Prescriber            Prescriber             `xml:"Prescriber"`
	Patient               NCPDPPatient           `xml:"Patient"`
	Pharmacy              *NCPDPPharmacy         `xml:"Pharmacy,omitempty"`
	BenefitsCoordination  *BenefitsCoordination  `xml:"BenefitsCoordination,omitempty"`
	Observations          []NCPDPObservation     `xml:"Observation,omitempty"`
	Notes                 string                 `xml:"Notes,omitempty"`
}

// MedicationPrescribed contains the drug and dosage details.
type MedicationPrescribed struct {
	DrugDescription string      `xml:"DrugDescription"`
	DrugCoded       *DrugCoded  `xml:"DrugCoded,omitempty"`
	Quantity        QuantityVal `xml:"Quantity"`
	DaysSupply      string      `xml:"DaysSupply,omitempty"`
	Directions      string      `xml:"Directions"`           // SIG
	Refills         string      `xml:"Refills,omitempty"`
	Substitution    string      `xml:"Substitution,omitempty"` // 0=no sub, 1=allow generic, 2=allow therapeutic
	WrittenDate     string      `xml:"WrittenDate"`          // YYYY-MM-DD
	LastFillDate    string      `xml:"LastFillDate,omitempty"`
	Diagnosis       *DiagnosisRef `xml:"Diagnosis,omitempty"`
}

// DrugCoded contains a coded drug identifier (NDC or RxNorm).
type DrugCoded struct {
	ProductCode          string `xml:"ProductCode"`
	ProductCodeQualifier string `xml:"ProductCodeQualifier"` // ND=NDC, RX=RxNorm
}

// QuantityVal holds the quantity value and unit of measure.
type QuantityVal struct {
	Value string `xml:"Value"`
	Unit  string `xml:"CodeListID,attr"` // e.g. "UnitOfMeasure"
}

// DiagnosisRef references a diagnosis code qualifying the prescription.
type DiagnosisRef struct {
	Code        string `xml:"DiagnosisCode"`
	Qualifier   string `xml:"Qualifier,omitempty"` // ABF=ICD-10, ABK=ICD-9
}

// Prescriber holds the prescriber's identifying information.
type Prescriber struct {
	NPI           string  `xml:"PrescriberNPI"`
	LastName      string  `xml:"LastName"`
	FirstName     string  `xml:"FirstName"`
	MiddleName    string  `xml:"MiddleName,omitempty"`
	DEANumber     string  `xml:"DEANumber,omitempty"`
	SPI           string  `xml:"SPI,omitempty"`          // State Prescriber ID
	Phone         *NCPDPPhone `xml:"Phone,omitempty"`
	Address       *NCPDPAddress `xml:"Address,omitempty"`
	ClinicName    string  `xml:"ClinicName,omitempty"`
}

// NCPDPPatient holds the patient's identifying and demographic information.
type NCPDPPatient struct {
	LastName     string        `xml:"LastName"`
	FirstName    string        `xml:"FirstName"`
	MiddleName   string        `xml:"MiddleName,omitempty"`
	DateOfBirth  string        `xml:"DateOfBirth"`       // YYYY-MM-DD
	Gender       string        `xml:"Gender"`            // M/F/U
	Address      *NCPDPAddress `xml:"Address,omitempty"`
	Phone        *NCPDPPhone   `xml:"Phone,omitempty"`
	PatientID    string        `xml:"Patient,omitempty"` // Patient identifier (MRN)
}

// NCPDPPharmacy holds the pharmacy's identifying information.
type NCPDPPharmacy struct {
	NCPDPID     string        `xml:"NCPDPID"`
	NPI         string        `xml:"NPI,omitempty"`
	StoreName   string        `xml:"StoreName,omitempty"`
	Address     *NCPDPAddress `xml:"Address,omitempty"`
	Phone       *NCPDPPhone   `xml:"Phone,omitempty"`
	Fax         *NCPDPPhone   `xml:"Fax,omitempty"`
}

// NCPDPAddress holds a postal address.
type NCPDPAddress struct {
	AddressLine1 string `xml:"AddressLine1"`
	AddressLine2 string `xml:"AddressLine2,omitempty"`
	City         string `xml:"City"`
	State        string `xml:"State"`
	ZipCode      string `xml:"ZipCode"`
}

// NCPDPPhone holds a phone number.
type NCPDPPhone struct {
	Number    string `xml:"Number"`
	Extension string `xml:"Extension,omitempty"`
}

// BenefitsCoordination contains insurance coverage information.
type BenefitsCoordination struct {
	PayerID           string `xml:"PayerID,omitempty"`
	PayerName         string `xml:"PayerName,omitempty"`
	CardholderID      string `xml:"CardholderID,omitempty"`
	GroupID           string `xml:"GroupID,omitempty"`
	PersonCode        string `xml:"PersonCode,omitempty"` // 01=cardholder, 02=spouse, 03=child
}

// NCPDPObservation captures additional clinical data attached to the prescription.
type NCPDPObservation struct {
	Type  string `xml:"ObservationType"`
	Value string `xml:"ObservationValue"`
}

// =============================================================================
// RenewalRequest — Prescription Renewal
// =============================================================================

// RenewalRequest represents a prescription renewal authorization request.
type RenewalRequest struct {
	XMLName              xml.Name             `xml:"RenewalRequest"`
	PrescriberOrderNumber string              `xml:"PrescriberOrderNumber"`
	Pharmacy             NCPDPPharmacy        `xml:"Pharmacy"`
	MedicationPrescribed MedicationPrescribed `xml:"MedicationPrescribed"`
	Patient              NCPDPPatient         `xml:"Patient"`
	Prescriber           Prescriber           `xml:"Prescriber"`
}

// =============================================================================
// CancelRx — Cancel Prescription
// =============================================================================

// CancelRx represents a cancel prescription request.
type CancelRx struct {
	XMLName               xml.Name `xml:"CancelRx"`
	PrescriberOrderNumber  string   `xml:"PrescriberOrderNumber"`
	Reason                 string   `xml:"CancelReason,omitempty"`
}

// =============================================================================
// Status — Status Message
// =============================================================================

// Status represents a transaction status response.
type Status struct {
	XMLName xml.Name `xml:"Status"`
	Code    string   `xml:"Code"`    // 010=success, 020=error
	Message string   `xml:"Message,omitempty"`
}

// NCPDPError represents an error response.
type NCPDPError struct {
	XMLName xml.Name `xml:"Error"`
	Code    string   `xml:"Code"`
	Message string   `xml:"Message,omitempty"`
}
