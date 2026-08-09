package ccda

import (
	"encoding/xml"
	"fmt"
	"strings"
	"testing"
)

// buildCCDFixture constructs a minimal valid CCD XML fixture with
// allergies, medications, problems, and vitals sections.
func buildCCDFixture() []byte {
	doc := ClinicalDocument{
		Xmlns:     "urn:hl7-org:v3",
		RealmCode: CS{XMLName: xmlName("realmCode"), Code: "US"},
		TypeID:    II{XMLName: xmlName("typeId"), Root: "2.16.840.1.113883.1.3", Extension: "POCD_HD000040"},
		TemplateIDs: []II{
			{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.1.1"},
		},
		ID:    II{XMLName: xmlName("id"), Root: "test-root", Extension: "CCD-001"},
		Code:  CE{XMLName: xmlName("code"), Code: "34133-9", CodeSystem: "2.16.840.1.113883.6.1", DisplayName: "Summarization of episode note"},
		Title: ST{XMLName: xmlName("title"), Text: "Continuity of Care Document"},
		EffectiveTime:       TS{XMLName: xmlName("effectiveTime"), Value: "20240809000000-0500"},
		ConfidentialityCode: CS{XMLName: xmlName("confidentialityCode"), Code: "N", CodeSystem: "2.16.840.1.113883.5.25"},
		LanguageCode:        CS{XMLName: xmlName("languageCode"), Code: "en-US"},

		// Patient demographics
		RecordTargets: []RecordTarget{
			{
				XMLName: xmlName("recordTarget"),
				PatientRole: RecordTargetRole{
					XMLName: xmlName("patientRole"),
					IDs: []II{
						{XMLName: xmlName("id"), Root: "2.16.840.1.113883.19.5", Extension: "MRN-12345"},
					},
					Addr:    &AD{XMLName: xmlName("addr"), Use: "HP", City: "Springfield", State: "IL", PostalCode: "62701", Lines: []string{"456 Oak Ave"}},
					Telecom: []TEL{{XMLName: xmlName("telecom"), Use: "HP", Value: "tel:+1-555-0100"}},
					Patient: RecordTargetP{
						XMLName:                xmlName("patient"),
						Name:                   PN{XMLName: xmlName("name"), Family: "Smith", Given: "Alice"},
						AdministrativeGenderCd: CE{XMLName: xmlName("administrativeGenderCode"), Code: "F", CodeSystem: "2.16.840.1.113883.5.1", DisplayName: "Female"},
						BirthTime:              TS{XMLName: xmlName("birthTime"), Value: "19750515"},
					},
				},
			},
		},

		// Author
		Authors: []Author{
			{
				XMLName: xmlName("author"),
				Time:    TS{XMLName: xmlName("time"), Value: "20240809000000-0500"},
				AssignedAuthor: AuthorAssigned{
					XMLName:   xmlName("assignedAuthor"),
					AssignedP: &AuthorPerson{XMLName: xmlName("assignedPerson"), Name: PN{XMLName: xmlName("name"), Family: "Jones", Given: "Robert"}},
				},
			},
		},

		// Custodian
		Custodian: Custodian{
			XMLName: xmlName("custodian"),
			AssignedCust: CustodianAssigned{
				XMLName: xmlName("assignedCustodian"),
				RepOrg: CustodianOrganization{
					XMLName: xmlName("representedCustodianOrganization"),
					Name:    ON{XMLName: xmlName("name"), Text: "Springfield General Hospital"},
				},
			},
		},

		// Structured body with sections
		Component: ComponentBody{
			XMLName: xmlName("component"),
			StructuredBody: StructuredBody{
				XMLName: xmlName("structuredBody"),
				Components: []SectionComp{
					// Allergies section
					{
						XMLName: xmlName("component"),
						Section: Section{
							XMLName:    xmlName("section"),
							TemplateID: II{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.2.6"},
							Code:       CE{XMLName: xmlName("code"), Code: "48765-2", CodeSystem: "2.16.840.1.113883.6.1", DisplayName: "Allergies"},
							Title:      ST{XMLName: xmlName("title"), Text: "Allergies and Intolerances"},
							Entries: []SectionEntry{
								{
									XMLName: xmlName("entry"),
									Obs: &ObsEntry{
										XMLName: xmlName("observation"),
										Code:    CE{XMLName: xmlName("code"), Code: "91935001", CodeSystem: "2.16.840.1.113883.6.96", DisplayName: "Allergy to penicillin"},
										StatusCd: CS{XMLName: xmlName("statusCode"), Code: "active"},
										Value:    &CD{XMLName: xmlName("value"), Code: "247472004", CodeSystem: "2.16.840.1.113883.6.96", DisplayName: "Hives"},
										EffTime:  &TS{XMLName: xmlName("effectiveTime"), Value: "20200101"},
									},
								},
								{
									XMLName: xmlName("entry"),
									Obs: &ObsEntry{
										XMLName: xmlName("observation"),
										Code:    CE{XMLName: xmlName("code"), Code: "91936008", CodeSystem: "2.16.840.1.113883.6.96", DisplayName: "Allergy to sulfonamides"},
										StatusCd: CS{XMLName: xmlName("statusCode"), Code: "active"},
										EffTime:  &TS{XMLName: xmlName("effectiveTime"), Value: "20190510"},
									},
								},
							},
						},
					},

					// Medications section
					{
						XMLName: xmlName("component"),
						Section: Section{
							XMLName:    xmlName("section"),
							TemplateID: II{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.2.1"},
							Code:       CE{XMLName: xmlName("code"), Code: "10160-0", CodeSystem: "2.16.840.1.113883.6.1", DisplayName: "Medications"},
							Title:      ST{XMLName: xmlName("title"), Text: "Medications"},
							Entries: []SectionEntry{
								{
									XMLName: xmlName("entry"),
									SubAdm: &SubstanceAdm{
										XMLName:    xmlName("substanceAdministration"),
										StatusCode: CS{XMLName: xmlName("statusCode"), Code: "active"},
										EffTime: &EffTimeInterval{
											XMLName: xmlName("effectiveTime"),
											Low:     &TS{XMLName: xmlName("low"), Value: "20240101"},
										},
										Consumable: &SubAdmConsumable{
											XMLName: xmlName("consumable"),
											ManufacturedPd: SubAdmManufacturedPd{
												XMLName:    xmlName("manufacturedProduct"),
												ManufLabel: "Lisinopril 10mg Tablet",
											},
										},
									},
								},
								{
									XMLName: xmlName("entry"),
									SubAdm: &SubstanceAdm{
										XMLName:    xmlName("substanceAdministration"),
										StatusCode: CS{XMLName: xmlName("statusCode"), Code: "active"},
										EffTime: &EffTimeInterval{
											XMLName: xmlName("effectiveTime"),
											Low:     &TS{XMLName: xmlName("low"), Value: "20240315"},
										},
										Consumable: &SubAdmConsumable{
											XMLName: xmlName("consumable"),
											ManufacturedPd: SubAdmManufacturedPd{
												XMLName:    xmlName("manufacturedProduct"),
												ManufLabel: "Metformin 500mg Tablet",
											},
										},
									},
								},
								{
									XMLName: xmlName("entry"),
									SubAdm: &SubstanceAdm{
										XMLName:    xmlName("substanceAdministration"),
										StatusCode: CS{XMLName: xmlName("statusCode"), Code: "completed"},
										EffTime: &EffTimeInterval{
											XMLName: xmlName("effectiveTime"),
											Low:     &TS{XMLName: xmlName("low"), Value: "20230601"},
											High:    &TS{XMLName: xmlName("high"), Value: "20231231"},
										},
										Consumable: &SubAdmConsumable{
											XMLName: xmlName("consumable"),
											ManufacturedPd: SubAdmManufacturedPd{
												XMLName:    xmlName("manufacturedProduct"),
												ManufLabel: "Prednisone 5mg Tablet",
											},
										},
									},
								},
							},
						},
					},

					// Problems section
					{
						XMLName: xmlName("component"),
						Section: Section{
							XMLName:    xmlName("section"),
							TemplateID: II{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.2.5"},
							Code:       CE{XMLName: xmlName("code"), Code: "11450-4", CodeSystem: "2.16.840.1.113883.6.1", DisplayName: "Problem List"},
							Title:      ST{XMLName: xmlName("title"), Text: "Problems"},
							Entries: []SectionEntry{
								{
									XMLName: xmlName("entry"),
									Obs: &ObsEntry{
										XMLName: xmlName("observation"),
										Code:    CE{XMLName: xmlName("code"), Code: "I10", CodeSystem: "2.16.840.1.113883.6.90", DisplayName: "Essential hypertension"},
										StatusCd: CS{XMLName: xmlName("statusCode"), Code: "active"},
										EffTime:  &TS{XMLName: xmlName("effectiveTime"), Value: "20200101"},
									},
								},
								{
									XMLName: xmlName("entry"),
									Obs: &ObsEntry{
										XMLName: xmlName("observation"),
										Code:    CE{XMLName: xmlName("code"), Code: "E11.9", CodeSystem: "2.16.840.1.113883.6.90", DisplayName: "Type 2 diabetes mellitus"},
										StatusCd: CS{XMLName: xmlName("statusCode"), Code: "active"},
										EffTime:  &TS{XMLName: xmlName("effectiveTime"), Value: "20180601"},
									},
								},
								{
									XMLName: xmlName("entry"),
									Obs: &ObsEntry{
										XMLName: xmlName("observation"),
										Code:    CE{XMLName: xmlName("code"), Code: "J45.909", CodeSystem: "2.16.840.1.113883.6.90", DisplayName: "Unspecified asthma"},
										StatusCd: CS{XMLName: xmlName("statusCode"), Code: "resolved"},
										EffTime:  &TS{XMLName: xmlName("effectiveTime"), Value: "20150101"},
									},
								},
							},
						},
					},

					// Vitals section
					{
						XMLName: xmlName("component"),
						Section: Section{
							XMLName:    xmlName("section"),
							TemplateID: II{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.2.4"},
							Code:       CE{XMLName: xmlName("code"), Code: "8716-3", CodeSystem: "2.16.840.1.113883.6.1", DisplayName: "Vital Signs"},
							Title:      ST{XMLName: xmlName("title"), Text: "Vital Signs"},
							Entries: []SectionEntry{
								{
									XMLName: xmlName("entry"),
									Org: &Organizer{
										XMLName:    xmlName("organizer"),
										StatusCode: CS{XMLName: xmlName("statusCode"), Code: "completed"},
										Components: []OrgComp{
											{Obs: ObsEntry{XMLName: xmlName("observation"), Code: CE{XMLName: xmlName("code"), Code: "8480-6", DisplayName: "Systolic BP"}, Value: &CD{XMLName: xmlName("value"), Code: "120"}}},
											{Obs: ObsEntry{XMLName: xmlName("observation"), Code: CE{XMLName: xmlName("code"), Code: "8462-4", DisplayName: "Diastolic BP"}, Value: &CD{XMLName: xmlName("value"), Code: "80"}}},
											{Obs: ObsEntry{XMLName: xmlName("observation"), Code: CE{XMLName: xmlName("code"), Code: "8867-4", DisplayName: "Heart Rate"}, Value: &CD{XMLName: xmlName("value"), Code: "72"}}},
										},
									},
								},
								{
									XMLName: xmlName("entry"),
									Org: &Organizer{
										XMLName:    xmlName("organizer"),
										StatusCode: CS{XMLName: xmlName("statusCode"), Code: "completed"},
										Components: []OrgComp{
											{Obs: ObsEntry{XMLName: xmlName("observation"), Code: CE{XMLName: xmlName("code"), Code: "8480-6", DisplayName: "Systolic BP"}, Value: &CD{XMLName: xmlName("value"), Code: "118"}}},
											{Obs: ObsEntry{XMLName: xmlName("observation"), Code: CE{XMLName: xmlName("code"), Code: "8462-4", DisplayName: "Diastolic BP"}, Value: &CD{XMLName: xmlName("value"), Code: "78"}}},
											{Obs: ObsEntry{XMLName: xmlName("observation"), Code: CE{XMLName: xmlName("code"), Code: "29463-7", DisplayName: "Weight"}, Value: &CD{XMLName: xmlName("value"), Code: "68.5"}}},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	data, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		panic("failed to build CCD fixture: " + err.Error())
	}
	// Prepend XML header for proper parsing
	return append([]byte(xml.Header), data...)
}

func TestParseCCD_PatientDemographics(t *testing.T) {
	xmlData := buildCCDFixture()

	parsed, err := ParseCCD(xmlData)
	if err != nil {
		t.Fatalf("ParseCCD() error = %v", err)
	}
	if parsed == nil {
		t.Fatal("ParseCCD() returned nil")
	}

	p := parsed.Patient
	if p.FirstName != "Alice" {
		t.Errorf("Patient.FirstName = %q, want %q", p.FirstName, "Alice")
	}
	if p.LastName != "Smith" {
		t.Errorf("Patient.LastName = %q, want %q", p.LastName, "Smith")
	}
	if p.Gender != "F" {
		t.Errorf("Patient.Gender = %q, want %q", p.Gender, "F")
	}
	if p.DOB != "19750515" {
		t.Errorf("Patient.DOB = %q, want %q", p.DOB, "19750515")
	}
	if p.MRN != "MRN-12345" {
		t.Errorf("Patient.MRN = %q, want %q", p.MRN, "MRN-12345")
	}
	if p.City != "Springfield" {
		t.Errorf("Patient.City = %q, want %q", p.City, "Springfield")
	}
	if p.State != "IL" {
		t.Errorf("Patient.State = %q, want %q", p.State, "IL")
	}
	if p.Zip != "62701" {
		t.Errorf("Patient.Zip = %q, want %q", p.Zip, "62701")
	}
	if p.Phone == "" {
		t.Error("Patient.Phone should not be empty")
	}
}

func TestParseCCD_Allergies(t *testing.T) {
	xmlData := buildCCDFixture()

	parsed, err := ParseCCD(xmlData)
	if err != nil {
		t.Fatalf("ParseCCD() error = %v", err)
	}

	if len(parsed.Allergies) != 2 {
		t.Fatalf("Allergies count = %d, want 2", len(parsed.Allergies))
	}

	a0 := parsed.Allergies[0]
	if a0.Substance != "Allergy to penicillin" {
		t.Errorf("Allergies[0].Substance = %q, want %q", a0.Substance, "Allergy to penicillin")
	}
	if a0.Status != "active" {
		t.Errorf("Allergies[0].Status = %q, want %q", a0.Status, "active")
	}
	if a0.Reaction != "Hives" {
		t.Errorf("Allergies[0].Reaction = %q, want %q", a0.Reaction, "Hives")
	}
	if a0.OnsetDate != "20200101" {
		t.Errorf("Allergies[0].OnsetDate = %q, want %q", a0.OnsetDate, "20200101")
	}

	a1 := parsed.Allergies[1]
	if a1.Substance != "Allergy to sulfonamides" {
		t.Errorf("Allergies[1].Substance = %q, want %q", a1.Substance, "Allergy to sulfonamides")
	}
}

func TestParseCCD_Medications(t *testing.T) {
	xmlData := buildCCDFixture()

	parsed, err := ParseCCD(xmlData)
	if err != nil {
		t.Fatalf("ParseCCD() error = %v", err)
	}

	if len(parsed.Medications) != 3 {
		t.Fatalf("Medications count = %d, want 3", len(parsed.Medications))
	}

	m0 := parsed.Medications[0]
	if m0.DrugName != "Lisinopril 10mg Tablet" {
		t.Errorf("Medications[0].DrugName = %q, want %q", m0.DrugName, "Lisinopril 10mg Tablet")
	}
	if m0.Status != "active" {
		t.Errorf("Medications[0].Status = %q, want %q", m0.Status, "active")
	}
	if m0.StartDate != "20240101" {
		t.Errorf("Medications[0].StartDate = %q, want %q", m0.StartDate, "20240101")
	}

	m1 := parsed.Medications[1]
	if m1.DrugName != "Metformin 500mg Tablet" {
		t.Errorf("Medications[1].DrugName = %q, want %q", m1.DrugName, "Metformin 500mg Tablet")
	}

	// Completed medication should have end date
	m2 := parsed.Medications[2]
	if m2.DrugName != "Prednisone 5mg Tablet" {
		t.Errorf("Medications[2].DrugName = %q, want %q", m2.DrugName, "Prednisone 5mg Tablet")
	}
	if m2.Status != "completed" {
		t.Errorf("Medications[2].Status = %q, want %q", m2.Status, "completed")
	}
	if m2.EndDate != "20231231" {
		t.Errorf("Medications[2].EndDate = %q, want %q", m2.EndDate, "20231231")
	}
}

func TestParseCCD_Problems(t *testing.T) {
	xmlData := buildCCDFixture()

	parsed, err := ParseCCD(xmlData)
	if err != nil {
		t.Fatalf("ParseCCD() error = %v", err)
	}

	if len(parsed.Problems) != 3 {
		t.Fatalf("Problems count = %d, want 3", len(parsed.Problems))
	}

	p0 := parsed.Problems[0]
	if p0.Condition != "Essential hypertension" {
		t.Errorf("Problems[0].Condition = %q, want %q", p0.Condition, "Essential hypertension")
	}
	if p0.ICD10Code != "I10" {
		t.Errorf("Problems[0].ICD10Code = %q, want %q", p0.ICD10Code, "I10")
	}
	if p0.Status != "active" {
		t.Errorf("Problems[0].Status = %q, want %q", p0.Status, "active")
	}
	if p0.OnsetDate != "20200101" {
		t.Errorf("Problems[0].OnsetDate = %q, want %q", p0.OnsetDate, "20200101")
	}

	// Third problem (asthma) should be resolved
	p2 := parsed.Problems[2]
	if p2.Condition != "Unspecified asthma" {
		t.Errorf("Problems[2].Condition = %q", p2.Condition)
	}
	if p2.Status != "resolved" {
		t.Errorf("Problems[2].Status = %q, want %q, got %q", p2.Status, "resolved", p2.Status)
	}
}

func TestParseCCD_Vitals(t *testing.T) {
	xmlData := buildCCDFixture()

	parsed, err := ParseCCD(xmlData)
	if err != nil {
		t.Fatalf("ParseCCD() error = %v", err)
	}

	if len(parsed.Vitals) != 2 {
		t.Fatalf("Vitals count = %d, want 2", len(parsed.Vitals))
	}

	// First vital sign set: BP + HR
	v0 := parsed.Vitals[0]
	if v0.Systolic == nil || *v0.Systolic != 120.0 {
		t.Errorf("Vitals[0].Systolic = %v, want 120.0", ptrFloatStr(v0.Systolic))
	}
	if v0.Diastolic == nil || *v0.Diastolic != 80.0 {
		t.Errorf("Vitals[0].Diastolic = %v, want 80.0", ptrFloatStr(v0.Diastolic))
	}
	if v0.HeartRate == nil || *v0.HeartRate != 72.0 {
		t.Errorf("Vitals[0].HeartRate = %v, want 72.0", ptrFloatStr(v0.HeartRate))
	}

	// Second vital sign set: BP + Weight
	v1 := parsed.Vitals[1]
	if v1.Systolic == nil || *v1.Systolic != 118.0 {
		t.Errorf("Vitals[1].Systolic = %v, want 118.0", ptrFloatStr(v1.Systolic))
	}
	if v1.WeightKg == nil || *v1.WeightKg != 68.5 {
		t.Errorf("Vitals[1].WeightKg = %v, want 68.5", ptrFloatStr(v1.WeightKg))
	}
}

func TestParseCCD_CategoryCounts(t *testing.T) {
	xmlData := buildCCDFixture()

	parsed, err := ParseCCD(xmlData)
	if err != nil {
		t.Fatalf("ParseCCD() error = %v", err)
	}

	// Verify counts for all sections
	if len(parsed.Allergies) != 2 {
		t.Errorf("Allergies = %d, want 2", len(parsed.Allergies))
	}
	if len(parsed.Medications) != 3 {
		t.Errorf("Medications = %d, want 3", len(parsed.Medications))
	}
	if len(parsed.Problems) != 3 {
		t.Errorf("Problems = %d, want 3", len(parsed.Problems))
	}
	if len(parsed.Vitals) != 2 {
		t.Errorf("Vitals = %d, want 2", len(parsed.Vitals))
	}
	// Sections not in fixture should be empty
	if len(parsed.Immunizations) != 0 {
		t.Errorf("Immunizations = %d, want 0", len(parsed.Immunizations))
	}
	if len(parsed.Procedures) != 0 {
		t.Errorf("Procedures = %d, want 0", len(parsed.Procedures))
	}
	if len(parsed.Results) != 0 {
		t.Errorf("Results = %d, want 0", len(parsed.Results))
	}
	if len(parsed.SocialHistory) != 0 {
		t.Errorf("SocialHistory = %d, want 0", len(parsed.SocialHistory))
	}
	if len(parsed.Encounters) != 0 {
		t.Errorf("Encounters = %d, want 0", len(parsed.Encounters))
	}
}

func TestParseCCD_InvalidXML(t *testing.T) {
	_, err := ParseCCD([]byte("<not>valid xml"))
	if err == nil {
		t.Error("ParseCCD() should return error for invalid XML")
	}
}

func TestParseCCD_EmptyXML(t *testing.T) {
	_, err := ParseCCD([]byte{})
	if err == nil {
		t.Error("ParseCCD() should return error for empty XML")
	}
}

func TestParseCCD_NoPatient(t *testing.T) {
	// Document with no recordTarget
	doc := ClinicalDocument{
		Xmlns:  "urn:hl7-org:v3",
		TypeID: II{XMLName: xmlName("typeId"), Root: "2.16.840.1.113883.1.3", Extension: "POCD_HD000040"},
		ID:     II{XMLName: xmlName("id"), Root: "test", Extension: "1"},
		Code:   CE{XMLName: xmlName("code"), Code: "34133-9"},
		Title:  ST{XMLName: xmlName("title"), Text: "No Patient"},
		Component: ComponentBody{
			XMLName: xmlName("component"),
			StructuredBody: StructuredBody{
				XMLName:    xmlName("structuredBody"),
				Components: []SectionComp{},
			},
		},
	}
	data, _ := xml.MarshalIndent(doc, "", "  ")
	fullData := append([]byte(xml.Header), data...)

	parsed, err := ParseCCD(fullData)
	if err != nil {
		t.Fatalf("ParseCCD() error = %v", err)
	}
	if parsed.Patient.FirstName != "" {
		t.Error("Patient.FirstName should be empty when no recordTarget")
	}
}

// ptrFloatStr returns a string representation of a *float64 for error messages.
func ptrFloatStr(v *float64) string {
	if v == nil {
		return "nil"
	}
	return strings.TrimRight(strings.TrimRight(
		fmt.Sprintf("%.1f", *v),
		"0"), ".")
}
