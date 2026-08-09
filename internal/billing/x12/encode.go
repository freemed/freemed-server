package x12

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// ---------- Data structures ----------

// Claim837P holds all data needed to build a complete 837P transaction set.
type Claim837P struct {
	SubmitterName string
	SubmitterID   string
	ReceiverName  string
	ReceiverID    string

	BillingProvider ProviderInfo
	PayToProvider   *ProviderInfo // optional; nil if same as billing
	Subscriber      SubscriberInfo
	Patient         PatientInfo
	Claim           ClaimInfo
	ServiceLines    []ServiceLineInfo

	// Optional: Referring and Rendering providers
	ReferringProvider  *ProviderRefInfo
	RenderingProvider  *ProviderRefInfo
	RenderingProvider2 *ProviderRefInfo // for additional rendering providers
}

// ProviderInfo holds billing/pay-to provider demographic information.
type ProviderInfo struct {
	Name    string
	NPI     string
	TaxID   string
	Address1 string
	City    string
	State   string
	Zip     string
	Phone   string
}

// ProviderRefInfo holds referring or rendering provider information.
type ProviderRefInfo struct {
	LastName   string
	FirstName  string
	MiddleName string
	NPI        string
	Taxonomy   string // e.g. "207Q00000X"
}

// SubscriberInfo holds subscriber demographic and payer information.
type SubscriberInfo struct {
	LastName   string
	FirstName  string
	MiddleName string
	MemberID   string
	SSN        string
	DOB        string // YYYYMMDD
	Address1   string
	City       string
	State      string
	Zip        string
	PayerName  string
	PayerID    string
}

// PatientInfo holds patient demographic information.
type PatientInfo struct {
	LastName                 string
	FirstName                string
	DOB                      string // YYYYMMDD
	RelationshipToSubscriber string // "18" = self, "01" = spouse, "19" = child
	SSN                      string
	Address1                 string
	City                     string
	State                    string
	Zip                      string
}

// ClaimInfo holds claim-level data.
type ClaimInfo struct {
	ClaimID        string
	TotalCharges   float64
	PlaceOfService string   // "11" = office
	DiagnosisCodes []string // ICD-10 codes
	StatementFrom  string   // YYYYMMDD
	StatementTo    string   // YYYYMMDD
	AdmissionDate  string   // optional
	DischargeDate  string   // optional
	// Value codes (optional)
	ValueCodes []string
}

// ServiceLineInfo holds a single service line.
type ServiceLineInfo struct {
	LineNumber        int
	ProcedureCode     string // CPT/HCPCS
	Modifier1         string
	Modifier2         string
	Modifier3         string
	Modifier4         string
	Charge            float64
	Units             int
	DiagnosisPointers []int  // 1-based into DiagnosisCodes
	ServiceDate       string // YYYYMMDD
}

// ---------- Encoder ----------

// Encode837Professional assembles a complete 837P (005010X222A1) transaction set.
func Encode837Professional(claim *Claim837P) ([]byte, error) {
	if claim == nil {
		return nil, fmt.Errorf("x12: claim is nil")
	}

	var buf bytes.Buffer

	// Determine transaction set control number (use claim ID hash or timestamp)
	tsn := fmt.Sprintf("%04d", time.Now().UnixNano()%10000)
	groupCtrl := fmt.Sprintf("%d", time.Now().UnixNano()%1000000)

	// ---- ISA ----
	isa := NewSegment("ISA",
		"00",          // ISA01 authorization qualifier
		"          ",  // ISA02 authorization info
		"00",          // ISA03 security qualifier
		"          ",  // ISA04 security info
		"ZZ",          // ISA05 sender qualifier
		padOrTrunc(claim.SubmitterID, 15),   // ISA06 sender ID
		"ZZ",          // ISA07 receiver qualifier
		padOrTrunc(claim.ReceiverID, 15),    // ISA08 receiver ID
		time.Now().Format("060102"),          // ISA09 date
		fmt.Sprintf("%04d", time.Now().Hour()*100+time.Now().Minute()), // ISA10 time
		"^",           // ISA11 repetition separator
		"00501",       // ISA12 version
		tsn,           // ISA13 control number
		"0",           // ISA14 acknowledgment requested
		"P",           // ISA15 test/production indicator
		SubElementSep, // ISA16 component element separator
	)
	buf.WriteString(isa.String())

	// ---- GS ----
	gs := NewSegment("GS",
		"HC",            // GS01 functional ID
		claim.SubmitterID, // GS02 sender
		claim.ReceiverID,  // GS03 receiver
		time.Now().Format("20060102"), // GS04 date
		fmt.Sprintf("%04d", time.Now().Hour()*100+time.Now().Minute()), // GS05 time
		groupCtrl,       // GS06 group control
		"X",             // GS07 agency
		"005010X222A1",  // GS08 version
	)
	buf.WriteString(gs.String())

	// ---- ST ----
	st := NewSegment("ST", "837", tsn, "005010X222A1")
	buf.WriteString(st.String())

	// ---- BHT ----
	bht := NewSegment("BHT",
		"0019", // BHT01 – transaction set purpose (information only… 00=original)
		"00",   // BHT02 – original
		claim.Claim.ClaimID, // BHT03 – originator application transaction identifier
		time.Now().Format("20060102"), // BHT04 – date created
		fmt.Sprintf("%04d", time.Now().Hour()*100+time.Now().Minute()), // BHT05 – time created
		"CH",  // BHT06 – transaction type code (chargeable)
	)
	buf.WriteString(bht.String())

	// ---- Loop 1000A: Submitter ----
	buf.WriteString(NewSegment("NM1",
		"41",              // NM101 – submitter
		"2",               // NM102 – non-person entity
		claim.SubmitterName,
		"", "", "",
		"", "",
		"46",              // NM108 – EIN
		claim.SubmitterID, // NM109
	).String())

	// ---- Loop 1000B: Receiver ----
	buf.WriteString(NewSegment("NM1",
		"40",              // NM101 – receiver
		"2",               // NM102 – non-person entity
		claim.ReceiverName,
		"", "", "",
		"", "",
		"46",              // NM108 – EIN
		claim.ReceiverID,  // NM109
	).String())

	// ---- Loop 2000A: Billing Provider HL ----
	hlBilling := NewSegment("HL",
		"1",    // HL01 – hierarchical ID
		"",     // HL02 – parent (none for top level)
		"20",   // HL03 – hierarchical level (information source)
		"1",    // HL04 – hierarchical child code (has children)
	)
	buf.WriteString(hlBilling.String())

	// PRV – Billing provider specialty
	buf.WriteString(NewSegment("PRV",
		"BI",  // PRV01 – billing
		"PXC", // PRV02 – provider taxonomy code
		"207Q00000X", // PRV03 – taxonomy
	).String())

	// ---- Loop 2010AA: Billing Provider Name ----
	bp := claim.BillingProvider
	buf.WriteString(NewSegment("NM1",
		"85",              // NM101 – billing provider
		"2",               // NM102 – non-person entity
		bp.Name,
		"", "", "",
		"", "",
		"XX",              // NM108 – NPI
		bp.NPI,            // NM109
	).String())
	buf.WriteString(NewSegment("N3", bp.Address1).String())
	buf.WriteString(NewSegment("N4", bp.City, bp.State, bp.Zip).String())
	buf.WriteString(NewSegment("REF", "EI", bp.TaxID).String())
	if bp.Phone != "" {
		buf.WriteString(NewSegment("PER", "IC", "", "TE", bp.Phone).String())
	}

	// ---- Loop 2010AB: Pay-to Provider (optional) ----
	if claim.PayToProvider != nil {
		pp := claim.PayToProvider
		buf.WriteString(NewSegment("NM1",
			"87",             // NM101 – pay-to provider
			"2",              // NM102 – non-person entity
			pp.Name,
			"", "", "",
			"", "",
			"XX",             // NM108 – NPI
			pp.NPI,           // NM109
		).String())
		buf.WriteString(NewSegment("N3", pp.Address1).String())
		buf.WriteString(NewSegment("N4", pp.City, pp.State, pp.Zip).String())
	}

	// ---- Loop 2000B: Subscriber HL ----
	hlSub := NewSegment("HL",
		"2",    // HL01 – hierarchical ID
		"1",    // HL02 – parent (billing provider)
		"22",   // HL03 – subscriber
		hasPatientClaims(claim), // HL04 – "1" if child HLs exist, "0" if not
	)
	buf.WriteString(hlSub.String())

	// SBR – Subscriber
	sub := claim.Subscriber
	buf.WriteString(NewSegment("SBR",
		"P",               // SBR01 – primary payer
		claim.Patient.RelationshipToSubscriber, // SBR02 – relationship
		sub.MemberID,      // SBR03 – insured group/policy number
		"",                // SBR04 – group name
		"",                // SBR05 – insurance type code
		"",                // SBR06 – Medicare plan code
		"",                // SBR07 – COBRA
		"",                // SBR08 – employment status
		"CI",              // SBR09 – claim filing indicator code
	).String())

	// PAT
	buf.WriteString(NewSegment("PAT",
		claim.Patient.RelationshipToSubscriber, // PAT01 – relationship
	).String())

	// ---- Loop 2010BA: Subscriber Name ----
	buf.WriteString(NewSegment("NM1",
		"IL",             // NM101 – insured/subscriber
		"1",              // NM102 – person
		sub.LastName,
		sub.FirstName,
		sub.MiddleName,
		"", "",
		"MI",             // NM108 – member ID
		sub.MemberID,     // NM109
	).String())
	if sub.Address1 != "" {
		buf.WriteString(NewSegment("N3", sub.Address1).String())
		buf.WriteString(NewSegment("N4", sub.City, sub.State, sub.Zip).String())
	}
	if sub.DOB != "" {
		buf.WriteString(NewSegment("DMG", "D8", sub.DOB).String())
	}
	if sub.SSN != "" {
		buf.WriteString(NewSegment("REF", "SY", sub.SSN).String())
	}

	// ---- Loop 2010BB: Payer Name ----
	buf.WriteString(NewSegment("NM1",
		"PR",             // NM101 – payer
		"2",              // NM102 – non-person entity
		sub.PayerName,
		"", "", "",
		"", "",
		"PI",             // NM108 – payer ID
		sub.PayerID,      // NM109
	).String())
	buf.WriteString(NewSegment("REF", "2U", sub.PayerID).String())

	// ---- Patient HL (if patient ≠ subscriber) ----
	patientIsSubscriber := claim.Patient.RelationshipToSubscriber == "18"
	if !patientIsSubscriber {
		// Loop 2000C: Patient HL
		hlPatient := NewSegment("HL",
			"3",    // HL01
			"2",    // HL02 – parent (subscriber)
			"23",   // HL03 – patient
			"0",    // HL04 – no children
		)
		buf.WriteString(hlPatient.String())

		buf.WriteString(NewSegment("PAT",
			claim.Patient.RelationshipToSubscriber,
		).String())

		// Loop 2010CA: Patient Name
		pat := claim.Patient
		buf.WriteString(NewSegment("NM1",
			"QC",             // NM101 – patient
			"1",              // NM102 – person
			pat.LastName,
			pat.FirstName,
			"", "", "",
			"", "",
		).String())
		if pat.Address1 != "" {
			buf.WriteString(NewSegment("N3", pat.Address1).String())
			buf.WriteString(NewSegment("N4", pat.City, pat.State, pat.Zip).String())
		}
		if pat.DOB != "" {
			buf.WriteString(NewSegment("DMG", "D8", pat.DOB).String())
		}
		if pat.SSN != "" {
			buf.WriteString(NewSegment("REF", "SY", pat.SSN).String())
		}
	}

	// ---- Loop 2300: Claim Information ----
	clm := claim.Claim
	clmSeg := BuildCLM(clm.ClaimID, clm.TotalCharges, clm.PlaceOfService,
		"A", "B", "Y")
	buf.WriteString(clmSeg.String())

	// Statement dates
	if clm.StatementFrom != "" && clm.StatementTo != "" {
		buf.WriteString(BuildDTP("434", "RD8", clm.StatementFrom+"-"+clm.StatementTo).String())
	}
	// Admission / Discharge dates
	if clm.AdmissionDate != "" {
		buf.WriteString(BuildDTP("435", "D8", clm.AdmissionDate).String())
	}
	if clm.DischargeDate != "" {
		buf.WriteString(BuildDTP("096", "D8", clm.DischargeDate).String())
	}

	// HI – diagnosis codes
	if len(clm.DiagnosisCodes) > 0 {
		buf.WriteString(BuildHI(clm.DiagnosisCodes).String())
	}

	// HI – value codes
	for _, vc := range clm.ValueCodes {
		buf.WriteString(BuildHIValue(vc).String())
	}

	// ---- Loop 2310A: Referring Provider (optional) ----
	if claim.ReferringProvider != nil {
		rp := claim.ReferringProvider
		buf.WriteString(NewSegment("NM1",
			"DN",            // NM101 – referring provider
			"1",             // NM102 – person
			rp.LastName,
			rp.FirstName,
			rp.MiddleName,
			"", "",
			"XX",            // NM108 – NPI
			rp.NPI,          // NM109
		).String())
		buf.WriteString(NewSegment("REF", "1C", rp.NPI).String())
	}

	// ---- Loop 2310B: Rendering Provider (optional) ----
	if claim.RenderingProvider != nil {
		rp := claim.RenderingProvider
		buf.WriteString(NewSegment("NM1",
			"82",            // NM101 – rendering provider
			"1",             // NM102 – person
			rp.LastName,
			rp.FirstName,
			rp.MiddleName,
			"", "",
			"XX",            // NM108 – NPI
			rp.NPI,          // NM109
		).String())
		buf.WriteString(NewSegment("REF", "1C", rp.NPI).String())
		if rp.Taxonomy != "" {
			buf.WriteString(NewSegment("PRV",
				"PE",         // PRV01 – performing
				"PXC",        // PRV02 – taxonomy code
				rp.Taxonomy,  // PRV03
			).String())
		}
	}

	if claim.RenderingProvider2 != nil {
		rp := claim.RenderingProvider2
		buf.WriteString(NewSegment("NM1",
			"82",            // NM101 – rendering provider
			"1",             // NM102 – person
			rp.LastName,
			rp.FirstName,
			rp.MiddleName,
			"", "",
			"XX",            // NM108 – NPI
			rp.NPI,          // NM109
		).String())
		buf.WriteString(NewSegment("REF", "1C", rp.NPI).String())
		if rp.Taxonomy != "" {
			buf.WriteString(NewSegment("PRV",
				"PE",         // PRV01 – performing
				"PXC",        // PRV02 – taxonomy code
				rp.Taxonomy,  // PRV03
			).String())
		}
	}

	// ---- Loop 2400: Service Lines ----
	for _, sl := range claim.ServiceLines {
		buf.WriteString(BuildLX(sl.LineNumber).String())

		// Collect modifiers
		mods := []string{}
		if sl.Modifier1 != "" {
			mods = append(mods, sl.Modifier1)
		}
		if sl.Modifier2 != "" {
			mods = append(mods, sl.Modifier2)
		}
		if sl.Modifier3 != "" {
			mods = append(mods, sl.Modifier3)
		}
		if sl.Modifier4 != "" {
			mods = append(mods, sl.Modifier4)
		}

		buf.WriteString(BuildSV1(
			sl.ProcedureCode,
			mods,
			sl.Charge,
			sl.Units,
			sl.DiagnosisPointers,
		).String())

		if sl.ServiceDate != "" {
			buf.WriteString(BuildDTP("472", "D8", sl.ServiceDate).String())
		}
	}

	// ---- SE ----
	// For now we use a placeholder count; in production you'd count segments.
	se := NewSegment("SE", "0", tsn)
	buf.WriteString(se.String())

	// ---- GE ----
	ge := NewSegment("GE", "1", groupCtrl)
	buf.WriteString(ge.String())

	// ---- IEA ----
	iea := NewSegment("IEA", "1", tsn)
	buf.WriteString(iea.String())

	return buf.Bytes(), nil
}

// hasPatientClaims returns "0" or "1" based on whether the claim has a separate patient loop.
func hasPatientClaims(claim *Claim837P) string {
	if claim.Patient.RelationshipToSubscriber != "" &&
		claim.Patient.RelationshipToSubscriber != "18" {
		return "1"
	}
	return "0"
}

// padOrTrunc pads or truncates a string to the given length.
func padOrTrunc(s string, length int) string {
	if len(s) > length {
		return s[:length]
	}
	return s + strings.Repeat(" ", length-len(s))
}
