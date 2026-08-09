package ccda

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/freemed/freemed-server/model"
)

// NewII creates a new Instance Identifier.
func NewII(root, extension string) II {
	return II{
		XMLName:   xmlName("id"),
		Root:      root,
		Extension: extension,
	}
}

// NewCE creates a new Coded Element.
func NewCE(code, codeSystem, displayName string) CE {
	return CE{
		XMLName:     xmlName("code"),
		Code:        code,
		CodeSystem:  codeSystem,
		DisplayName: displayName,
	}
}

// NewCS creates a new Coded Simple.
func NewCS(code, codeSystem string) CS {
	return CS{
		XMLName:    xmlName("statusCode"),
		Code:       code,
		CodeSystem: codeSystem,
	}
}

// hlv3CS creates a CS without an XMLName (caller sets it via struct field).
func hlv3CS(xmlTag, code, codeSystem string) CS {
	return CS{
		XMLName:    xmlName(xmlTag),
		Code:       code,
		CodeSystem: codeSystem,
	}
}

// NewTS creates a new Timestamp with the given time in HL7 format.
func NewTS(t time.Time) TS {
	return TS{
		XMLName: xmlName("effectiveTime"),
		Value:   t.Format("20060102150405-0700"),
	}
}

// tsTag creates a TS with a specific XML tag name.
func tsTag(tag string, t time.Time) TS {
	return TS{
		XMLName: xmlName(tag),
		Value:   t.Format("20060102150405-0700"),
	}
}

func xmlName(local string) xml.Name {
	return xml.Name{Local: local}
}

// ============================================================================
// GenerateCCD builds a Continuity of Care Document for a patient.
// ============================================================================

// GenerateCCD queries all clinical tables and assembles a CCD XML document.
// Returns the marshalled XML bytes or an error.
func GenerateCCD(patientID int64) ([]byte, error) {
	db := model.SqlDb
	if db == nil {
		return nil, fmt.Errorf("model.SqlDb is nil")
	}
	ctx := context.Background()

	doc := ClinicalDocument{
		Xmlns: "urn:hl7-org:v3",
		RealmCode: CS{
			XMLName: xmlName("realmCode"),
			Code:    "US",
		},
		TypeID: II{
			XMLName:   xmlName("typeId"),
			Root:      "2.16.840.1.113883.1.3",
			Extension: "POCD_HD000040",
		},
		TemplateIDs: []II{
			{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.1.1"},
		},
		Code: CE{
			XMLName:     xmlName("code"),
			Code:        "34133-9",
			CodeSystem:  "2.16.840.1.113883.6.1",
			DisplayName: "Summarization of episode note",
		},
		Title: ST{
			XMLName: xmlName("title"),
			Text:    "Continuity of Care Document",
		},
		EffectiveTime: TS{
			XMLName: xmlName("effectiveTime"),
			Value:   time.Now().Format("20060102150405-0700"),
		},
		ConfidentialityCode: CS{
			XMLName:    xmlName("confidentialityCode"),
			Code:       "N",
			CodeSystem: "2.16.840.1.113883.5.25",
		},
		LanguageCode: CS{
			XMLName: xmlName("languageCode"),
			Code:    "en-US",
		},
	}

	// --- recordTarget (patient) ---
	patientRow, err := model.Queries.FhirPatientById(ctx, patientID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("patient %d not found", patientID)
		}
		return nil, fmt.Errorf("querying patient: %w", err)
	}

	doc.ID = II{
		XMLName:   xmlName("id"),
		Root:      BuildDocID(),
		Extension: strconv.FormatInt(patientID, 10),
	}

	// Build recordTarget
	rt := RecordTarget{
		XMLName: xmlName("recordTarget"),
		PatientRole: RecordTargetRole{
			XMLName: xmlName("patientRole"),
			IDs: []II{
				II{XMLName: xmlName("id"), Root: "2.16.840.1.113883.19.5.99999.1", Extension: patientRow.Ptid},
			},
		},
	}

	// Address
	if patientRow.AddressLine1.Valid && patientRow.AddressLine1.String != "" {
		lines := []string{patientRow.AddressLine1.String}
		if patientRow.AddressLine2.Valid && patientRow.AddressLine2.String != "" {
			lines = append(lines, patientRow.AddressLine2.String)
		}
		rt.PatientRole.Addr = &AD{
			XMLName:    xmlName("addr"),
			Use:        "HP",
			Lines:      lines,
			City:       patientRow.AddressCity.String,
			State:      patientRow.AddressState.String,
			PostalCode: patientRow.AddressPostal.String,
		}
	}

	// Telecom
	if patientRow.Pemail.Valid && patientRow.Pemail.String != "" {
		rt.PatientRole.Telecom = []TEL{
			{XMLName: xmlName("telecom"), Use: "HP", Value: "mailto:" + patientRow.Pemail.String},
		}
	}

	// Patient name
	rt.PatientRole.Patient = RecordTargetP{
		XMLName: xmlName("patient"),
		Name: PN{
			XMLName: xmlName("name"),
			Use:     "L",
			Family:  patientRow.Ptlname,
			Given:   patientRow.Ptfname,
		},
		AdministrativeGenderCd: CE{
			XMLName:     xmlName("administrativeGenderCode"),
			Code:        mapPtSexToGenderCode(patientRow.Ptsex),
			CodeSystem:  "2.16.840.1.113883.5.1",
			DisplayName: mapPtSexToGenderDisplay(patientRow.Ptsex),
		},
	}
	if patientRow.Ptsex != "" {
		rt.PatientRole.Patient.AdministrativeGenderCd = CE{
			XMLName:     xmlName("administrativeGenderCode"),
			Code:        mapPtSexToGenderCode(patientRow.Ptsex),
			CodeSystem:  "2.16.840.1.113883.5.1",
			DisplayName: mapPtSexToGenderDisplay(patientRow.Ptsex),
		}
	}
	if patientRow.Ptdob.Valid {
		rt.PatientRole.Patient.BirthTime = tsTag("birthTime", patientRow.Ptdob.Time)
	}
	doc.RecordTargets = []RecordTarget{rt}

	// --- author ---
	author := Author{
		XMLName: xmlName("author"),
		Time:    tsTag("time", time.Now()),
		AssignedAuthor: AuthorAssigned{
			XMLName: xmlName("assignedAuthor"),
			AssignedP: &AuthorPerson{
				XMLName: xmlName("assignedPerson"),
				Name: PN{
					XMLName: xmlName("name"),
					Family:  "System",
					Given:   "FreeMED",
				},
			},
		},
	}
	doc.Authors = []Author{author}

	// --- custodian ---
	doc.Custodian = Custodian{
		XMLName: xmlName("custodian"),
		AssignedCust: CustodianAssigned{
			XMLName: xmlName("assignedCustodian"),
			RepOrg: CustodianOrganization{
				XMLName: xmlName("representedCustodianOrganization"),
				Name: ON{XMLName: xmlName("name"), Text: "FreeMED EMR"},
			},
		},
	}

	// --- documentationOf ---
	doc.DocumentationOf = &DocumentationOf{
		XMLName: xmlName("documentationOf"),
		ServiceEvent: ServiceEvent{
			XMLName: xmlName("serviceEvent"),
			EffTime: &TS{XMLName: xmlName("effectiveTime"), Value: time.Now().Format("20060102150405-0700")},
		},
	}

	// --- sections ---
	sections := []SectionComp{}

	// 1. Allergies
	if allergySection, secErr := buildAllergySection(ctx, patientID); secErr == nil {
		sections = append(sections, allergySection)
	}

	// 2. Medications
	if medSection, secErr := buildMedicationSection(ctx, patientID); secErr == nil {
		sections = append(sections, medSection)
	}

	// 3. Problems
	if probSection, secErr := buildProblemSection(ctx, patientID); secErr == nil {
		sections = append(sections, probSection)
	}

	// 4. Procedures
	if procSection, secErr := buildProcedureSection(ctx, patientID); secErr == nil {
		sections = append(sections, procSection)
	}

	// 5. Immunizations
	if immSection, secErr := buildImmunizationSection(ctx, patientID); secErr == nil {
		sections = append(sections, immSection)
	}

	// 6. Vital Signs
	if vitalSection, secErr := buildVitalSignsSection(ctx, patientID); secErr == nil {
		sections = append(sections, vitalSection)
	}

	// 7. Results (labs)
	if labSection, secErr := buildResultsSection(ctx, patientID); secErr == nil {
		sections = append(sections, labSection)
	}

	// 8. Encounters
	if encSection, secErr := buildEncounterSection(ctx, patientID); secErr == nil {
		sections = append(sections, encSection)
	}

	doc.Component = ComponentBody{
		XMLName: xmlName("component"),
		StructuredBody: StructuredBody{
			XMLName:    xmlName("structuredBody"),
			Components: sections,
		},
	}

	// Marshall to XML
	output, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshalling CCD XML: %w", err)
	}

	// Add XML declaration
	result := append([]byte(xml.Header), output...)
	return result, nil
}

// ============================================================================
// Section builders
// ============================================================================

func buildAllergySection(ctx context.Context, patientID int64) (SectionComp, error) {
	rows, err := model.Queries.ListAllergies(ctx, patientID)
	if err != nil {
		return SectionComp{}, err
	}

	section := Section{
		XMLName:    xmlName("section"),
		TemplateID: II{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.2.6"},
		Code: CE{
			XMLName:     xmlName("code"),
			Code:        "48765-2",
			CodeSystem:  "2.16.840.1.113883.6.1",
			DisplayName: "Allergies and Intolerances",
		},
		Title: ST{XMLName: xmlName("title"), Text: "Allergies and Intolerances"},
	}

	if len(rows) == 0 {
		section.Text = SectionText{
			XMLName: xmlName("text"),
			Content: "<paragraph>No known allergies</paragraph>",
		}
		return SectionComp{XMLName: xmlName("component"), Section: section}, nil
	}

	var sb strings.Builder
	sb.WriteString("<table><thead><tr><th>Substance</th><th>Reaction</th><th>Status</th></tr></thead><tbody>")
	entries := make([]SectionEntry, 0, len(rows))
	for _, r := range rows {
		sb.WriteString(fmt.Sprintf("<tr><td>Allergy %d</td><td></td><td>%s</td></tr>", r.ID, r.Active))
		entries = append(entries, SectionEntry{
			XMLName: xmlName("entry"),
			Obs: &ObsEntry{
				XMLName:    xmlName("observation"),
				TemplateID: []II{{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.4.7"}},
				ID:         &II{XMLName: xmlName("id"), Root: BuildEntryID("allergy", r.ID)},
				Code:       NewCE("ASSERTION", "2.16.840.1.113883.5.4", "Assertion"),
				StatusCd:   hlv3CS("statusCode", "completed", ""),
				Value:      &CD{XMLName: xmlName("value"), Code: strconv.FormatInt(r.ID, 10), DisplayName: fmt.Sprintf("Allergy %d", r.ID)},
			},
		})
	}
	sb.WriteString("</tbody></table>")
	section.Text = SectionText{XMLName: xmlName("text"), Content: sb.String()}
	section.Entries = entries

	return SectionComp{XMLName: xmlName("component"), Section: section}, nil
}

func buildMedicationSection(ctx context.Context, patientID int64) (SectionComp, error) {
	rows, err := model.Queries.ListPrescriptions(ctx, patientID)
	if err != nil {
		return SectionComp{}, err
	}

	section := Section{
		XMLName:    xmlName("section"),
		TemplateID: II{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.2.1"},
		Code: CE{
			XMLName:     xmlName("code"),
			Code:        "10160-0",
			CodeSystem:  "2.16.840.1.113883.6.1",
			DisplayName: "History of Medication Use",
		},
		Title: ST{XMLName: xmlName("title"), Text: "Medications"},
	}

	if len(rows) == 0 {
		section.Text = SectionText{
			XMLName: xmlName("text"),
			Content: "<paragraph>No known medications</paragraph>",
		}
		return SectionComp{XMLName: xmlName("component"), Section: section}, nil
	}

	var sb strings.Builder
	sb.WriteString("<table><thead><tr><th>Medication</th><th>Dosage</th><th>Frequency</th><th>Status</th></tr></thead><tbody>")
	entries := make([]SectionEntry, 0, len(rows))
	for _, r := range rows {
		sb.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>",
			escapeXML(r.DrugName), escapeXML(r.Dosage), escapeXML(r.Frequency), r.Status))

		entry := SectionEntry{
			XMLName: xmlName("entry"),
			SubAdm: &SubstanceAdm{
				XMLName:    xmlName("substanceAdministration"),
				TemplateID: []II{{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.4.16"}},
				ID:         &II{XMLName: xmlName("id"), Root: BuildEntryID("med", r.ID)},
				StatusCode: hlv3CS("statusCode", "active", ""),
				Consumable: &SubAdmConsumable{
					XMLName: xmlName("consumable"),
					ManufacturedPd: SubAdmManufacturedPd{
						XMLName:    xmlName("manufacturedProduct"),
						ManufLabel: r.DrugName,
					},
				},
			},
		}
		if !r.DateWritten.IsZero() {
			entry.SubAdm.EffTime = &EffTimeInterval{
				XMLName: xmlName("effectiveTime"),
				Low:     &TS{XMLName: xmlName("low"), Value: r.DateWritten.Format("20060102150405-0700")},
			}
		}
		entries = append(entries, entry)
	}
	sb.WriteString("</tbody></table>")
	section.Text = SectionText{XMLName: xmlName("text"), Content: sb.String()}
	section.Entries = entries

	return SectionComp{XMLName: xmlName("component"), Section: section}, nil
}

func buildProblemSection(ctx context.Context, patientID int64) (SectionComp, error) {
	currentRows, err := model.Queries.ListCurrentProblemsByPatient(ctx, patientID)
	if err != nil {
		return SectionComp{}, err
	}
	chronicRows, err := model.Queries.ListChronicProblemsByPatient(ctx, patientID)
	if err != nil {
		return SectionComp{}, err
	}

	section := Section{
		XMLName:    xmlName("section"),
		TemplateID: II{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.2.5"},
		Code: CE{
			XMLName:     xmlName("code"),
			Code:        "11450-4",
			CodeSystem:  "2.16.840.1.113883.6.1",
			DisplayName: "Problem List",
		},
		Title: ST{XMLName: xmlName("title"), Text: "Problems"},
	}

	allRows := make([]struct {
		Problem string
		Date    time.Time
		Type    string
		ID      int64
	}, 0, len(currentRows)+len(chronicRows))

	for _, r := range currentRows {
		allRows = append(allRows, struct {
			Problem string
			Date    time.Time
			Type    string
			ID      int64
		}{Problem: r.Problem, Date: r.Date, Type: "current", ID: r.ID})
	}
	for _, r := range chronicRows {
		allRows = append(allRows, struct {
			Problem string
			Date    time.Time
			Type    string
			ID      int64
		}{Problem: r.Problem, Date: r.Date, Type: "chronic", ID: r.ID})
	}

	if len(allRows) == 0 {
		section.Text = SectionText{
			XMLName: xmlName("text"),
			Content: "<paragraph>No known problems</paragraph>",
		}
		return SectionComp{XMLName: xmlName("component"), Section: section}, nil
	}

	var sb strings.Builder
	sb.WriteString("<table><thead><tr><th>Problem</th><th>Date</th><th>Type</th></tr></thead><tbody>")
	entries := make([]SectionEntry, 0, len(allRows))
	for _, r := range allRows {
		sb.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td></tr>",
			escapeXML(r.Problem), r.Date.Format("2006-01-02"), r.Type))

		entries = append(entries, SectionEntry{
			XMLName: xmlName("entry"),
			Obs: &ObsEntry{
				XMLName:    xmlName("observation"),
				TemplateID: []II{{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.4.4"}},
				ID:         &II{XMLName: xmlName("id"), Root: BuildEntryID("problem", r.ID)},
				Code:       NewCE("55607006", "2.16.840.1.113883.6.96", "Problem"),
				StatusCd:   hlv3CS("statusCode", "completed", ""),
				Value:      &CD{XMLName: xmlName("value"), DisplayName: r.Problem},
			},
		})
	}
	sb.WriteString("</tbody></table>")
	section.Text = SectionText{XMLName: xmlName("text"), Content: sb.String()}
	section.Entries = entries

	return SectionComp{XMLName: xmlName("component"), Section: section}, nil
}

func buildProcedureSection(ctx context.Context, patientID int64) (SectionComp, error) {
	rows, err := model.Queries.ListPreviousOperationsByPatient(ctx, patientID)
	if err != nil {
		return SectionComp{}, err
	}

	section := Section{
		XMLName:    xmlName("section"),
		TemplateID: II{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.2.7"},
		Code: CE{
			XMLName:     xmlName("code"),
			Code:        "47519-4",
			CodeSystem:  "2.16.840.1.113883.6.1",
			DisplayName: "History of Procedures",
		},
		Title: ST{XMLName: xmlName("title"), Text: "Procedures"},
	}

	if len(rows) == 0 {
		section.Text = SectionText{
			XMLName: xmlName("text"),
			Content: "<paragraph>No known procedures</paragraph>",
		}
		return SectionComp{XMLName: xmlName("component"), Section: section}, nil
	}

	var sb strings.Builder
	sb.WriteString("<table><thead><tr><th>Procedure</th><th>Date</th></tr></thead><tbody>")
	entries := make([]SectionEntry, 0, len(rows))
	for _, r := range rows {
		dateStr := ""
		if r.OperationDate.Valid {
			dateStr = r.OperationDate.Time.Format("2006-01-02")
		}
		sb.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td></tr>",
			escapeXML(r.Operation), dateStr))

		entry := SectionEntry{
			XMLName: xmlName("entry"),
			Proc: &ProcEntry{
				XMLName:    xmlName("procedure"),
				TemplateID: []II{{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.4.14"}},
				ID:         &II{XMLName: xmlName("id"), Root: BuildEntryID("proc", r.ID)},
				Code:       &CE{XMLName: xmlName("code"), DisplayName: r.Operation},
				StatusCode: hlv3CS("statusCode", "completed", ""),
			},
		}
		if r.OperationDate.Valid {
			entry.Proc.EffTime = &TS{XMLName: xmlName("effectiveTime"), Value: r.OperationDate.Time.Format("20060102")}
		}
		entries = append(entries, entry)
	}
	sb.WriteString("</tbody></table>")
	section.Text = SectionText{XMLName: xmlName("text"), Content: sb.String()}
	section.Entries = entries

	return SectionComp{XMLName: xmlName("component"), Section: section}, nil
}

func buildImmunizationSection(ctx context.Context, patientID int64) (SectionComp, error) {
	rows, err := model.Queries.ListImmunizations(ctx, patientID)
	if err != nil {
		return SectionComp{}, err
	}

	section := Section{
		XMLName:    xmlName("section"),
		TemplateID: II{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.2.2"},
		Code: CE{
			XMLName:     xmlName("code"),
			Code:        "11369-6",
			CodeSystem:  "2.16.840.1.113883.6.1",
			DisplayName: "History of Immunizations",
		},
		Title: ST{XMLName: xmlName("title"), Text: "Immunizations"},
	}

	if len(rows) == 0 {
		section.Text = SectionText{
			XMLName: xmlName("text"),
			Content: "<paragraph>No known immunizations</paragraph>",
		}
		return SectionComp{XMLName: xmlName("component"), Section: section}, nil
	}

	var sb strings.Builder
	sb.WriteString("<table><thead><tr><th>Immunization</th><th>Date</th><th>Lot</th></tr></thead><tbody>")
	entries := make([]SectionEntry, 0, len(rows))
	for _, r := range rows {
		lotInfo := ""
		if r.LotNumber.Valid {
			lotInfo = r.LotNumber.String
		}
		sb.WriteString(fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%s</td></tr>",
			r.Immunization, r.Dateof.Format("2006-01-02"), lotInfo))

		entries = append(entries, SectionEntry{
			XMLName: xmlName("entry"),
			SubAdm: &SubstanceAdm{
				XMLName:    xmlName("substanceAdministration"),
				TemplateID: []II{{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.4.52"}},
				ID:         &II{XMLName: xmlName("id"), Root: BuildEntryID("imm", r.ID)},
				StatusCode: hlv3CS("statusCode", "completed", ""),
				EffTime: &EffTimeInterval{
					XMLName: xmlName("effectiveTime"),
					Low:     &TS{XMLName: xmlName("low"), Value: r.Dateof.Format("20060102")},
				},
			},
		})
	}
	sb.WriteString("</tbody></table>")
	section.Text = SectionText{XMLName: xmlName("text"), Content: sb.String()}
	section.Entries = entries

	return SectionComp{XMLName: xmlName("component"), Section: section}, nil
}

func buildVitalSignsSection(ctx context.Context, patientID int64) (SectionComp, error) {
	rows, err := model.Queries.FhirVitalsByPatient(ctx, patientID)
	if err != nil {
		return SectionComp{}, err
	}

	section := Section{
		XMLName:    xmlName("section"),
		TemplateID: II{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.2.4"},
		Code: CE{
			XMLName:     xmlName("code"),
			Code:        "8716-3",
			CodeSystem:  "2.16.840.1.113883.6.1",
			DisplayName: "Vital Signs",
		},
		Title: ST{XMLName: xmlName("title"), Text: "Vital Signs"},
	}

	if len(rows) == 0 {
		section.Text = SectionText{
			XMLName: xmlName("text"),
			Content: "<paragraph>No vital signs recorded</paragraph>",
		}
		return SectionComp{XMLName: xmlName("component"), Section: section}, nil
	}

	// Use most recent vital signs
	latest := rows[0]

	var sb strings.Builder
	sb.WriteString("<table><thead><tr><th>Measurement</th><th>Value</th><th>Unit</th></tr></thead><tbody>")

	// Write narrative table
	writeVitalRow(&sb, "Systolic BP", latest.Systolic, "mm[Hg]")
	writeVitalRow(&sb, "Diastolic BP", latest.Diastolic, "mm[Hg]")
	writeVitalRow(&sb, "Heart Rate", latest.HeartRate, "/min")
	writeVitalRow(&sb, "Respiratory Rate", latest.RespiratoryRate, "/min")
	writeVitalRow(&sb, "Temperature", latest.Temperature, "Cel")
	writeVitalRow(&sb, "O2 Saturation", latest.OxygenSaturation, "%")
	writeVitalRow(&sb, "Height", latest.HeightCm, "cm")
	writeVitalRow(&sb, "Weight", latest.WeightKg, "kg")
	writeVitalRow(&sb, "BMI", latest.Bmi, "kg/m2")
	sb.WriteString("</tbody></table>")
	section.Text = SectionText{XMLName: xmlName("text"), Content: sb.String()}

	// Build organizer with component observations
	orgComponents := make([]OrgComp, 0, 9)
	addVitalSign(&orgComponents, latest.ID, "8480-6", "Systolic BP", latest.Systolic, "mm[Hg]")
	addVitalSign(&orgComponents, latest.ID, "8462-4", "Diastolic BP", latest.Diastolic, "mm[Hg]")
	addVitalSign(&orgComponents, latest.ID, "8867-4", "Heart Rate", latest.HeartRate, "/min")
	addVitalSign(&orgComponents, latest.ID, "9279-1", "Respiratory Rate", latest.RespiratoryRate, "/min")
	addVitalSign(&orgComponents, latest.ID, "8310-5", "Temperature", latest.Temperature, "Cel")
	addVitalSign(&orgComponents, latest.ID, "2710-2", "O2 Saturation", latest.OxygenSaturation, "%")
	addVitalSign(&orgComponents, latest.ID, "8302-2", "Height", latest.HeightCm, "cm")
	addVitalSign(&orgComponents, latest.ID, "29463-7", "Weight", latest.WeightKg, "kg")
	addVitalSign(&orgComponents, latest.ID, "39156-5", "BMI", latest.Bmi, "kg/m2")

	entry := SectionEntry{
		XMLName: xmlName("entry"),
		Org: &Organizer{
			XMLName:    xmlName("organizer"),
			TemplateID: []II{{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.4.26"}},
			ID:         &II{XMLName: xmlName("id"), Root: BuildEntryID("vitals", latest.ID)},
			Code:       CE{XMLName: xmlName("code"), Code: "46680005", CodeSystem: "2.16.840.1.113883.6.96", DisplayName: "Vital signs"},
			StatusCode: hlv3CS("statusCode", "completed", ""),
			EffTime:    &TS{XMLName: xmlName("effectiveTime"), Value: latest.DateTaken.Format("20060102150405-0700")},
			Components: orgComponents,
		},
	}
	section.Entries = []SectionEntry{entry}

	return SectionComp{XMLName: xmlName("component"), Section: section}, nil
}

func writeVitalRow(sb *strings.Builder, name string, val interface{}, unit string) {
	var valStr string
	switch v := val.(type) {
	case sql.NullInt32:
		if v.Valid {
			valStr = strconv.FormatInt(int64(v.Int32), 10)
		}
	case sql.NullInt64:
		if v.Valid {
			valStr = strconv.FormatInt(v.Int64, 10)
		}
	case sql.NullFloat64:
		if v.Valid {
			valStr = fmt.Sprintf("%.1f", v.Float64)
		}
	case sql.NullString:
		if v.Valid {
			valStr = v.String
		}
	default:
		valStr = fmt.Sprintf("%v", v)
	}

	if valStr == "" || valStr == "0" {
		return
	}
	sb.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td></tr>",
		escapeXML(name), escapeXML(valStr), escapeXML(unit)))
}

func addVitalSign(comps *[]OrgComp, refID int64, code, name string, val interface{}, unit string) {
	var valStr string
	var valNum float64
	var hasVal bool

	switch v := val.(type) {
	case sql.NullInt32:
		if v.Valid && v.Int32 != 0 {
			valNum = float64(v.Int32)
			valStr = strconv.FormatInt(int64(v.Int32), 10)
			hasVal = true
		}
	case sql.NullInt64:
		if v.Valid && v.Int64 != 0 {
			valNum = float64(v.Int64)
			valStr = strconv.FormatInt(v.Int64, 10)
			hasVal = true
		}
	case sql.NullFloat64:
		if v.Valid && v.Float64 != 0 {
			valNum = v.Float64
			valStr = fmt.Sprintf("%.1f", v.Float64)
			hasVal = true
		}
	case sql.NullString:
		if v.Valid && v.String != "" {
			if parsed, err := strconv.ParseFloat(v.String, 64); err == nil {
				valNum = parsed
				valStr = v.String
				hasVal = true
			}
		}
	}

	if !hasVal {
		return
	}

	ob := ObsEntry{
		XMLName:    xmlName("observation"),
		TemplateID: []II{{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.4.27"}},
		ID:         &II{XMLName: xmlName("id"), Root: BuildEntryID(fmt.Sprintf("vital-%s", code), refID)},
		Code:       CE{XMLName: xmlName("code"), Code: code, CodeSystem: "2.16.840.1.113883.6.1", DisplayName: name},
		StatusCd:   hlv3CS("statusCode", "completed", ""),
		Value: &CD{
			XMLName:     xmlName("value"),
			Code:        fmt.Sprintf("%.1f", valNum),
			CodeSystem:  "2.16.840.1.113883.6.1",
			DisplayName: fmt.Sprintf("%s %s", valStr, unit),
		},
	}
	*comps = append(*comps, OrgComp{XMLName: xmlName("component"), Obs: ob})
}

func buildResultsSection(ctx context.Context, patientID int64) (SectionComp, error) {
	rows, err := model.Queries.ListLabs(ctx, patientID)
	if err != nil {
		return SectionComp{}, err
	}

	section := Section{
		XMLName:    xmlName("section"),
		TemplateID: II{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.2.3"},
		Code: CE{
			XMLName:     xmlName("code"),
			Code:        "30954-2",
			CodeSystem:  "2.16.840.1.113883.6.1",
			DisplayName: "Relevant Diagnostic Tests And/or Laboratory Data",
		},
		Title: ST{XMLName: xmlName("title"), Text: "Results"},
	}

	if len(rows) == 0 {
		section.Text = SectionText{
			XMLName: xmlName("text"),
			Content: "<paragraph>No lab results available</paragraph>",
		}
		return SectionComp{XMLName: xmlName("component"), Section: section}, nil
	}

	var sb strings.Builder
	sb.WriteString("<table><thead><tr><th>Lab</th><th>Date</th><th>Result</th><th>Unit</th><th>Reference</th></tr></thead><tbody>")
	entries := make([]SectionEntry, 0, len(rows))
	for _, r := range rows {
		sb.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>",
			escapeXML(r.LabName), r.LabDate.Format("2006-01-02"),
			escapeXML(r.Result), escapeXML(r.Unit), escapeXML(r.ReferenceRange)))

		entries = append(entries, SectionEntry{
			XMLName: xmlName("entry"),
			Obs: &ObsEntry{
				XMLName:    xmlName("observation"),
				TemplateID: []II{{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.4.2"}},
				ID:         &II{XMLName: xmlName("id"), Root: BuildEntryID("lab", r.ID)},
				Code:       CE{XMLName: xmlName("code"), DisplayName: r.LabName},
				StatusCd:   hlv3CS("statusCode", "completed", ""),
				EffTime:    &TS{XMLName: xmlName("effectiveTime"), Value: r.LabDate.Format("20060102")},
				Value:      &CD{XMLName: xmlName("value"), DisplayName: fmt.Sprintf("%s %s", r.Result, r.Unit)},
			},
		})
	}
	sb.WriteString("</tbody></table>")
	section.Text = SectionText{XMLName: xmlName("text"), Content: sb.String()}
	section.Entries = entries

	return SectionComp{XMLName: xmlName("component"), Section: section}, nil
}

func buildEncounterSection(ctx context.Context, patientID int64) (SectionComp, error) {
	// Query procrec using raw SQL (no sqlc query exists)
	db := model.SqlDb
	if db == nil {
		return SectionComp{}, fmt.Errorf("database not available")
	}

	rows, err := db.QueryContext(ctx,
		`SELECT id, procdt, procphysician, proccpt, proccomment FROM procrec
		 WHERE procpatient = ? AND deleted_at IS NULL
		 ORDER BY procdt DESC LIMIT 20`, patientID)
	if err != nil {
		return SectionComp{}, err
	}
	defer rows.Close()

	type encRow struct {
		ID          int64
		Date        time.Time
		Physician   int64
		CPT         int64
		Comment     sql.NullString
	}

	encRows := make([]encRow, 0)
	for rows.Next() {
		var r encRow
		if err := rows.Scan(&r.ID, &r.Date, &r.Physician, &r.CPT, &r.Comment); err != nil {
			return SectionComp{}, err
		}
		encRows = append(encRows, r)
	}

	section := Section{
		XMLName:    xmlName("section"),
		TemplateID: II{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.2.22"},
		Code: CE{
			XMLName:     xmlName("code"),
			Code:        "46240-8",
			CodeSystem:  "2.16.840.1.113883.6.1",
			DisplayName: "History of Encounters",
		},
		Title: ST{XMLName: xmlName("title"), Text: "Encounters"},
	}

	if len(encRows) == 0 {
		section.Text = SectionText{
			XMLName: xmlName("text"),
			Content: "<paragraph>No encounters recorded</paragraph>",
		}
		return SectionComp{XMLName: xmlName("component"), Section: section}, nil
	}

	var sb strings.Builder
	sb.WriteString("<table><thead><tr><th>Date</th><th>Provider</th><th>CPT</th><th>Comment</th></tr></thead><tbody>")
	entries := make([]SectionEntry, 0, len(encRows))
	for _, r := range encRows {
		comment := ""
		if r.Comment.Valid {
			comment = r.Comment.String
		}
		sb.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%d</td><td>%d</td><td>%s</td></tr>",
			r.Date.Format("2006-01-02"), r.Physician, r.CPT, escapeXML(comment)))

		entries = append(entries, SectionEntry{
			XMLName: xmlName("entry"),
			Enc: &EncounterEntry{
				XMLName:    xmlName("encounter"),
				TemplateID: []II{{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.4.49"}},
				ID:         &II{XMLName: xmlName("id"), Root: BuildEntryID("enc", r.ID)},
				Code:       &CE{XMLName: xmlName("code"), DisplayName: fmt.Sprintf("CPT %d", r.CPT)},
				StatusCode: hlv3CS("statusCode", "completed", ""),
				EffTime:    &TS{XMLName: xmlName("effectiveTime"), Value: r.Date.Format("20060102")},
			},
		})
	}
	sb.WriteString("</tbody></table>")
	section.Text = SectionText{XMLName: xmlName("text"), Content: sb.String()}
	section.Entries = entries

	return SectionComp{XMLName: xmlName("component"), Section: section}, nil
}

// ============================================================================
// Helpers
// ============================================================================

// BuildDocID generates a UUID-based root for the document ID.
func BuildDocID() string {
	return "2.16.840.1.113883.19.5.99999.2"
}

// BuildEntryID generates an entry ID root.
func BuildEntryID(prefix string, id int64) string {
	return fmt.Sprintf("2.16.840.1.113883.19.5.99999.3.%s.%d", prefix, id)
}

func mapPtSexToGenderCode(ptsex string) string {
	switch strings.ToLower(strings.TrimSpace(ptsex)) {
	case "m", "male":
		return "M"
	case "f", "female":
		return "F"
	default:
		return "UN"
	}
}

func mapPtSexToGenderDisplay(ptsex string) string {
	switch strings.ToLower(strings.TrimSpace(ptsex)) {
	case "m", "male":
		return "Male"
	case "f", "female":
		return "Female"
	default:
		return "Undifferentiated"
	}
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
