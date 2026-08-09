package x12

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidateClaim issues edit checks against a populated Claim837P.
// Returns nil if no issues are found.
func ValidateClaim(claim *Claim837P) error {
	if claim == nil {
		return fmt.Errorf("claim is nil")
	}

	var errs []string

	// Required top-level fields
	if claim.SubmitterID == "" {
		errs = append(errs, "submitter ID is required")
	}
	if claim.ReceiverID == "" {
		errs = append(errs, "receiver ID is required")
	}
	if claim.BillingProvider.NPI == "" && claim.BillingProvider.TaxID == "" {
		errs = append(errs, "billing provider must have NPI or TaxID")
	}
	if claim.Subscriber.MemberID == "" {
		errs = append(errs, "subscriber member ID is required")
	}
	if claim.Claim.ClaimID == "" {
		errs = append(errs, "claim ID is required")
	}
	if claim.Claim.PlaceOfService == "" {
		errs = append(errs, "place of service is required")
	}
	if claim.Claim.StatementFrom == "" {
		errs = append(errs, "statement from date is required")
	}
	if claim.Claim.StatementTo == "" {
		errs = append(errs, "statement to date is required")
	}

	// Validate dates
	if claim.Claim.StatementFrom != "" && !isValidDate(claim.Claim.StatementFrom) {
		errs = append(errs, fmt.Sprintf("invalid statement from date: %s", claim.Claim.StatementFrom))
	}
	if claim.Claim.StatementTo != "" && !isValidDate(claim.Claim.StatementTo) {
		errs = append(errs, fmt.Sprintf("invalid statement to date: %s", claim.Claim.StatementTo))
	}
	if claim.Claim.AdmissionDate != "" && !isValidDate(claim.Claim.AdmissionDate) {
		errs = append(errs, fmt.Sprintf("invalid admission date: %s", claim.Claim.AdmissionDate))
	}
	if claim.Claim.DischargeDate != "" && !isValidDate(claim.Claim.DischargeDate) {
		errs = append(errs, fmt.Sprintf("invalid discharge date: %s", claim.Claim.DischargeDate))
	}
	if claim.Subscriber.DOB != "" && !isValidDate(claim.Subscriber.DOB) {
		errs = append(errs, fmt.Sprintf("invalid subscriber DOB: %s", claim.Subscriber.DOB))
	}
	if claim.Patient.DOB != "" && !isValidDate(claim.Patient.DOB) {
		errs = append(errs, fmt.Sprintf("invalid patient DOB: %s", claim.Patient.DOB))
	}

	// Validate ICD-10 codes (basic format check)
	for i, code := range claim.Claim.DiagnosisCodes {
		if !isValidDiagnosisCode(code) {
			errs = append(errs, fmt.Sprintf("invalid diagnosis code at index %d: %s", i, code))
		}
	}

	// Validate charges
	if claim.Claim.TotalCharges < 0 {
		errs = append(errs, "total charges cannot be negative")
	}

	// Validate service lines
	if len(claim.ServiceLines) == 0 {
		errs = append(errs, "at least one service line is required")
	}
	for i, sl := range claim.ServiceLines {
		if sl.ProcedureCode == "" {
			errs = append(errs, fmt.Sprintf("service line %d: procedure code is required", i+1))
		}
		if sl.Charge < 0 {
			errs = append(errs, fmt.Sprintf("service line %d: charge cannot be negative", i+1))
		}
		if sl.Units <= 0 {
			errs = append(errs, fmt.Sprintf("service line %d: units must be positive", i+1))
		}
		if sl.ServiceDate != "" && !isValidDate(sl.ServiceDate) {
			errs = append(errs, fmt.Sprintf("service line %d: invalid service date: %s", i+1, sl.ServiceDate))
		}

		// Validate diagnosis pointers (must be 1-based, within range)
		numDiag := len(claim.Claim.DiagnosisCodes)
		for j, ptr := range sl.DiagnosisPointers {
			if ptr < 1 {
				errs = append(errs, fmt.Sprintf("service line %d: diagnosis pointer %d is less than 1 (%d)", i+1, j+1, ptr))
			} else if numDiag > 0 && ptr > numDiag {
				errs = append(errs, fmt.Sprintf("service line %d: diagnosis pointer %d out of range (%d > %d)", i+1, j+1, ptr, numDiag))
			}
		}
	}

	// Validate subscriber relationship code (must be valid)
	validRel := map[string]bool{
		"01": true, "18": true, "19": true, "20": true, "21": true,
		"39": true, "40": true, "53": true, "G8": true,
	}
	if claim.Patient.RelationshipToSubscriber != "" && !validRel[claim.Patient.RelationshipToSubscriber] {
		errs = append(errs, fmt.Sprintf("invalid relationship to subscriber: %s", claim.Patient.RelationshipToSubscriber))
	}

	// Validate NPI format (10-digit numeric)
	if claim.BillingProvider.NPI != "" && !isValidNPI(claim.BillingProvider.NPI) {
		errs = append(errs, fmt.Sprintf("invalid billing provider NPI: %s", claim.BillingProvider.NPI))
	}
	if claim.ReferringProvider != nil && claim.ReferringProvider.NPI != "" && !isValidNPI(claim.ReferringProvider.NPI) {
		errs = append(errs, fmt.Sprintf("invalid referring provider NPI: %s", claim.ReferringProvider.NPI))
	}
	if claim.RenderingProvider != nil && claim.RenderingProvider.NPI != "" && !isValidNPI(claim.RenderingProvider.NPI) {
		errs = append(errs, fmt.Sprintf("invalid rendering provider NPI: %s", claim.RenderingProvider.NPI))
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation errors:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}

// isValidDate checks YYYYMMDD format and basic validity.
func isValidDate(date string) bool {
	if len(date) != 8 {
		return false
	}
	dateRE := regexp.MustCompile(`^\d{8}$`)
	if !dateRE.MatchString(date) {
		return false
	}
	// Accept any 8-digit number as valid date format
	// (full date parsing would require time.Parse but we're validating format only)
	return true
}

// isValidDiagnosisCode checks ICD-10 code format (3-7 alphanumeric with optional decimal).
func isValidDiagnosisCode(code string) bool {
	// ICD-10: starts with letter, 2-7 chars total, may contain a decimal
	if len(code) < 3 || len(code) > 8 {
		return false
	}
	// Loose match: starts with letter, rest alphanumeric with optional dot
	matched, _ := regexp.MatchString(`^[A-Za-z][A-Za-z0-9]*\.?[A-Za-z0-9]*$`, code)
	return matched
}

// isValidNPI checks NPI format (10-digit numeric).
func isValidNPI(npi string) bool {
	return regexp.MustCompile(`^\d{10}$`).MatchString(npi)
}
