package ccda

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

// =============================================================================
// C-CDA XML Parser — extracts clinical data from CCD documents
// =============================================================================

// ParseCCD parses a C-CDA Continuity of Care Document XML and returns
// structured clinical data from all recognized sections.
func ParseCCD(xmlData []byte) (*ParsedCCD, error) {
	var doc ClinicalDocument
	if err := xml.Unmarshal(xmlData, &doc); err != nil {
		return nil, fmt.Errorf("ccda: failed to unmarshal ClinicalDocument: %w", err)
	}

	parsed := &ParsedCCD{}

	// Extract patient demographics from recordTarget
	if len(doc.RecordTargets) > 0 {
		rt := doc.RecordTargets[0]
		parsed.Patient.FirstName = rt.PatientRole.Patient.Name.Given
		parsed.Patient.LastName = rt.PatientRole.Patient.Name.Family
		parsed.Patient.Gender = normalizeCCDAGender(rt.PatientRole.Patient.AdministrativeGenderCd.Code)
		parsed.Patient.DOB = rt.PatientRole.Patient.BirthTime.Value
		if len(rt.PatientRole.IDs) > 0 {
			parsed.Patient.MRN = rt.PatientRole.IDs[0].Extension
		}
		if rt.PatientRole.Addr != nil {
			addr := rt.PatientRole.Addr
			if len(addr.Lines) > 0 {
				parsed.Patient.AddressLine1 = addr.Lines[0]
			}
			if len(addr.Lines) > 1 {
				parsed.Patient.AddressLine2 = addr.Lines[1]
			}
			parsed.Patient.City = addr.City
			parsed.Patient.State = addr.State
			parsed.Patient.Zip = addr.PostalCode
		}
		if len(rt.PatientRole.Telecom) > 0 {
			parsed.Patient.Phone = rt.PatientRole.Telecom[0].Value
		}
	}

	// Walk structured body sections
	for _, sectionComp := range doc.Component.StructuredBody.Components {
		section := sectionComp.Section
		templateID := sectionTemplateID(section.TemplateID.Root)

		switch templateID {
		case "2.16.840.1.113883.10.20.22.2.6":
			// Allergies / Adverse Reactions
			parsed.Allergies = parseAllergiesSection(section)
		case "2.16.840.1.113883.10.20.22.2.1", "2.16.840.1.113883.10.20.22.2.1.1":
			// Medications
			parsed.Medications = parseMedicationsSection(section)
		case "2.16.840.1.113883.10.20.22.2.5", "2.16.840.1.113883.10.20.22.2.5.1":
			// Problem List
			parsed.Problems = parseProblemsSection(section)
		case "2.16.840.1.113883.10.20.22.2.4", "2.16.840.1.113883.10.20.22.2.4.1":
			// Vital Signs
			parsed.Vitals = parseVitalsSection(section)
		case "2.16.840.1.113883.10.20.22.2.2", "2.16.840.1.113883.10.20.22.2.2.1":
			// Immunizations
			parsed.Immunizations = parseImmunizationsSection(section)
		case "2.16.840.1.113883.10.20.22.2.7", "2.16.840.1.113883.10.20.22.2.7.1":
			// Procedures
			parsed.Procedures = parseProceduresSection(section)
		case "2.16.840.1.113883.10.20.22.2.3", "2.16.840.1.113883.10.20.22.2.3.1":
			// Results
			parsed.Results = parseResultsSection(section)
		case "2.16.840.1.113883.10.20.22.2.17":
			// Social History
			parsed.SocialHistory = parseSocialHistorySection(section)
		case "2.16.840.1.113883.10.20.22.2.22", "2.16.840.1.113883.10.20.22.2.22.1":
			// Encounters
			parsed.Encounters = parseEncountersSection(section)
		}
	}

	return parsed, nil
}

// sectionTemplateID returns the root template ID for a section, stripping version suffixes.
func sectionTemplateID(root string) string {
	return root
}

// parseAllergiesSection extracts allergy entries from an allergies section.
func parseAllergiesSection(section Section) []ParsedAllergy {
	var allergies []ParsedAllergy
	for _, entry := range section.Entries {
		if entry.Obs == nil {
			continue
		}
		obs := entry.Obs
		allergy := ParsedAllergy{
			Substance: obs.Code.DisplayName,
			Code:      obs.Code.Code,
			CodeSystem: obs.Code.CodeSystem,
			Status:    normalizeObsStatus(obs.StatusCd.Code),
		}
		if obs.Value != nil {
			allergy.Reaction = obs.Value.DisplayName
		}
		if obs.EffTime != nil {
			allergy.OnsetDate = obs.EffTime.Value
		}
		allergies = append(allergies, allergy)
	}
	return allergies
}

// parseMedicationsSection extracts medication entries from a medications section.
func parseMedicationsSection(section Section) []ParsedMedication {
	var meds []ParsedMedication
	for _, entry := range section.Entries {
		if entry.SubAdm == nil {
			continue
		}
		sa := entry.SubAdm
		med := ParsedMedication{
			Status: normalizeObsStatus(sa.StatusCode.Code),
		}
		if sa.Consumable != nil {
			med.DrugName = sa.Consumable.ManufacturedPd.ManufLabel
		}
		if sa.EffTime != nil {
			if sa.EffTime.Low != nil {
				med.StartDate = sa.EffTime.Low.Value
			}
			if sa.EffTime.High != nil {
				med.EndDate = sa.EffTime.High.Value
			}
		}
		if med.DrugName != "" {
			meds = append(meds, med)
		}
	}
	return meds
}

// parseProblemsSection extracts problem/condition entries from a problems section.
func parseProblemsSection(section Section) []ParsedProblem {
	var problems []ParsedProblem
	for _, entry := range section.Entries {
		if entry.Obs == nil {
			continue
		}
		obs := entry.Obs
		problem := ParsedProblem{
			Condition: obs.Code.DisplayName,
			ICD10Code: obs.Code.Code,
			Status:    normalizeObsStatus(obs.StatusCd.Code),
		}
		if obs.EffTime != nil {
			problem.OnsetDate = obs.EffTime.Value
		}
		if problem.Condition != "" {
			problems = append(problems, problem)
		}
	}
	return problems
}

// parseVitalsSection extracts vital sign observations from a vitals section.
func parseVitalsSection(section Section) []ParsedVitalSign {
	var vitals []ParsedVitalSign
	for _, entry := range section.Entries {
		if entry.Org != nil {
			// Vital signs organizer — contains component observations
			vs := ParsedVitalSign{}
			for _, comp := range entry.Org.Components {
				obs := comp.Obs
				switch obs.Code.Code {
				case "8480-6": vs.Systolic = parseCCDAValue(obs.Value)
				case "8462-4": vs.Diastolic = parseCCDAValue(obs.Value)
				case "8867-4": vs.HeartRate = parseCCDAValue(obs.Value)
				case "9279-1": vs.RespRate = parseCCDAValue(obs.Value)
				case "8310-5": vs.TempF = parseCCDAValue(obs.Value)
				case "2710-2": vs.O2Sat = parseCCDAValue(obs.Value)
				case "8302-2": vs.HeightCm = parseCCDAValue(obs.Value)
				case "29463-7": vs.WeightKg = parseCCDAValue(obs.Value)
				case "39156-5": vs.BMI = parseCCDAValue(obs.Value)
				}
			}
			vitals = append(vitals, vs)
		}
	}
	return vitals
}

// parseImmunizationsSection extracts immunization entries.
func parseImmunizationsSection(section Section) []ParsedImmunization {
	var imms []ParsedImmunization
	for _, entry := range section.Entries {
		if entry.SubAdm != nil {
			sa := entry.SubAdm
			imm := ParsedImmunization{
				Vaccine: sa.Consumable.ManufacturedPd.ManufLabel,
			}
			if sa.EffTime != nil && sa.EffTime.Low != nil {
				imm.DateGiven = sa.EffTime.Low.Value
			}
			if imm.Vaccine != "" {
				imms = append(imms, imm)
			}
		}
	}
	return imms
}

// parseProceduresSection extracts procedure entries from a procedures section.
func parseProceduresSection(section Section) []ParsedProcedure {
	var procs []ParsedProcedure
	for _, entry := range section.Entries {
		if entry.Proc != nil {
			p := entry.Proc
			proc := ParsedProcedure{}
			if p.Code != nil {
				proc.Procedure = p.Code.DisplayName
				proc.CPTCode = p.Code.Code
			}
			if p.EffTime != nil {
				proc.Date = p.EffTime.Value
			}
			if proc.Procedure != "" {
				procs = append(procs, proc)
			}
		}
	}
	return procs
}

// parseResultsSection extracts lab/diagnostic result entries.
func parseResultsSection(section Section) []ParsedResult {
	var results []ParsedResult
	for _, entry := range section.Entries {
		if entry.Org != nil {
			for _, comp := range entry.Org.Components {
				obs := comp.Obs
				result := ParsedResult{
					TestName:  obs.Code.DisplayName,
					LOINCCode: obs.Code.Code,
				}
				if obs.Value != nil {
					result.ResultValue = obs.Value.Code
				}
				if obs.EffTime != nil {
					result.ResultDate = obs.EffTime.Value
				}
				results = append(results, result)
			}
		} else if entry.Obs != nil {
			obs := entry.Obs
			result := ParsedResult{
				TestName:  obs.Code.DisplayName,
				LOINCCode: obs.Code.Code,
			}
			if obs.Value != nil {
				result.ResultValue = obs.Value.Code
			}
			if obs.EffTime != nil {
				result.ResultDate = obs.EffTime.Value
			}
			results = append(results, result)
		}
	}
	return results
}

// parseSocialHistorySection extracts social history observations.
func parseSocialHistorySection(section Section) []ParsedSocialHistory {
	var sh []ParsedSocialHistory
	for _, entry := range section.Entries {
		if entry.Obs == nil {
			continue
		}
		obs := entry.Obs
		cat := categorizeSocialHistory(obs.Code.Code, obs.Code.DisplayName)
		s := ParsedSocialHistory{
			Category: cat,
		}
		if obs.Value != nil {
			s.Value = obs.Value.Code
			s.Detail = obs.Value.DisplayName
		}
		if obs.EffTime != nil {
			s.EffectiveDate = obs.EffTime.Value
		}
		if s.Category != "" {
			sh = append(sh, s)
		}
	}
	return sh
}

// parseEncountersSection extracts encounter entries.
func parseEncountersSection(section Section) []ParsedEncounter {
	var encs []ParsedEncounter
	for _, entry := range section.Entries {
		if entry.Enc != nil {
			e := entry.Enc
			enc := ParsedEncounter{}
			if e.Code != nil {
				enc.EncounterType = e.Code.DisplayName
			}
			if e.EffTime != nil {
				enc.Date = e.EffTime.Value
			}
			encs = append(encs, enc)
		}
	}
	return encs
}

// =============================================================================
// Helper functions
// =============================================================================

// parseCCDAValue extracts a numeric value from a CD element's code attribute.
func parseCCDAValue(cd *CD) *float64 {
	if cd == nil || cd.Code == "" {
		return nil
	}
	v, err := strconv.ParseFloat(cd.Code, 64)
	if err != nil {
		return nil
	}
	return &v
}

// normalizeCCDAGender converts CCDA gender codes (M/F/UN) to FreeMED format.
func normalizeCCDAGender(code string) string {
	switch strings.ToUpper(code) {
	case "M":
		return "M"
	case "F":
		return "F"
	default:
		return "U"
	}
}

// normalizeObsStatus converts CDA status codes to human-readable status.
func normalizeObsStatus(code string) string {
	switch code {
	case "active", "55561003":
		return "active"
	case "completed", "normal":
		return "completed"
	case "aborted", "cancelled":
		return "inactive"
	case "resolved", "413322009":
		return "resolved"
	default:
		return code
	}
}

// categorizeSocialHistory maps LOINC/SNOMED codes or display names to social history categories.
func categorizeSocialHistory(code, displayName string) string {
	codeLower := strings.ToLower(code)
	nameLower := strings.ToLower(displayName)

	if strings.Contains(codeLower, "smoking") || strings.Contains(nameLower, "smoking") ||
		strings.Contains(nameLower, "tobacco") {
		return "smoking"
	}
	if strings.Contains(codeLower, "alcohol") || strings.Contains(nameLower, "alcohol") {
		return "alcohol"
	}
	if strings.Contains(codeLower, "drug") || strings.Contains(nameLower, "drug") ||
		strings.Contains(nameLower, "substance") {
		return "drug_use"
	}
	if strings.Contains(nameLower, "occupation") || strings.Contains(nameLower, "employment") {
		return "occupation"
	}
	if strings.Contains(nameLower, "exercise") || strings.Contains(nameLower, "activity") {
		return "exercise"
	}
	return "other"
}
