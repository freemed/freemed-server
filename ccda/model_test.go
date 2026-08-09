package ccda

import (
	"encoding/xml"
	"strings"
	"testing"
)

// TestClinicalDocument_MarshalValidXML verifies that a minimal ClinicalDocument
// marshals to well-formed XML with the correct root element.
func TestClinicalDocument_MarshalValidXML(t *testing.T) {
	doc := ClinicalDocument{
		Xmlns:     "urn:hl7-org:v3",
		RealmCode: CS{XMLName: xmlName("realmCode"), Code: "US"},
		TypeID:    II{XMLName: xmlName("typeId"), Root: "2.16.840.1.113883.1.3", Extension: "POCD_HD000040"},
		TemplateIDs: []II{
			{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.1.1"},
		},
		ID:    II{XMLName: xmlName("id"), Root: "test-root", Extension: "1"},
		Code:  CE{XMLName: xmlName("code"), Code: "34133-9", CodeSystem: "2.16.840.1.113883.6.1", DisplayName: "Summarization of episode note"},
		Title: ST{XMLName: xmlName("title"), Text: "Test CCD"},
		EffectiveTime: TS{XMLName: xmlName("effectiveTime"), Value: "20260101000000-0500"},
		ConfidentialityCode: CS{XMLName: xmlName("confidentialityCode"), Code: "N", CodeSystem: "2.16.840.1.113883.5.25"},
		LanguageCode:        CS{XMLName: xmlName("languageCode"), Code: "en-US"},
		RecordTargets: []RecordTarget{
			{
				XMLName: xmlName("recordTarget"),
				PatientRole: RecordTargetRole{
					XMLName: xmlName("patientRole"),
					IDs:     []II{{XMLName: xmlName("id"), Root: "2.16.840.1.113883.19.5.99999.1", Extension: "PT-001"}},
					Patient: RecordTargetP{
						XMLName:                xmlName("patient"),
						Name:                   PN{XMLName: xmlName("name"), Family: "Doe", Given: "John"},
						AdministrativeGenderCd: CE{XMLName: xmlName("administrativeGenderCode"), Code: "M", CodeSystem: "2.16.840.1.113883.5.1", DisplayName: "Male"},
						BirthTime:              TS{XMLName: xmlName("birthTime"), Value: "19800101"},
					},
				},
			},
		},
		Authors: []Author{
			{
				XMLName: xmlName("author"),
				Time:    TS{XMLName: xmlName("time"), Value: "20260101000000-0500"},
				AssignedAuthor: AuthorAssigned{
					XMLName:   xmlName("assignedAuthor"),
					AssignedP: &AuthorPerson{XMLName: xmlName("assignedPerson"), Name: PN{XMLName: xmlName("name"), Family: "System", Given: "FreeMED"}},
				},
			},
		},
		Custodian: Custodian{
			XMLName: xmlName("custodian"),
			AssignedCust: CustodianAssigned{
				XMLName: xmlName("assignedCustodian"),
				RepOrg: CustodianOrganization{
					XMLName: xmlName("representedCustodianOrganization"),
					Name:    ON{XMLName: xmlName("name"), Text: "FreeMED EMR"},
				},
			},
		},
		DocumentationOf: &DocumentationOf{
			XMLName: xmlName("documentationOf"),
			ServiceEvent: ServiceEvent{
				XMLName: xmlName("serviceEvent"),
				EffTime: &TS{XMLName: xmlName("effectiveTime"), Value: "20260101000000-0500"},
			},
		},
		Component: ComponentBody{
			XMLName: xmlName("component"),
			StructuredBody: StructuredBody{
				XMLName:    xmlName("structuredBody"),
				Components: []SectionComp{}, // empty sections
			},
		},
	}

	output, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("xml.MarshalIndent failed: %v", err)
	}

	// Verify root element name
	if !strings.Contains(string(output), "<ClinicalDocument") {
		t.Error("output missing root ClinicalDocument element")
	}

	// Verify it's valid XML by unmarshalling back
	var doc2 ClinicalDocument
	if err := xml.Unmarshal(output, &doc2); err != nil {
		t.Fatalf("xml.Unmarshal of generated XML failed: %v", err)
	}

	// Check key fields round-tripped
	if doc2.ID.Extension != "1" {
		t.Errorf("round-trip: ID.Extension = %q, want %q", doc2.ID.Extension, "1")
	}
	if doc2.ID.Root != "test-root" {
		t.Errorf("round-trip: ID.Root = %q, want %q", doc2.ID.Root, "test-root")
	}
}

// TestClinicalDocument_EmptySections verifies that a document with no sections
// still produces valid XML.
func TestClinicalDocument_EmptySections(t *testing.T) {
	doc := ClinicalDocument{
		Xmlns:     "urn:hl7-org:v3",
		RealmCode: CS{XMLName: xmlName("realmCode"), Code: "US"},
		TypeID:    II{XMLName: xmlName("typeId"), Root: "2.16.840.1.113883.1.3", Extension: "POCD_HD000040"},
		ID:        II{XMLName: xmlName("id"), Root: "empty-test", Extension: "0"},
		Code:      CE{XMLName: xmlName("code"), Code: "34133-9", CodeSystem: "2.16.840.1.113883.6.1"},
		Title:     ST{XMLName: xmlName("title"), Text: "Empty Test"},
		EffectiveTime: TS{XMLName: xmlName("effectiveTime"), Value: "20260101"},
		ConfidentialityCode: CS{XMLName: xmlName("confidentialityCode"), Code: "N"},
		LanguageCode:        CS{XMLName: xmlName("languageCode"), Code: "en-US"},
		Component: ComponentBody{
			XMLName: xmlName("component"),
			StructuredBody: StructuredBody{
				XMLName:    xmlName("structuredBody"),
				Components: []SectionComp{},
			},
		},
	}

	output, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("xml.MarshalIndent failed for empty sections: %v", err)
	}

	// Verify it parses back
	var doc2 ClinicalDocument
	if err := xml.Unmarshal(output, &doc2); err != nil {
		t.Fatalf("xml.Unmarshal of empty-section document failed: %v", err)
	}
}

// TestCCD_RootElementName confirms the root element name is "ClinicalDocument".
func TestCCD_RootElementName(t *testing.T) {
	doc := ClinicalDocument{
		Xmlns: "urn:hl7-org:v3",
		TypeID: II{XMLName: xmlName("typeId"), Root: "2.16.840.1.113883.1.3", Extension: "POCD_HD000040"},
		ID:     II{XMLName: xmlName("id"), Root: "root-test", Extension: "42"},
		Code:   CE{XMLName: xmlName("code"), Code: "34133-9", CodeSystem: "2.16.840.1.113883.6.1"},
		Title:  ST{XMLName: xmlName("title"), Text: "Root Element Test"},
		Component: ComponentBody{
			XMLName: xmlName("component"),
			StructuredBody: StructuredBody{
				XMLName:    xmlName("structuredBody"),
				Components: []SectionComp{},
			},
		},
	}

	output, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("xml.MarshalIndent failed: %v", err)
	}

	// Root element should begin with <ClinicalDocument
	xmlStr := string(output)
	if !strings.HasPrefix(strings.TrimSpace(xmlStr), "<ClinicalDocument") {
		t.Errorf("expected root element to start with <ClinicalDocument, got: %s", xmlStr[:min(80, len(xmlStr))])
	}
}

// TestCCD_BasicTypes_Marshal verifies that basic HL7v3 types marshal correctly.
func TestCCD_BasicTypes_Marshal(t *testing.T) {
	// Test II
	ii := II{XMLName: xmlName("id"), Root: "2.16.840.1", Extension: "123"}
	data, err := xml.Marshal(ii)
	if err != nil {
		t.Fatalf("II marshal failed: %v", err)
	}
	if !strings.Contains(string(data), "2.16.840.1") {
		t.Error("II missing root attribute")
	}

	// Test CE
	ce := CE{XMLName: xmlName("code"), Code: "48765-2", CodeSystem: "2.16.840.1.113883.6.1", DisplayName: "Allergies"}
	data, err = xml.Marshal(ce)
	if err != nil {
		t.Fatalf("CE marshal failed: %v", err)
	}
	if !strings.Contains(string(data), "48765-2") {
		t.Error("CE missing code attribute")
	}
}

// TestCCD_Section_Marshal verifies section marshalling.
func TestCCD_Section_Marshal(t *testing.T) {
	section := Section{
		XMLName:    xmlName("section"),
		TemplateID: II{XMLName: xmlName("templateId"), Root: "2.16.840.1.113883.10.20.22.2.6"},
		Code:       CE{XMLName: xmlName("code"), Code: "48765-2", CodeSystem: "2.16.840.1.113883.6.1", DisplayName: "Allergies"},
		Title:      ST{XMLName: xmlName("title"), Text: "Allergies and Intolerances"},
		Text:       SectionText{XMLName: xmlName("text"), Content: "<paragraph>No known allergies</paragraph>"},
	}

	data, err := xml.MarshalIndent(section, "", "  ")
	if err != nil {
		t.Fatalf("section marshal failed: %v", err)
	}

	var s2 Section
	if err := xml.Unmarshal(data, &s2); err != nil {
		t.Fatalf("section unmarshal failed: %v", err)
	}

	if s2.Title.Text != "Allergies and Intolerances" {
		t.Errorf("Section title = %q, want %q", s2.Title.Text, "Allergies and Intolerances")
	}
}
