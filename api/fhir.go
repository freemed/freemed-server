package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["fhir"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/metadata", fhirCapabilityStatement)
			r.GET("/Patient/:id", fhirPatientGet)
			r.GET("/Observation", fhirObservationList)
		},
	}
}

// fhirContentType sets the FHIR JSON MIME type on the response.
func fhirContentType(c *gin.Context) {
	c.Header("Content-Type", "application/fhir+json; charset=utf-8")
}

// ============================================================================
// OperationOutcome — FHIR error response
// ============================================================================

type fhirOperationOutcomeIssue struct {
	Severity    string `json:"severity"`
	Code        string `json:"code"`
	Diagnostics string `json:"diagnostics"`
}

type fhirOperationOutcome struct {
	ResourceType string                       `json:"resourceType"`
	Issue        []fhirOperationOutcomeIssue  `json:"issue"`
}

func fhirError(c *gin.Context, httpStatus int, severity, code, msg string) {
	fhirContentType(c)
	c.AbortWithStatusJSON(httpStatus, fhirOperationOutcome{
		ResourceType: "OperationOutcome",
		Issue: []fhirOperationOutcomeIssue{
			{Severity: severity, Code: code, Diagnostics: msg},
		},
	})
}

// ============================================================================
// FHIR base types
// ============================================================================

type fhirMeta struct {
	VersionID   string   `json:"versionId,omitempty"`
	LastUpdated string   `json:"lastUpdated,omitempty"`
	Profile     []string `json:"profile,omitempty"`
}

type fhirNarrative struct {
	Status string `json:"status"`
	Div    string `json:"div"`
}

type fhirIdentifier struct {
	Use    string `json:"use,omitempty"`
	System string `json:"system"`
	Value  string `json:"value"`
}

type fhirHumanName struct {
	Use    string   `json:"use,omitempty"`
	Family string   `json:"family"`
	Given  []string `json:"given,omitempty"`
	Prefix []string `json:"prefix,omitempty"`
	Suffix []string `json:"suffix,omitempty"`
}

type fhirCoding struct {
	System  string `json:"system,omitempty"`
	Code    string `json:"code,omitempty"`
	Display string `json:"display,omitempty"`
}

type fhirCodeableConcept struct {
	Coding []fhirCoding `json:"coding,omitempty"`
	Text   string       `json:"text,omitempty"`
}

type fhirReference struct {
	Reference string `json:"reference,omitempty"`
	Display   string `json:"display,omitempty"`
}

type fhirQuantity struct {
	Value  float64 `json:"value"`
	Unit   string  `json:"unit,omitempty"`
	System string  `json:"system,omitempty"`
	Code   string  `json:"code,omitempty"`
}

type fhirAddress struct {
	Use        string   `json:"use,omitempty"`
	Line       []string `json:"line,omitempty"`
	City       string   `json:"city,omitempty"`
	State      string   `json:"state,omitempty"`
	PostalCode string   `json:"postalCode,omitempty"`
}

type fhirContactPoint struct {
	System string `json:"system"`
	Value  string `json:"value"`
	Use    string `json:"use,omitempty"`
}

type fhirObservationComponent struct {
	Code          fhirCodeableConcept `json:"code"`
	ValueQuantity *fhirQuantity       `json:"valueQuantity,omitempty"`
}

// ============================================================================
// CapabilityStatement (GET /api/fhir/metadata)
// ============================================================================

func fhirCapabilityStatement(c *gin.Context) {
	fhirContentType(c)
	c.JSON(http.StatusOK, gin.H{
		"resourceType": "CapabilityStatement",
		"status":       "active",
		"date":         "2026-08-07",
		"kind":         "instance",
		"fhirVersion":  "4.0.1",
		"format":       []string{"application/fhir+json"},
		"rest": []gin.H{{
			"mode": "server",
			"resource": []gin.H{
				{
					"type":       "Patient",
					"profile":    "http://hl7.org/fhir/StructureDefinition/Patient",
					"interaction": []gin.H{
						{"code": "read"},
					},
				},
				{
					"type":       "Observation",
					"profile":    "http://hl7.org/fhir/StructureDefinition/vitalsigns",
					"interaction": []gin.H{
						{"code": "search-type"},
					},
					"searchParam": []gin.H{
						{"name": "patient", "type": "reference"},
						{"name": "category", "type": "token"},
						{"name": "_count", "type": "number"},
					},
				},
			},
		}},
	})
}

// ============================================================================
// Patient resource (GET /api/fhir/Patient/:id)
// ============================================================================

type fhirPatient struct {
	ResourceType        string              `json:"resourceType"`
	ID                  string              `json:"id"`
	Meta                fhirMeta            `json:"meta"`
	Text                *fhirNarrative      `json:"text,omitempty"`
	Identifier          []fhirIdentifier    `json:"identifier,omitempty"`
	Name                []fhirHumanName     `json:"name,omitempty"`
	Gender              string              `json:"gender,omitempty"`
	BirthDate           string              `json:"birthDate,omitempty"`
	DeceasedBoolean     *bool               `json:"deceasedBoolean,omitempty"`
	DeceasedDateTime    *string             `json:"deceasedDateTime,omitempty"`
	Address             []fhirAddress       `json:"address,omitempty"`
	Telecom             []fhirContactPoint  `json:"telecom,omitempty"`
	GeneralPractitioner []fhirReference     `json:"generalPractitioner,omitempty"`
}

func fhirPatientGet(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		fhirError(c, http.StatusBadRequest, "error", "required", "Patient ID is required")
		return
	}

	patientID := common.ParseInt(id)
	if patientID == 0 {
		fhirError(c, http.StatusBadRequest, "error", "value", "Invalid patient ID")
		return
	}

	row, err := model.Queries.FhirPatientById(c.Request.Context(), patientID)
	if err != nil {
		if err == sql.ErrNoRows {
			fhirError(c, http.StatusNotFound, "error", "not-found",
				fmt.Sprintf("Patient/%d not found", patientID))
			return
		}
		log.Printf("fhirPatientGet: %v", err)
		fhirError(c, http.StatusInternalServerError, "error", "exception",
			"Internal server error")
		return
	}

	// Build the FHIR Patient resource
	res := fhirPatient{
		ResourceType: "Patient",
		ID:           strconv.FormatInt(row.ID, 10),
		Meta: fhirMeta{
			VersionID:   strconv.FormatInt(row.UpdatedAt.Unix(), 10),
			LastUpdated: row.UpdatedAt.Format(time.RFC3339),
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/Patient"},
		},
	}

	// Narrative — auto-generated human-readable summary
	nameDisplay := strings.TrimSpace(row.Ptlname)
	if row.Ptfname != "" {
		nameDisplay = row.Ptfname + " " + nameDisplay
	}
	res.Text = &fhirNarrative{
		Status: "generated",
		Div: fmt.Sprintf(
			`<div xmlns="http://www.w3.org/1999/xhtml">Patient %s (MRN: %s)</div>`,
			nameDisplay, row.Ptid),
	}

	// Identifiers: MRN and optionally SSN
	res.Identifier = []fhirIdentifier{
		{
			Use:    "usual",
			System: "http://freemed.local/fhir/identifier/mrn",
			Value:  row.Ptid,
		},
	}
	if row.Ssn.Valid && row.Ssn.String != "" {
		res.Identifier = append(res.Identifier, fhirIdentifier{
			System: "http://hl7.org/fhir/sid/us-ssn",
			Value:  row.Ssn.String,
		})
	}

	// Human name
	name := fhirHumanName{Use: "official", Family: row.Ptlname, Given: make([]string, 0)}
	if row.Ptsalut != "" {
		name.Prefix = []string{row.Ptsalut}
	}
	if row.Ptfname != "" {
		name.Given = append(name.Given, row.Ptfname)
	}
	if row.Ptmname.Valid && row.Ptmname.String != "" {
		name.Given = append(name.Given, row.Ptmname.String)
	}
	if row.Ptsuffix != "" {
		name.Suffix = []string{row.Ptsuffix}
	}
	res.Name = []fhirHumanName{name}

	// Gender
	res.Gender = mapPtsexToFhirGender(row.Ptsex)

	// Birth date
	if row.Ptdob.Valid {
		res.BirthDate = row.Ptdob.Time.Format("2006-01-02")
	}

	// Deceased
	if row.Ptdead != 0 {
		if row.Ptdeaddt.Valid {
			dt := row.Ptdeaddt.Time.Format("2006-01-02")
			res.DeceasedDateTime = &dt
		} else {
			t := true
			res.DeceasedBoolean = &t
		}
	}

	// Address from patient_address join
	if row.AddressLine1.Valid && row.AddressLine1.String != "" {
		addr := fhirAddress{Use: "home"}
		addr.Line = append(addr.Line, row.AddressLine1.String)
		if row.AddressLine2.Valid && row.AddressLine2.String != "" {
			addr.Line = append(addr.Line, row.AddressLine2.String)
		}
		if row.AddressCity.Valid {
			addr.City = row.AddressCity.String
		}
		if row.AddressState.Valid {
			addr.State = row.AddressState.String
		}
		if row.AddressPostal.Valid {
			addr.PostalCode = row.AddressPostal.String
		}
		res.Address = []fhirAddress{addr}
	}

	// Telecom: email only (phone is in separate phone table)
	if row.Pemail.Valid && row.Pemail.String != "" {
		res.Telecom = []fhirContactPoint{
			{System: "email", Value: row.Pemail.String, Use: "home"},
		}
	}

	// General practitioner (PCP reference)
	if row.Ptpcp > 0 {
		res.GeneralPractitioner = []fhirReference{
			{Reference: fmt.Sprintf("Practitioner/%d", row.Ptpcp)},
		}
	}

	fhirContentType(c)
	c.JSON(http.StatusOK, res)
}

func mapPtsexToFhirGender(ptsex string) string {
	switch strings.ToLower(strings.TrimSpace(ptsex)) {
	case "m", "male":
		return "male"
	case "f", "female":
		return "female"
	case "o", "other":
		return "other"
	default:
		return "unknown"
	}
}

// ============================================================================
// Observation resource (vital signs) — GET /api/fhir/Observation
// ============================================================================

type fhirObservation struct {
	ResourceType      string                       `json:"resourceType"`
	ID                string                       `json:"id"`
	Meta              fhirMeta                     `json:"meta"`
	Text              *fhirNarrative               `json:"text,omitempty"`
	Status            string                       `json:"status"`
	Category          []fhirCodeableConcept        `json:"category,omitempty"`
	Code              fhirCodeableConcept          `json:"code"`
	Subject           *fhirReference               `json:"subject,omitempty"`
	Performer         []fhirReference              `json:"performer,omitempty"`
	EffectiveDateTime string                       `json:"effectiveDateTime,omitempty"`
	Component         []fhirObservationComponent   `json:"component,omitempty"`
}

func fhirObservationList(c *gin.Context) {
	patientParam := c.Query("patient")

	var rows []dbgen.FhirVitalsByPatientRow
	var err error

	if patientParam != "" {
		patientID := common.ParseInt(patientParam)
		if patientID == 0 {
			fhirError(c, http.StatusBadRequest, "error", "value", "Invalid patient parameter")
			return
		}
		rows, err = model.Queries.FhirVitalsByPatient(c.Request.Context(), patientID)
	} else {
		allRows, err2 := model.Queries.FhirVitalsAll(c.Request.Context(), dbgen.FhirVitalsAllParams{
			PatientID: sql.NullInt64{},
		})
		err = err2
		for _, r := range allRows {
			rows = append(rows, dbgen.FhirVitalsByPatientRow{
				ID:               r.ID,
				Patient:          r.Patient,
				DateTaken:        r.DateTaken,
				Systolic:         r.Systolic,
				Diastolic:        r.Diastolic,
				HeartRate:        r.HeartRate,
				RespiratoryRate:  r.RespiratoryRate,
				Temperature:      r.Temperature,
				OxygenSaturation: r.OxygenSaturation,
				HeightCm:         r.HeightCm,
				WeightKg:         r.WeightKg,
				Bmi:              r.Bmi,
				Notes:            r.Notes,
				CreatedAt:        r.CreatedAt,
				UpdatedAt:        r.UpdatedAt,
			})
		}
	}

	if err != nil {
		log.Printf("fhirObservationList: %v", err)
		fhirError(c, http.StatusInternalServerError, "error", "exception",
			"Internal server error querying observations")
		return
	}

	if rows == nil {
		rows = make([]dbgen.FhirVitalsByPatientRow, 0)
	}

	// Build Observation resources
	entries := make([]fhirBundleEntry, 0, len(rows))
	for _, v := range rows {
		obs := buildFhirObservationFromVitals(v)
		entries = append(entries, fhirBundleEntry{
			FullURL:  fmt.Sprintf("urn:uuid:%d", v.ID),
			Resource: obs,
		})
	}

	// Build searchset Bundle
	bundle := fhirBundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Total:        len(entries),
		Timestamp:    time.Now().Format(time.RFC3339),
		Meta: &fhirMeta{
			LastUpdated: time.Now().Format(time.RFC3339),
		},
		Entry: entries,
	}

	// Build self link
	baseURL := "http://" + c.Request.Host + "/api/fhir/Observation"
	if patientParam != "" {
		baseURL += "?patient=" + patientParam
	}
	bundle.Link = []fhirBundleLink{
		{Relation: "self", URL: baseURL},
	}

	fhirContentType(c)
	c.JSON(http.StatusOK, bundle)
}

type fhirBundle struct {
	ResourceType string            `json:"resourceType"`
	Type         string            `json:"type"`
	Total        int               `json:"total"`
	Meta         *fhirMeta         `json:"meta,omitempty"`
	Timestamp    string            `json:"timestamp,omitempty"`
	Link         []fhirBundleLink  `json:"link,omitempty"`
	Entry        []fhirBundleEntry `json:"entry,omitempty"`
}

type fhirBundleLink struct {
	Relation string `json:"relation"`
	URL      string `json:"url"`
}

type fhirBundleEntry struct {
	FullURL  string      `json:"fullUrl,omitempty"`
	Resource interface{} `json:"resource"`
}

// buildFhirObservationFromVitals constructs a single FHIR Observation (vital
// signs panel) from a vitals database row.
func buildFhirObservationFromVitals(v dbgen.FhirVitalsByPatientRow) fhirObservation {
	obs := fhirObservation{
		ResourceType: "Observation",
		ID:           strconv.FormatInt(v.ID, 10),
		Meta: fhirMeta{
			VersionID:   strconv.FormatInt(v.UpdatedAt.Unix(), 10),
			LastUpdated: v.UpdatedAt.Format(time.RFC3339),
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/vitalsigns"},
		},
		Status: "final",
		Category: []fhirCodeableConcept{{
			Coding: []fhirCoding{{
				System:  "http://terminology.hl7.org/CodeSystem/observation-category",
				Code:    "vital-signs",
				Display: "Vital Signs",
			}},
		}},
		Code: fhirCodeableConcept{
			Coding: []fhirCoding{{
				System:  "http://loinc.org",
				Code:    "85353-1",
				Display: "Vital signs panel",
			}},
			Text: "Vital Signs Panel",
		},
		Subject: &fhirReference{
			Reference: fmt.Sprintf("Patient/%d", v.Patient),
		},
		EffectiveDateTime: v.DateTaken.Format(time.RFC3339),
		Component:         make([]fhirObservationComponent, 0),
	}

	// Narrative
	obs.Text = &fhirNarrative{
		Status: "generated",
		Div: fmt.Sprintf(
			`<div xmlns="http://www.w3.org/1999/xhtml">Vital signs taken %s</div>`,
			v.DateTaken.Format("2006-01-02 15:04")),
	}

	// Systolic BP — LOINC 8480-6, UCUM mm[Hg]
	if v.Systolic.Valid {
		obs.Component = append(obs.Component, fhirObservationComponent{
			Code: fhirCodeableConcept{
				Coding: []fhirCoding{
					{System: "http://loinc.org", Code: "8480-6", Display: "Systolic blood pressure"},
				},
				Text: "Systolic Blood Pressure",
			},
			ValueQuantity: &fhirQuantity{
				Value:  float64(v.Systolic.Int32),
				Unit:   "mm[Hg]",
				System: "http://unitsofmeasure.org",
				Code:   "mm[Hg]",
			},
		})
	}

	// Diastolic BP — LOINC 8462-4, UCUM mm[Hg]
	if v.Diastolic.Valid {
		obs.Component = append(obs.Component, fhirObservationComponent{
			Code: fhirCodeableConcept{
				Coding: []fhirCoding{
					{System: "http://loinc.org", Code: "8462-4", Display: "Diastolic blood pressure"},
				},
				Text: "Diastolic Blood Pressure",
			},
			ValueQuantity: &fhirQuantity{
				Value:  float64(v.Diastolic.Int32),
				Unit:   "mm[Hg]",
				System: "http://unitsofmeasure.org",
				Code:   "mm[Hg]",
			},
		})
	}

	// Heart rate — LOINC 8867-4, UCUM {beats}/min
	if v.HeartRate.Valid {
		obs.Component = append(obs.Component, fhirObservationComponent{
			Code: fhirCodeableConcept{
				Coding: []fhirCoding{
					{System: "http://loinc.org", Code: "8867-4", Display: "Heart rate"},
				},
				Text: "Heart Rate",
			},
			ValueQuantity: &fhirQuantity{
				Value:  float64(v.HeartRate.Int32),
				Unit:   "beats/min",
				System: "http://unitsofmeasure.org",
				Code:   "{beats}/min",
			},
		})
	}

	// Respiratory rate — LOINC 9279-1, UCUM {breaths}/min
	if v.RespiratoryRate.Valid {
		obs.Component = append(obs.Component, fhirObservationComponent{
			Code: fhirCodeableConcept{
				Coding: []fhirCoding{
					{System: "http://loinc.org", Code: "9279-1", Display: "Respiratory rate"},
				},
				Text: "Respiratory Rate",
			},
			ValueQuantity: &fhirQuantity{
				Value:  float64(v.RespiratoryRate.Int32),
				Unit:   "breaths/min",
				System: "http://unitsofmeasure.org",
				Code:   "{breaths}/min",
			},
		})
	}

	// Temperature — LOINC 8310-5, UCUM Cel
	if v.Temperature.Valid && v.Temperature.String != "" {
		if tempVal, err := strconv.ParseFloat(v.Temperature.String, 64); err == nil {
			obs.Component = append(obs.Component, fhirObservationComponent{
				Code: fhirCodeableConcept{
					Coding: []fhirCoding{
						{System: "http://loinc.org", Code: "8310-5", Display: "Body temperature"},
					},
					Text: "Body Temperature",
				},
				ValueQuantity: &fhirQuantity{
					Value:  tempVal,
					Unit:   "°C",
					System: "http://unitsofmeasure.org",
					Code:   "Cel",
				},
			})
		}
	}

	// Oxygen saturation — LOINC 2710-2, UCUM %
	if v.OxygenSaturation.Valid {
		obs.Component = append(obs.Component, fhirObservationComponent{
			Code: fhirCodeableConcept{
				Coding: []fhirCoding{
					{System: "http://loinc.org", Code: "2710-2", Display: "Oxygen saturation in Arterial blood"},
				},
				Text: "Oxygen Saturation",
			},
			ValueQuantity: &fhirQuantity{
				Value:  float64(v.OxygenSaturation.Int32),
				Unit:   "%",
				System: "http://unitsofmeasure.org",
				Code:   "%",
			},
		})
	}

	// Height — LOINC 8302-2, UCUM cm
	if v.HeightCm.Valid && v.HeightCm.String != "" {
		if h, err := strconv.ParseFloat(v.HeightCm.String, 64); err == nil {
			obs.Component = append(obs.Component, fhirObservationComponent{
				Code: fhirCodeableConcept{
					Coding: []fhirCoding{
						{System: "http://loinc.org", Code: "8302-2", Display: "Body height"},
					},
					Text: "Body Height",
				},
				ValueQuantity: &fhirQuantity{
					Value:  h,
					Unit:   "cm",
					System: "http://unitsofmeasure.org",
					Code:   "cm",
				},
			})
		}
	}

	// Weight — LOINC 29463-7, UCUM kg
	if v.WeightKg.Valid && v.WeightKg.String != "" {
		if w, err := strconv.ParseFloat(v.WeightKg.String, 64); err == nil {
			obs.Component = append(obs.Component, fhirObservationComponent{
				Code: fhirCodeableConcept{
					Coding: []fhirCoding{
						{System: "http://loinc.org", Code: "29463-7", Display: "Body weight"},
					},
					Text: "Body Weight",
				},
				ValueQuantity: &fhirQuantity{
					Value:  w,
					Unit:   "kg",
					System: "http://unitsofmeasure.org",
					Code:   "kg",
				},
			})
		}
	}

	// BMI — LOINC 39156-5, UCUM kg/m2
	if v.Bmi.Valid && v.Bmi.String != "" {
		if b, err := strconv.ParseFloat(v.Bmi.String, 64); err == nil {
			obs.Component = append(obs.Component, fhirObservationComponent{
				Code: fhirCodeableConcept{
					Coding: []fhirCoding{
						{System: "http://loinc.org", Code: "39156-5", Display: "Body mass index (BMI)"},
					},
					Text: "Body Mass Index",
				},
				ValueQuantity: &fhirQuantity{
					Value:  b,
					Unit:   "kg/m2",
					System: "http://unitsofmeasure.org",
					Code:   "kg/m2",
				},
			})
		}
	}

	return obs
}
