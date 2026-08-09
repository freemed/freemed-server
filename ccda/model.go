// Package ccda provides C-CDA (Consolidated Clinical Document Architecture) types
// and generation for Continuity of Care Documents (CCD).
package ccda

import "encoding/xml"

// ============================================================================
// CDA Header types
// ============================================================================

// II is the HL7v3 Instance Identifier type (root + extension pair).
type II struct {
	XMLName   xml.Name
	Root      string `xml:"root,attr,omitempty"`
	Extension string `xml:"extension,attr,omitempty"`
}

// CE is the HL7v3 Coded Element type.
type CE struct {
	XMLName     xml.Name
	Code        string `xml:"code,attr,omitempty"`
	CodeSystem  string `xml:"codeSystem,attr,omitempty"`
	DisplayName string `xml:"displayName,attr,omitempty"`
}

// CS is the HL7v3 Coded Simple type.
type CS struct {
	XMLName    xml.Name
	Code       string `xml:"code,attr,omitempty"`
	CodeSystem string `xml:"codeSystem,attr,omitempty"`
}

// CD is the HL7v3 Concept Descriptor type.
type CD struct {
	XMLName     xml.Name
	Code        string `xml:"code,attr,omitempty"`
	CodeSystem  string `xml:"codeSystem,attr,omitempty"`
	DisplayName string `xml:"displayName,attr,omitempty"`
}

// TS is the HL7v3 Timestamp type.
type TS struct {
	XMLName xml.Name
	Value   string `xml:"value,attr,omitempty"`
}

// ST is the HL7v3 String type.
type ST struct {
	XMLName xml.Name
	Text    string `xml:",chardata"`
}

// AD is the HL7v3 Address type.
type AD struct {
	XMLName    xml.Name
	Use        string   `xml:"use,attr,omitempty"`
	City       string   `xml:"city,omitempty"`
	State      string   `xml:"state,omitempty"`
	PostalCode string   `xml:"postalCode,omitempty"`
	Lines      []string `xml:"streetAddressLine"`
}

// TEL is the HL7v3 Telecommunication type.
type TEL struct {
	XMLName xml.Name
	Use     string `xml:"use,attr,omitempty"`
	Value   string `xml:"value,attr,omitempty"`
}

// PN is the HL7v3 Person Name type.
type PN struct {
	XMLName xml.Name
	Use     string `xml:"use,attr,omitempty"`
	Given   string `xml:"given,omitempty"`
	Family  string `xml:"family,omitempty"`
	Suffix  string `xml:"suffix,omitempty"`
	Prefix  string `xml:"prefix,omitempty"`
}

// ON is the HL7v3 Organization Name type.
type ON struct {
	XMLName xml.Name
	Text    string `xml:",chardata"`
}

// ============================================================================
// Header participants
// ============================================================================

// RecordTarget is the patient record target.
type RecordTarget struct {
	XMLName     xml.Name         `xml:"recordTarget"`
	PatientRole RecordTargetRole `xml:"patientRole"`
}

// RecordTargetRole contains the patient role details.
type RecordTargetRole struct {
	XMLName xml.Name      `xml:"patientRole"`
	IDs     []II          `xml:"id"`
	Addr    *AD           `xml:"addr"`
	Telecom []TEL         `xml:"telecom"`
	Patient RecordTargetP `xml:"patient"`
}

// RecordTargetP is the patient within recordTarget.
type RecordTargetP struct {
	XMLName                xml.Name `xml:"patient"`
	Name                   PN       `xml:"name"`
	AdministrativeGenderCd CE       `xml:"administrativeGenderCode"`
	BirthTime              TS       `xml:"birthTime"`
}

// Author is the document author.
type Author struct {
	XMLName        xml.Name          `xml:"author"`
	Time           TS                `xml:"time"`
	AssignedAuthor AuthorAssigned    `xml:"assignedAuthor"`
}

// AuthorAssigned is the assigned author details.
type AuthorAssigned struct {
	XMLName    xml.Name       `xml:"assignedAuthor"`
	IDs        []II           `xml:"id"`
	Addr       *AD            `xml:"addr"`
	Telecom    []TEL          `xml:"telecom"`
	AssignedP  *AuthorPerson  `xml:"assignedPerson"`
}

// AuthorPerson is the person assigned as author.
type AuthorPerson struct {
	XMLName xml.Name `xml:"assignedPerson"`
	Name    PN       `xml:"name"`
}

// Custodian is the document custodian.
type Custodian struct {
	XMLName       xml.Name            `xml:"custodian"`
	AssignedCust  CustodianAssigned   `xml:"assignedCustodian"`
}

// CustodianAssigned is the assigned custodian.
type CustodianAssigned struct {
	XMLName xml.Name              `xml:"assignedCustodian"`
	RepOrg  CustodianOrganization `xml:"representedCustodianOrganization"`
}

// CustodianOrganization is the custodian organization.
type CustodianOrganization struct {
	XMLName xml.Name `xml:"representedCustodianOrganization"`
	IDs     []II     `xml:"id"`
	Name    ON       `xml:"name"`
	Telecom []TEL    `xml:"telecom"`
	Addr    *AD      `xml:"addr"`
}

// DocumentationOf is the service event documentation.
type DocumentationOf struct {
	XMLName      xml.Name          `xml:"documentationOf"`
	ServiceEvent ServiceEvent      `xml:"serviceEvent"`
}

// ServiceEvent is the service event details.
type ServiceEvent struct {
	XMLName    xml.Name              `xml:"serviceEvent"`
	Code       *CE                   `xml:"code"`
	EffTime    *TS                   `xml:"effectiveTime"`
	Performer  []ServicePerformer    `xml:"performer"`
}

// ServicePerformer is a performer on the service event.
type ServicePerformer struct {
	XMLName         xml.Name              `xml:"performer"`
	AssignedEnt     ServicePerformerEnt    `xml:"assignedEntity"`
}

// ServicePerformerEnt is the assigned entity for a performer.
type ServicePerformerEnt struct {
	XMLName xml.Name `xml:"assignedEntity"`
	IDs     []II     `xml:"id"`
}

// ============================================================================
// Section / Component types
// ============================================================================

// ComponentBody is the overall structured body component.
type ComponentBody struct {
	XMLName        xml.Name        `xml:"component"`
	StructuredBody StructuredBody  `xml:"structuredBody"`
}

// StructuredBody holds all sections.
type StructuredBody struct {
	XMLName    xml.Name        `xml:"structuredBody"`
	Components []SectionComp   `xml:"component"`
}

// SectionComp is a component wrapper for a section.
type SectionComp struct {
	XMLName xml.Name `xml:"component"`
	Section Section  `xml:"section"`
}

// Section is a CCD section (allergies, medications, etc).
type Section struct {
	XMLName     xml.Name  `xml:"section"`
	TemplateID  II        `xml:"templateId"`
	Code        CE        `xml:"code"`
	Title       ST        `xml:"title"`
	Text        SectionText `xml:"text"`
	Entries     []SectionEntry `xml:"entry"`
}

// SectionText holds the narrative text for a section.
type SectionText struct {
	XMLName xml.Name `xml:"text"`
	Content string   `xml:",innerxml"`
}

// SectionEntry wraps a clinical statement.
type SectionEntry struct {
	XMLName xml.Name       `xml:"entry"`
	Act     *ActEntry      `xml:"act,omitempty"`
	Obs     *ObsEntry      `xml:"observation,omitempty"`
	SubAdm  *SubstanceAdm  `xml:"substanceAdministration,omitempty"`
	Proc    *ProcEntry     `xml:"procedure,omitempty"`
	Org     *Organizer     `xml:"organizer,omitempty"`
	Enc     *EncounterEntry `xml:"encounter,omitempty"`
}

// ============================================================================
// Clinical statement entry types
// ============================================================================

// ActEntry is a generic act entry.
type ActEntry struct {
	XMLName     xml.Name  `xml:"act"`
	TemplateID  []II      `xml:"templateId"`
	Code        CE        `xml:"code"`
	StatusCode  CS        `xml:"statusCode"`
	EffTime     *TS       `xml:"effectiveTime"`
}

// ObsEntry is an observation entry (allergies, problems, vitals, results, social history).
type ObsEntry struct {
	XMLName    xml.Name `xml:"observation"`
	TemplateID []II     `xml:"templateId"`
	ID         *II       `xml:"id"`
	Code       CE        `xml:"code"`
	StatusCd   CS        `xml:"statusCode"`
	EffTime    *TS       `xml:"effectiveTime"`
	Value      *CD       `xml:"value"`
	TargetSite *STS      `xml:"targetSiteCode,omitempty"`
}

// SubstanceAdm is a substance administration (medications).
type SubstanceAdm struct {
	XMLName     xml.Name          `xml:"substanceAdministration"`
	TemplateID  []II              `xml:"templateId"`
	ID          *II               `xml:"id"`
	StatusCode  CS                `xml:"statusCode"`
	EffTime     *EffTimeInterval  `xml:"effectiveTime"`
	Consumable  *SubAdmConsumable `xml:"consumable"`
}

// EffTimeInterval is an effectiveTime with low/high sub-elements.
type EffTimeInterval struct {
	XMLName xml.Name `xml:"effectiveTime"`
	Low     *TS      `xml:"low"`
	High    *TS      `xml:"high"`
}

// SubAdmConsumable is the medication consumable.
type SubAdmConsumable struct {
	XMLName        xml.Name             `xml:"consumable"`
	ManufacturedPd SubAdmManufacturedPd `xml:"manufacturedProduct"`
}

// SubAdmManufacturedPd is the manufactured product.
type SubAdmManufacturedPd struct {
	XMLName    xml.Name `xml:"manufacturedProduct"`
	ManufLabel string   `xml:"manufacturedLabeledDrug>name"`
}

// ProcEntry is a procedure entry.
type ProcEntry struct {
	XMLName    xml.Name `xml:"procedure"`
	TemplateID []II     `xml:"templateId"`
	ID         *II      `xml:"id"`
	Code       *CE      `xml:"code"`
	StatusCode CS       `xml:"statusCode"`
	EffTime    *TS      `xml:"effectiveTime"`
}

// Organizer is an organizer (vital signs panel, etc).
type Organizer struct {
	XMLName    xml.Name    `xml:"organizer"`
	TemplateID []II        `xml:"templateId"`
	ID         *II         `xml:"id"`
	Code       CE          `xml:"code"`
	StatusCode CS          `xml:"statusCode"`
	EffTime    *TS         `xml:"effectiveTime"`
	Components []OrgComp   `xml:"component"`
}

// OrgComp wraps an observation inside an organizer.
type OrgComp struct {
	XMLName xml.Name   `xml:"component"`
	Obs     ObsEntry   `xml:"observation"`
}

// EncounterEntry is an encounter entry.
type EncounterEntry struct {
	XMLName    xml.Name  `xml:"encounter"`
	TemplateID []II      `xml:"templateId"`
	ID         *II       `xml:"id"`
	Code       *CE       `xml:"code"`
	StatusCode CS        `xml:"statusCode"`
	EffTime    *TS       `xml:"effectiveTime"`
}

// STS is a site code (for immunizations body site).
type STS struct {
	XMLName    xml.Name `xml:"targetSiteCode"`
	Code       string   `xml:"code,attr,omitempty"`
	CodeSystem string   `xml:"codeSystem,attr,omitempty"`
}

// ============================================================================
// Root ClinicalDocument
// ============================================================================

// ClinicalDocument is the root C-CDA element.
type ClinicalDocument struct {
	XMLName             xml.Name              `xml:"ClinicalDocument"`
	Xmlns               string                `xml:"xmlns,attr"`
	RealmCode           CS                    `xml:"realmCode"`
	TypeID              II                    `xml:"typeId"`
	TemplateIDs         []II                  `xml:"templateId"`
	ID                  II                    `xml:"id"`
	Code                CE                    `xml:"code"`
	Title               ST                    `xml:"title"`
	EffectiveTime       TS                    `xml:"effectiveTime"`
	ConfidentialityCode CS                    `xml:"confidentialityCode"`
	LanguageCode        CS                    `xml:"languageCode"`
	SetID               *II                   `xml:"setId"`
	VersionNumber       *INT                  `xml:"versionNumber"`
	RecordTargets       []RecordTarget        `xml:"recordTarget"`
	Authors             []Author              `xml:"author"`
	Custodian           Custodian             `xml:"custodian"`
	DocumentationOf     *DocumentationOf      `xml:"documentationOf"`
	Component           ComponentBody         `xml:"component"`
}

// INT is the HL7v3 integer type.
type INT struct {
	XMLName xml.Name `xml:"versionNumber"`
	Value   int      `xml:"value,attr"`
}
