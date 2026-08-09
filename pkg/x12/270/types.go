// Package x12_270 provides X12 270 (Eligibility Inquiry) generation
// and X12 271 (Eligibility Response) parsing.
package x12_270

// =============================================================================
// X12 270 Eligibility Inquiry Types
// =============================================================================

// EligibilityInquiry270 represents an X12 270 eligibility request transaction.
type EligibilityInquiry270 struct {
	SenderID        string // ISA06 — sender ID (clearinghouse/sender)
	ReceiverID      string // ISA08 — receiver ID (payer)
	TransactionDate string // ISA09 — YYYYMMDD
	GSDate          string // GS04 — YYYYMMDD
	GSTime          string // GS05 — HHMM
	GroupControlNo  string // GS06 — group control number
	TransactionNo   string // ST02 — transaction set control number

	// Information Source (Loop 2000A) — Payer
	InfoSource InfoSource

	// Information Receiver (Loop 2000B) — Provider/biller
	InfoReceiver InfoReceiver

	// Subscriber (Loop 2000C) — the insured person
	Subscriber SubscriberInfo

	// Dependent (Loop 2000D) — if subscriber is not the patient
	Dependent *DependentInfo

	// Eligibility query (Loop 2110C/D)
	ServiceTypes []string // EQ01 values: 30=Health Benefit Plan Coverage, etc.
}

// InfoSource represents the Information Source (Loop 2000A) — the payer.
type InfoSource struct {
	EntityID     string // NM109
	EntityIDQual string // NM108 — typically "PI" (Payer Identification)
	Name         string // NM103 — payer organization name
}

// InfoReceiver represents the Information Receiver (Loop 2000B) — the provider.
type InfoReceiver struct {
	EntityID     string // NM109 — provider NPI or tax ID
	EntityIDQual string // NM108 — "XX" for NPI, "FI" for Federal Tax ID
	Name         string // NM103 — provider/billing name
}

// SubscriberInfo represents the subscriber (Loop 2000C).
type SubscriberInfo struct {
	LastName     string // NM103
	FirstName    string // NM104
	MiddleName   string // NM105
	IDQualifier  string // NM108 — "MI" (Member ID), "II" (Standard Unique Health ID)
	MemberID     string // NM109 — subscriber/member ID
	GroupNumber  string // REF02 when REF01="6P"
	DOB          string // DMG02 — YYYYMMDD
	Gender       string // DMG03 — M/F/U
	Address1     string // N301
	Address2     string // N302
	City         string // N401
	State        string // N402
	Zip          string // N403
}

// DependentInfo represents the dependent (Loop 2000D).
type DependentInfo struct {
	LastName     string // NM103
	FirstName    string // NM104
	DOB          string // DMG02
	Gender       string // DMG03
	Relationship string // INS02 — 01=Spouse, 19=Child, 34=Other, etc.
}

// =============================================================================
// X12 271 Eligibility Response Types
// =============================================================================

// EligibilityResponse271 represents a parsed X12 271 response.
type EligibilityResponse271 struct {
	PayerName       string // NM103 from Loop 2000A
	PayerID         string // NM109 from Loop 2000A
	TraceNumber     string // TRN02 — matching trace/control number
	ResponseDate    string // BPR16 — YYYYMMDD

	// Subscriber eligibility
	Subscriber EligibilitySubscriber

	// Dependent eligibility (if applicable)
	Dependent *EligibilitySubscriber

	Benefits []EligibilityBenefit
}

// EligibilitySubscriber represents subscriber or dependent eligibility.
type EligibilitySubscriber struct {
	LastName       string
	FirstName      string
	MemberID       string
	EligibilityDate string // DTP03
	ActiveCoverage bool
	PlanCoverage   string
}

// EligibilityBenefit represents a benefit line from the 271 response.
type EligibilityBenefit struct {
	ServiceTypeCode string // EB03 — 1=Medical Care, 30=Health Benefit, 33=Chiropractic, 35=Dental, 48=Hospital, etc.
	ServiceTypeDesc string // Human-readable description
	CoverageLevel   string // EB01 — "1"=Active, "A"=Co-Insurance, "B"=Co-Payment, "C"=Deductible, "F"=Limitations, etc.
	PlanCoverage    string // EB04 — Plan description text
	TimePeriodQual  string // DTP01 — e.g., "291"=Plan, "292"=Benefit
	TimePeriod      string // DTP03 — YYYYMMDD-YYYYMMDD or YYYYMMDD
	InPlanNetwork   string // MSG text indicating in/out of network

	// Monetary amounts (EB07)
	Amount     float64 // EB07 — generic monetary amount
	Quantity   float64 // EB09 — quantity value

	// Additional REF qualifiers
	RefIDQual string
	RefID     string
}
