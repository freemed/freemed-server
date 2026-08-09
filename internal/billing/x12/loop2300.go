package x12

// ---------- Loop 2300 – Claim Information ----------

// Loop2300 holds the claim-level data for an 837 Professional transaction.
type Loop2300 struct {
	CLM        Segment // CLM*CLAIM_ID*CHARGE***PLACE_OF_SERVICE:A:B:Y**Y*A*Y*Y*P~
	DTP434     Segment // DTP*434*RD8*FROM-TO~  (Statement Dates)
	DTP435     Segment // DTP*435*D8*CCYYMMDD~  (Admission Date, optional)
	DTP096     Segment // DTP*096*D8*CCYYMMDD~  (Discharge Date, optional)
	HI         Segment // HI*BK:DIAG1*BF:DIAG2*BF:DIAG3*BF:DIAG4~
	HI_Value   Segment // HI*BE:VALUE_INFO~  (optional value info)
	PWK        Segment // PWK*OZ*BM***DESC~  (Claim Supplemental Info, optional)
}

// ---------- Helper: build CLM segment ----------

// BuildCLM constructs a CLM segment from the given parameters.
//
//	sigOnFile: "A" or "B" — signature on file indicator
//	planPart:  "A" or "B" — plan participation code
//	benefitAssign: "Y" or "N" — benefits assignment certification
//
// Professional claims use signature-source "P" (provider).
func BuildCLM(claimID string, totalCharges float64, placeOfService, sigOnFile, planPart, benefitAssign string) Segment {
	return NewSegment("CLM",
		claimID,
		formatDollars(totalCharges),
		"",  // CLM03 reserved
		"",  // CLM04 reserved
		placeOfService+SubElementSep+sigOnFile+SubElementSep+planPart+SubElementSep+benefitAssign,
		"Y", // CLM06 – provider signature on file
		"A", // CLM07 – provider accepts assignment
		"Y", // CLM08 – benefits assignment certification
		"Y", // CLM09 – release of information
		"P", // CLM10 – provider signature source
	)
}

// BuildDTP constructs a date/time period segment.
func BuildDTP(qualifier, formatQual, dateValue string) Segment {
	return NewSegment("DTP", qualifier, formatQual, dateValue)
}

// BuildHI constructs an HI (Health Care Information Codes) segment for diagnoses.
// qualifier is typically "BK" for principal diagnosis and "BF" for additional.
func BuildHI(diagnosisCodes []string) Segment {
	// Format: BK:DIAG1*BF:DIAG2*BF:DIAG3...
	if len(diagnosisCodes) == 0 {
		return Segment{}
	}
	elements := make([]string, 0, len(diagnosisCodes))
	for i, code := range diagnosisCodes {
		qual := "BF" // additional
		if i == 0 {
			qual = "BK" // principal
		}
		elements = append(elements, qual+SubElementSep+code)
	}
	return NewSegment("HI", elements...)
}

// BuildHIValue constructs an HI segment for value information (BE qualifier).
func BuildHIValue(valueCode string) Segment {
	return NewSegment("HI", "BE"+SubElementSep+valueCode)
}
