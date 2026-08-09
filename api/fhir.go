package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/freemed/freemed-server/ccda"
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["fhir"] = common.ApiMapping{
		Authenticated: false, // FHIR middleware handles auth
		RouterFunction: func(r *gin.RouterGroup) {
			r.Use(FHIRAuth())
			r.GET("/metadata", fhirCapabilityStatement)
			r.GET("/Patient/:id", fhirPatientGet)
			r.GET("/Observation", fhirObservationList)
			r.GET("/Condition", fhirConditionList)
			r.GET("/Condition/:id", fhirConditionGet)
			r.GET("/AllergyIntolerance", fhirAllergyList)
			r.GET("/AllergyIntolerance/:id", fhirAllergyGet)
			r.GET("/MedicationRequest", fhirMedicationRequestList)
			r.GET("/MedicationRequest/:id", fhirMedicationRequestGet)
			r.GET("/Immunization", fhirImmunizationList)
			r.GET("/Immunization/:id", fhirImmunizationGet)
			r.GET("/Procedure", fhirProcedureList)
			r.GET("/Procedure/:id", fhirProcedureGet)
			r.GET("/Encounter", fhirEncounterList)
			r.GET("/Encounter/:id", fhirEncounterGet)
			r.GET("/FamilyMemberHistory", fhirFamilyMemberHistoryList)
			r.GET("/FamilyMemberHistory/:id", fhirFamilyMemberHistoryGet)
			r.GET("/Patient/:id/$document", fhirPatientDocument)
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
	issuer := fmt.Sprintf("http://%s", c.Request.Host)
	c.JSON(http.StatusOK, gin.H{
		"resourceType": "CapabilityStatement",
		"status":       "active",
		"date":         "2026-08-08",
		"kind":         "instance",
		"fhirVersion":  "4.0.1",
		"format":       []string{"application/fhir+json"},
		"implementation": gin.H{
			"description": "FreeMED EMR FHIR Server",
			"url":         issuer,
		},
		"rest": []gin.H{{
			"mode": "server",
			"security": map[string]interface{}{
				"service": []gin.H{{
					"coding": []gin.H{{
						"system": "http://terminology.hl7.org/CodeSystem/restful-security-service",
						"code":   "SMART-on-FHIR",
					}},
				}},
			},
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
				{
					"type":       "Condition",
					"profile":    "http://hl7.org/fhir/StructureDefinition/Condition",
					"interaction": []gin.H{
						{"code": "search-type"},
						{"code": "read"},
					},
					"searchParam": []gin.H{
						{"name": "patient", "type": "reference"},
						{"name": "_count", "type": "number"},
					},
				},
				{
					"type":       "AllergyIntolerance",
					"profile":    "http://hl7.org/fhir/StructureDefinition/AllergyIntolerance",
					"interaction": []gin.H{
						{"code": "search-type"},
						{"code": "read"},
					},
					"searchParam": []gin.H{
						{"name": "patient", "type": "reference"},
						{"name": "_count", "type": "number"},
					},
				},
				{
					"type":       "Immunization",
					"profile":    "http://hl7.org/fhir/StructureDefinition/Immunization",
					"interaction": []gin.H{
						{"code": "search-type"},
						{"code": "read"},
					},
					"searchParam": []gin.H{
						{"name": "patient", "type": "reference"},
						{"name": "_count", "type": "number"},
					},
				},
				{
					"type":       "Procedure",
					"profile":    "http://hl7.org/fhir/StructureDefinition/Procedure",
					"interaction": []gin.H{
						{"code": "search-type"},
						{"code": "read"},
					},
					"searchParam": []gin.H{
						{"name": "patient", "type": "reference"},
						{"name": "_count", "type": "number"},
					},
				},
				{
					"type":       "Encounter",
					"profile":    "http://hl7.org/fhir/StructureDefinition/Encounter",
					"interaction": []gin.H{
						{"code": "search-type"},
						{"code": "read"},
					},
					"searchParam": []gin.H{
						{"name": "patient", "type": "reference"},
						{"name": "_count", "type": "number"},
					},
				},
				{
					"type":       "MedicationRequest",
					"profile":    "http://hl7.org/fhir/StructureDefinition/MedicationRequest",
					"interaction": []gin.H{
						{"code": "search-type"},
						{"code": "read"},
					},
					"searchParam": []gin.H{
						{"name": "patient", "type": "reference"},
						{"name": "_count", "type": "number"},
					},
				},
				{
					"type":       "FamilyMemberHistory",
					"profile":    "http://hl7.org/fhir/StructureDefinition/FamilyMemberHistory",
					"interaction": []gin.H{
						{"code": "search-type"},
						{"code": "read"},
					},
					"searchParam": []gin.H{
						{"name": "patient", "type": "reference"},
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
	categoryParam := c.Query("category")

	// Social history Observations
	if categoryParam == "social-history" {
		if patientParam == "" {
			fhirError(c, http.StatusBadRequest, "error", "required",
				"patient parameter is required for social-history observations")
			return
		}
		patientID := common.ParseInt(patientParam)
		if patientID == 0 {
			fhirError(c, http.StatusBadRequest, "error", "value",
				"Invalid patient parameter")
			return
		}

		shRows, err := model.Queries.FhirSocialHistoryObservations(c.Request.Context(), patientID)
		if err != nil {
			log.Printf("fhirObservationList(social-history): %v", err)
			fhirError(c, http.StatusInternalServerError, "error", "exception",
				"Internal server error querying social history observations")
			return
		}
		if shRows == nil {
			shRows = make([]dbgen.FhirSocialHistoryObservationsRow, 0)
		}

		entries := make([]fhirBundleEntry, 0, len(shRows)*6) // up to 6 observations per row
		for _, sh := range shRows {
			entries = append(entries, buildFhirSocialHistoryObservations(sh)...)
		}

		bundle := fhirBundle{
			ResourceType: "Bundle",
			Type:         "searchset",
			Total:        len(entries),
			Timestamp:    time.Now().Format(time.RFC3339),
			Meta: &fhirMeta{
				LastUpdated: time.Now().Format(time.RFC3339),
			},
			Entry: entries,
			Link: []fhirBundleLink{
				{
					Relation: "self",
					URL:      "http://" + c.Request.Host + "/api/fhir/Observation?patient=" + patientParam + "&category=social-history",
				},
			},
		}

		fhirContentType(c)
		c.JSON(http.StatusOK, bundle)
		return
	}

	// Vital signs Observations (default)
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

// buildFhirSocialHistoryObservations converts a social_history row into a slice
// of FHIR Observation resources, one for each social history domain.
func buildFhirSocialHistoryObservations(sh dbgen.FhirSocialHistoryObservationsRow) []fhirBundleEntry {
	sharedObs := func(code fhirCoding, valueString string, detailString string) fhirObservation {
		obs := fhirObservation{
			ResourceType: "Observation",
			ID:           fmt.Sprintf("sh-%d-%s", sh.ID, code.Code),
			Meta: fhirMeta{
				VersionID:   strconv.FormatInt(sh.UpdatedAt.Unix(), 10),
				LastUpdated: sh.UpdatedAt.Format(time.RFC3339),
				Profile:     []string{"http://hl7.org/fhir/StructureDefinition/Observation"},
			},
			Status: "final",
			Category: []fhirCodeableConcept{{
				Coding: []fhirCoding{{
					System:  "http://terminology.hl7.org/CodeSystem/observation-category",
					Code:    "social-history",
					Display: "Social History",
				}},
			}},
			Code: fhirCodeableConcept{
				Coding: []fhirCoding{code},
			},
			Subject: &fhirReference{
				Reference: fmt.Sprintf("Patient/%d", sh.Patient),
			},
			EffectiveDateTime: sh.RecordedDate.Format(time.RFC3339),
		}

		// Build narrative
		obs.Text = &fhirNarrative{
			Status: "generated",
			Div: fmt.Sprintf(
				`<div xmlns="http://www.w3.org/1999/xhtml">%s: %s</div>`,
				code.Display, valueString),
		}

		return obs
	}

	var entries []fhirBundleEntry

	// Smoking status — LOINC 72166-2
	if sh.SmokingStatus != "" {
		displayText := sh.SmokingStatus
		if sh.SmokingDetail != "" {
			displayText = sh.SmokingStatus + " (" + sh.SmokingDetail + ")"
		}
		entries = append(entries, fhirBundleEntry{
			FullURL: fmt.Sprintf("urn:uuid:%d-smoking", sh.ID),
			Resource: sharedObs(
				fhirCoding{System: "http://loinc.org", Code: "72166-2", Display: "Tobacco smoking status"},
				displayText, sh.SmokingDetail,
			),
		})
	}

	// Alcohol use — LOINC 74013-4
	if sh.AlcoholUse != "" {
		displayText := sh.AlcoholUse
		if sh.AlcoholDetail != "" {
			displayText = sh.AlcoholUse + " (" + sh.AlcoholDetail + ")"
		}
		entries = append(entries, fhirBundleEntry{
			FullURL: fmt.Sprintf("urn:uuid:%d-alcohol", sh.ID),
			Resource: sharedObs(
				fhirCoding{System: "http://loinc.org", Code: "74013-4", Display: "Alcohol use"},
				displayText, sh.AlcoholDetail,
			),
		})
	}

	// Drug use — LOINC 74204-9
	if sh.DrugUse != "" {
		displayText := sh.DrugUse
		if sh.DrugDetail != "" {
			displayText = sh.DrugUse + " (" + sh.DrugDetail + ")"
		}
		entries = append(entries, fhirBundleEntry{
			FullURL: fmt.Sprintf("urn:uuid:%d-drug", sh.ID),
			Resource: sharedObs(
				fhirCoding{System: "http://loinc.org", Code: "74204-9", Display: "Drug use"},
				displayText, sh.DrugDetail,
			),
		})
	}

	// Exercise frequency — LOINC 77590-8
	if sh.ExerciseFrequency != "" {
		entries = append(entries, fhirBundleEntry{
			FullURL: fmt.Sprintf("urn:uuid:%d-exercise", sh.ID),
			Resource: sharedObs(
				fhirCoding{System: "http://loinc.org", Code: "77590-8", Display: "Frequency of moderate to vigorous aerobic physical activity"},
				sh.ExerciseFrequency, "",
			),
		})
	}

	// Occupation — LOINC 85658-3
	if sh.Occupation != "" {
		entries = append(entries, fhirBundleEntry{
			FullURL: fmt.Sprintf("urn:uuid:%d-occupation", sh.ID),
			Resource: sharedObs(
				fhirCoding{System: "http://loinc.org", Code: "85658-3", Display: "Occupation"},
				sh.Occupation, "",
			),
		})
	}

	// Living situation — SNOMED 103693007
	if sh.LivingSituation != "" {
		entries = append(entries, fhirBundleEntry{
			FullURL: fmt.Sprintf("urn:uuid:%d-living", sh.ID),
			Resource: sharedObs(
				fhirCoding{System: "http://snomed.info/sct", Code: "103693007", Display: "Living situation"},
				sh.LivingSituation, "",
			),
		})
	}

	return entries
}

// ============================================================================
// Condition resource (GET /api/fhir/Condition, GET /api/fhir/Condition/:id)
// ============================================================================

type fhirCondition struct {
	ResourceType        string                 `json:"resourceType"`
	ID                  string                 `json:"id"`
	Meta                fhirMeta               `json:"meta"`
	Text                *fhirNarrative         `json:"text,omitempty"`
	ClinicalStatus      *fhirCodeableConcept   `json:"clinicalStatus,omitempty"`
	VerificationStatus  *fhirCodeableConcept   `json:"verificationStatus,omitempty"`
	Category            []fhirCodeableConcept  `json:"category,omitempty"`
	Code                *fhirCodeableConcept   `json:"code,omitempty"`
	Subject             *fhirReference         `json:"subject,omitempty"`
	OnsetDateTime       string                 `json:"onsetDateTime,omitempty"`
	RecordedDate        string                 `json:"recordedDate,omitempty"`
}

func fhirConditionList(c *gin.Context) {
	patientParam := c.Query("patient")
	if patientParam == "" {
		fhirError(c, http.StatusBadRequest, "error", "required",
			"patient parameter is required")
		return
	}
	patientID := common.ParseInt(patientParam)
	if patientID == 0 {
		fhirError(c, http.StatusBadRequest, "error", "value",
			"Invalid patient parameter")
		return
	}

	rows, err := model.Queries.FhirConditionsByPatient(c.Request.Context(),
		dbgen.FhirConditionsByPatientParams{PatientID: patientID})
	if err != nil {
		log.Printf("fhirConditionList: %v", err)
		fhirError(c, http.StatusInternalServerError, "error", "exception",
			"Internal server error querying conditions")
		return
	}
	if rows == nil {
		rows = make([]dbgen.FhirConditionsByPatientRow, 0)
	}

	entries := make([]fhirBundleEntry, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, fhirBundleEntry{
			FullURL:  fmt.Sprintf("urn:uuid:%d", r.ConditionID),
			Resource: buildFhirCondition(r),
		})
	}

	bundle := fhirBundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Total:        len(entries),
		Timestamp:    time.Now().Format(time.RFC3339),
		Meta: &fhirMeta{
			LastUpdated: time.Now().Format(time.RFC3339),
		},
		Entry: entries,
		Link: []fhirBundleLink{
			{
				Relation: "self",
				URL:      "http://" + c.Request.Host + "/api/fhir/Condition?patient=" + patientParam,
			},
		},
	}

	fhirContentType(c)
	c.JSON(http.StatusOK, bundle)
}

func fhirConditionGet(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		fhirError(c, http.StatusBadRequest, "error", "required",
			"Condition ID is required")
		return
	}
	conditionID := common.ParseInt(id)
	if conditionID == 0 {
		fhirError(c, http.StatusBadRequest, "error", "value",
			"Invalid condition ID")
		return
	}

	row, err := model.Queries.FhirConditionById(c.Request.Context(),
		dbgen.FhirConditionByIdParams{ConditionID: conditionID})
	if err != nil {
		if err == sql.ErrNoRows {
			fhirError(c, http.StatusNotFound, "error", "not-found",
				fmt.Sprintf("Condition/%d not found", conditionID))
			return
		}
		log.Printf("fhirConditionGet: %v", err)
		fhirError(c, http.StatusInternalServerError, "error", "exception",
			"Internal server error")
		return
	}

	fhirContentType(c)
	c.JSON(http.StatusOK, buildFhirCondition(dbgen.FhirConditionsByPatientRow{
		ConditionID:      row.ConditionID,
		ConditionPatient: row.ConditionPatient,
		ConditionDate:    row.ConditionDate,
		ConditionText:    row.ConditionText,
		ConditionType:    row.ConditionType,
		ConditionActive:  row.ConditionActive,
		ConditionUpdated: row.ConditionUpdated,
	}))
}

func buildFhirCondition(r dbgen.FhirConditionsByPatientRow) fhirCondition {
	cond := fhirCondition{
		ResourceType: "Condition",
		ID:           strconv.FormatInt(r.ConditionID, 10),
		Meta: fhirMeta{
			VersionID:   strconv.FormatInt(r.ConditionUpdated.Unix(), 10),
			LastUpdated: r.ConditionUpdated.Format(time.RFC3339),
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/Condition"},
		},
		Category: []fhirCodeableConcept{{
			Coding: []fhirCoding{{
				System:  "http://terminology.hl7.org/CodeSystem/condition-category",
				Code:    "problem-list-item",
				Display: "Problem List Item",
			}},
		}},
		Code: &fhirCodeableConcept{
			Text: r.ConditionText,
		},
		Subject: &fhirReference{
			Reference: fmt.Sprintf("Patient/%d", r.ConditionPatient),
		},
		OnsetDateTime: r.ConditionDate.Format("2006-01-02"),
		RecordedDate:  r.ConditionUpdated.Format(time.RFC3339),
		Text: &fhirNarrative{
			Status: "generated",
			Div: fmt.Sprintf(
				`<div xmlns="http://www.w3.org/1999/xhtml">%s: %s</div>`,
				r.ConditionType, r.ConditionText),
		},
	}

	// Clinical status
	if r.ConditionActive == "active" {
		cond.ClinicalStatus = &fhirCodeableConcept{
			Coding: []fhirCoding{{
				System:  "http://terminology.hl7.org/CodeSystem/condition-clinical",
				Code:    "active",
				Display: "Active",
			}},
		}
	} else {
		cond.ClinicalStatus = &fhirCodeableConcept{
			Coding: []fhirCoding{{
				System:  "http://terminology.hl7.org/CodeSystem/condition-clinical",
				Code:    "resolved",
				Display: "Resolved",
			}},
		}
	}

	// Verification status — always "confirmed" for imported data
	cond.VerificationStatus = &fhirCodeableConcept{
		Coding: []fhirCoding{{
			System:  "http://terminology.hl7.org/CodeSystem/condition-ver-status",
			Code:    "confirmed",
			Display: "Confirmed",
		}},
	}

	return cond
}

// ============================================================================
// AllergyIntolerance resource (GET /api/fhir/AllergyIntolerance, GET /api/fhir/AllergyIntolerance/:id)
// ============================================================================

type fhirAllergyIntolerance struct {
	ResourceType       string               `json:"resourceType"`
	ID                 string               `json:"id"`
	Meta               fhirMeta             `json:"meta"`
	Text               *fhirNarrative       `json:"text,omitempty"`
	ClinicalStatus     *fhirCodeableConcept `json:"clinicalStatus,omitempty"`
	VerificationStatus *fhirCodeableConcept `json:"verificationStatus,omitempty"`
	Type               string               `json:"type,omitempty"`
	Category           []string             `json:"category,omitempty"`
	Code               *fhirCodeableConcept `json:"code,omitempty"`
	Patient            *fhirReference       `json:"patient,omitempty"`
	OnsetDateTime      string               `json:"onsetDateTime,omitempty"`
}

func fhirAllergyList(c *gin.Context) {
	patientParam := c.Query("patient")
	if patientParam == "" {
		fhirError(c, http.StatusBadRequest, "error", "required",
			"patient parameter is required")
		return
	}
	patientID := common.ParseInt(patientParam)
	if patientID == 0 {
		fhirError(c, http.StatusBadRequest, "error", "value",
			"Invalid patient parameter")
		return
	}

	rows, err := model.Queries.FhirAllergiesByPatient(c.Request.Context(), patientID)
	if err != nil {
		log.Printf("fhirAllergyList: %v", err)
		fhirError(c, http.StatusInternalServerError, "error", "exception",
			"Internal server error querying allergies")
		return
	}
	if rows == nil {
		rows = make([]dbgen.FhirAllergiesByPatientRow, 0)
	}

	entries := make([]fhirBundleEntry, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, fhirBundleEntry{
			FullURL:  fmt.Sprintf("urn:uuid:%d", r.ID),
			Resource: buildFhirAllergy(r),
		})
	}

	bundle := fhirBundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Total:        len(entries),
		Timestamp:    time.Now().Format(time.RFC3339),
		Meta: &fhirMeta{
			LastUpdated: time.Now().Format(time.RFC3339),
		},
		Entry: entries,
		Link: []fhirBundleLink{
			{
				Relation: "self",
				URL:      "http://" + c.Request.Host + "/api/fhir/AllergyIntolerance?patient=" + patientParam,
			},
		},
	}

	fhirContentType(c)
	c.JSON(http.StatusOK, bundle)
}

func fhirAllergyGet(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		fhirError(c, http.StatusBadRequest, "error", "required",
			"Allergy ID is required")
		return
	}
	allergyID := common.ParseInt(id)
	if allergyID == 0 {
		fhirError(c, http.StatusBadRequest, "error", "value",
			"Invalid allergy ID")
		return
	}

	row, err := model.Queries.FhirAllergyById(c.Request.Context(), allergyID)
	if err != nil {
		if err == sql.ErrNoRows {
			fhirError(c, http.StatusNotFound, "error", "not-found",
				fmt.Sprintf("AllergyIntolerance/%d not found", allergyID))
			return
		}
		log.Printf("fhirAllergyGet: %v", err)
		fhirError(c, http.StatusInternalServerError, "error", "exception",
			"Internal server error")
		return
	}

	fhirContentType(c)
	c.JSON(http.StatusOK, buildFhirAllergy(dbgen.FhirAllergiesByPatientRow{
		ID:        row.ID,
		Patient:   row.Patient,
		Active:    row.Active,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}))
}

func buildFhirAllergy(r dbgen.FhirAllergiesByPatientRow) fhirAllergyIntolerance {
	return fhirAllergyIntolerance{
		ResourceType: "AllergyIntolerance",
		ID:           strconv.FormatInt(r.ID, 10),
		Meta: fhirMeta{
			VersionID:   strconv.FormatInt(r.UpdatedAt.Unix(), 10),
			LastUpdated: r.UpdatedAt.Format(time.RFC3339),
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/AllergyIntolerance"},
		},
		Text: &fhirNarrative{
			Status: "generated",
			Div:    `<div xmlns="http://www.w3.org/1999/xhtml">Allergy record</div>`,
		},
		ClinicalStatus: &fhirCodeableConcept{
			Coding: []fhirCoding{{
				System:  "http://terminology.hl7.org/CodeSystem/allergyintolerance-clinical",
				Code:    "active",
				Display: "Active",
			}},
		},
		VerificationStatus: &fhirCodeableConcept{
			Coding: []fhirCoding{{
				System:  "http://terminology.hl7.org/CodeSystem/allergyintolerance-verification",
				Code:    "confirmed",
				Display: "Confirmed",
			}},
		},
		Type:     "allergy",
		Category: []string{"medication"},
		Code: &fhirCodeableConcept{
			Coding: []fhirCoding{{
				System:  "http://snomed.info/sct",
				Code:    "609328004",
				Display: "Allergy",
			}},
			Text: "Allergy",
		},
		Patient: &fhirReference{
			Reference: fmt.Sprintf("Patient/%d", r.Patient),
		},
		OnsetDateTime: r.CreatedAt.Format("2006-01-02"),
	}
}

// ============================================================================
// MedicationRequest resource (GET /api/fhir/MedicationRequest, GET /api/fhir/MedicationRequest/:id)
// ============================================================================

type fhirDosageInstruction struct {
	Text string `json:"text,omitempty"`
}

type fhirDispenseRequest struct {
	NumberOfRepeatsAllowed int64 `json:"numberOfRepeatsAllowed,omitempty"`
}

type fhirMedicationRequest struct {
	ResourceType             string                    `json:"resourceType"`
	ID                       string                    `json:"id"`
	Meta                     fhirMeta                  `json:"meta"`
	Text                     *fhirNarrative            `json:"text,omitempty"`
	Status                   string                    `json:"status"`
	Intent                   string                    `json:"intent"`
	MedicationCodeableConcept *fhirCodeableConcept     `json:"medicationCodeableConcept,omitempty"`
	Subject                  *fhirReference            `json:"subject,omitempty"`
	AuthoredOn               string                    `json:"authoredOn,omitempty"`
	Requester                *fhirReference            `json:"requester,omitempty"`
	DosageInstruction        []fhirDosageInstruction   `json:"dosageInstruction,omitempty"`
	DispenseRequest          *fhirDispenseRequest      `json:"dispenseRequest,omitempty"`
}

func fhirMedicationRequestList(c *gin.Context) {
	patientParam := c.Query("patient")
	if patientParam == "" {
		fhirError(c, http.StatusBadRequest, "error", "required",
			"patient parameter is required")
		return
	}
	patientID := common.ParseInt(patientParam)
	if patientID == 0 {
		fhirError(c, http.StatusBadRequest, "error", "value",
			"Invalid patient parameter")
		return
	}

	rows, err := model.Queries.FhirMedicationRequestsByPatient(c.Request.Context(), patientID)
	if err != nil {
		log.Printf("fhirMedicationRequestList: %v", err)
		fhirError(c, http.StatusInternalServerError, "error", "exception",
			"Internal server error querying medication requests")
		return
	}
	if rows == nil {
		rows = make([]dbgen.FhirMedicationRequestsByPatientRow, 0)
	}

	entries := make([]fhirBundleEntry, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, fhirBundleEntry{
			FullURL:  fmt.Sprintf("urn:uuid:%d", r.ID),
			Resource: buildFhirMedicationRequest(r),
		})
	}

	bundle := fhirBundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Total:        len(entries),
		Timestamp:    time.Now().Format(time.RFC3339),
		Meta: &fhirMeta{
			LastUpdated: time.Now().Format(time.RFC3339),
		},
		Entry: entries,
		Link: []fhirBundleLink{
			{
				Relation: "self",
				URL:      "http://" + c.Request.Host + "/api/fhir/MedicationRequest?patient=" + patientParam,
			},
		},
	}

	fhirContentType(c)
	c.JSON(http.StatusOK, bundle)
}

func fhirMedicationRequestGet(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		fhirError(c, http.StatusBadRequest, "error", "required",
			"Prescription ID is required")
		return
	}
	prescriptionID := common.ParseInt(id)
	if prescriptionID == 0 {
		fhirError(c, http.StatusBadRequest, "error", "value",
			"Invalid prescription ID")
		return
	}

	row, err := model.Queries.FhirMedicationRequestById(c.Request.Context(), prescriptionID)
	if err != nil {
		if err == sql.ErrNoRows {
			fhirError(c, http.StatusNotFound, "error", "not-found",
				fmt.Sprintf("MedicationRequest/%d not found", prescriptionID))
			return
		}
		log.Printf("fhirMedicationRequestGet: %v", err)
		fhirError(c, http.StatusInternalServerError, "error", "exception",
			"Internal server error")
		return
	}

	fhirContentType(c)
	c.JSON(http.StatusOK, buildFhirMedicationRequest(dbgen.FhirMedicationRequestsByPatientRow{
		ID:                  row.ID,
		Patient:             row.Patient,
		DrugName:            row.DrugName,
		Dosage:              row.Dosage,
		Frequency:           row.Frequency,
		Quantity:            row.Quantity,
		Refills:             row.Refills,
		DateWritten:         row.DateWritten,
		PrescribingProvider: row.PrescribingProvider,
		Pharmacy:            row.Pharmacy,
		Status:              row.Status,
		Notes:               row.Notes,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}))
}

func buildFhirMedicationRequest(r dbgen.FhirMedicationRequestsByPatientRow) fhirMedicationRequest {
	mr := fhirMedicationRequest{
		ResourceType: "MedicationRequest",
		ID:           strconv.FormatInt(r.ID, 10),
		Meta: fhirMeta{
			VersionID:   strconv.FormatInt(r.UpdatedAt.Unix(), 10),
			LastUpdated: r.UpdatedAt.Format(time.RFC3339),
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/MedicationRequest"},
		},
		Intent: "order",
		MedicationCodeableConcept: &fhirCodeableConcept{
			Text: r.DrugName,
		},
		Subject: &fhirReference{
			Reference: fmt.Sprintf("Patient/%d", r.Patient),
		},
		AuthoredOn: r.DateWritten.Format(time.RFC3339),
		Text: &fhirNarrative{
			Status: "generated",
			Div: fmt.Sprintf(
				`<div xmlns="http://www.w3.org/1999/xhtml">%s %s</div>`,
				r.DrugName, r.Dosage),
		},
	}

	// Status mapping
	switch r.Status {
	case "active":
		mr.Status = "active"
	case "discontinued":
		mr.Status = "stopped"
	case "completed":
		mr.Status = "completed"
	default:
		mr.Status = "active"
	}

	// Requester (prescribing provider)
	if r.PrescribingProvider > 0 {
		mr.Requester = &fhirReference{
			Reference: fmt.Sprintf("Practitioner/%d", r.PrescribingProvider),
		}
	}

	// Dosage instruction
	sig := strings.TrimSpace(r.Dosage + " " + r.Frequency)
	if r.Notes.Valid && r.Notes.String != "" {
		sig = r.Notes.String
	}
	if sig != "" {
		mr.DosageInstruction = []fhirDosageInstruction{{Text: sig}}
	}

	// Dispense request (refills)
	mr.DispenseRequest = &fhirDispenseRequest{
		NumberOfRepeatsAllowed: r.Refills,
	}

	return mr
}

// ============================================================================
// Immunization resource (GET /api/fhir/Immunization, GET /api/fhir/Immunization/:id)
// ============================================================================

type fhirImmunization struct {
	ResourceType         string               `json:"resourceType"`
	ID                   string               `json:"id"`
	Meta                 fhirMeta             `json:"meta"`
	Text                 *fhirNarrative       `json:"text,omitempty"`
	Status               string               `json:"status"`
	VaccineCode          fhirCodeableConcept  `json:"vaccineCode"`
	Patient              fhirReference        `json:"patient"`
	OccurrenceDateTime   string               `json:"occurrenceDateTime"`
	LotNumber            string               `json:"lotNumber,omitempty"`
	Manufacturer         *fhirReference       `json:"manufacturer,omitempty"`
	Performer            []fhirReference      `json:"performer,omitempty"`
	PrimarySource        bool                 `json:"primarySource"`
}

func fhirImmunizationList(c *gin.Context) {
	patientParam := c.Query("patient")
	if patientParam == "" {
		fhirError(c, http.StatusBadRequest, "error", "required",
			"patient parameter is required")
		return
	}
	patientID := common.ParseInt(patientParam)
	if patientID == 0 {
		fhirError(c, http.StatusBadRequest, "error", "value",
			"Invalid patient parameter")
		return
	}

	rows, err := model.Queries.FhirImmunizationsByPatient(c.Request.Context(), patientID)
	if err != nil {
		log.Printf("fhirImmunizationList: %v", err)
		fhirError(c, http.StatusInternalServerError, "error", "exception",
			"Internal server error querying immunizations")
		return
	}
	if rows == nil {
		rows = make([]dbgen.FhirImmunizationsByPatientRow, 0)
	}

	entries := make([]fhirBundleEntry, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, fhirBundleEntry{
			FullURL:  fmt.Sprintf("urn:uuid:%d", r.ID),
			Resource: buildFhirImmunization(r),
		})
	}

	bundle := fhirBundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Total:        len(entries),
		Timestamp:    time.Now().Format(time.RFC3339),
		Meta: &fhirMeta{
			LastUpdated: time.Now().Format(time.RFC3339),
		},
		Entry: entries,
		Link: []fhirBundleLink{
			{
				Relation: "self",
				URL:      "http://" + c.Request.Host + "/api/fhir/Immunization?patient=" + patientParam,
			},
		},
	}

	fhirContentType(c)
	c.JSON(http.StatusOK, bundle)
}

func fhirImmunizationGet(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		fhirError(c, http.StatusBadRequest, "error", "required",
			"Immunization ID is required")
		return
	}
	immunizationID := common.ParseInt(id)
	if immunizationID == 0 {
		fhirError(c, http.StatusBadRequest, "error", "value",
			"Invalid immunization ID")
		return
	}

	row, err := model.Queries.FhirImmunizationById(c.Request.Context(), immunizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			fhirError(c, http.StatusNotFound, "error", "not-found",
				fmt.Sprintf("Immunization/%d not found", immunizationID))
			return
		}
		log.Printf("fhirImmunizationGet: %v", err)
		fhirError(c, http.StatusInternalServerError, "error", "exception",
			"Internal server error")
		return
	}

	fhirContentType(c)
	c.JSON(http.StatusOK, buildFhirImmunization(dbgen.FhirImmunizationsByPatientRow{
		ID:            row.ID,
		Patient:       row.Patient,
		Dateof:        row.Dateof,
		Provider:      row.Provider,
		AdminProvider: row.AdminProvider,
		Eoc:           row.Eoc,
		Immunization:  row.Immunization,
		Route:         row.Route,
		BodySite:      row.BodySite,
		Manufacturer:  row.Manufacturer,
		LotNumber:     row.LotNumber,
		PreviousDoses: row.PreviousDoses,
		Recovered:     row.Recovered,
		Notes:         row.Notes,
		Orderid:       row.Orderid,
		Locked:        row.Locked,
		User:          row.User,
		Active:        row.Active,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}))
}

func buildFhirImmunization(r dbgen.FhirImmunizationsByPatientRow) fhirImmunization {
	imm := fhirImmunization{
		ResourceType: "Immunization",
		ID:           strconv.FormatInt(r.ID, 10),
		Meta: fhirMeta{
			VersionID:   strconv.FormatInt(r.UpdatedAt.Unix(), 10),
			LastUpdated: r.UpdatedAt.Format(time.RFC3339),
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/Immunization"},
		},
		Status: "completed",
		VaccineCode: fhirCodeableConcept{
			Coding: []fhirCoding{{
				System:  "http://hl7.org/fhir/sid/cvx",
				Code:    strconv.FormatInt(r.Immunization, 10),
				Display: "Vaccine " + strconv.FormatInt(r.Immunization, 10),
			}},
		},
		Patient: fhirReference{
			Reference: fmt.Sprintf("Patient/%d", r.Patient),
		},
		OccurrenceDateTime: r.Dateof.Format(time.RFC3339),
		PrimarySource:      true,
		Text: &fhirNarrative{
			Status: "generated",
			Div: fmt.Sprintf(
				`<div xmlns="http://www.w3.org/1999/xhtml">Immunization administered %s</div>`,
				r.Dateof.Format("2006-01-02")),
		},
	}

	// Lot number
	if r.LotNumber.Valid && r.LotNumber.String != "" {
		imm.LotNumber = r.LotNumber.String
	}

	// Manufacturer as reference string
	if r.Manufacturer.Valid && r.Manufacturer.String != "" {
		imm.Manufacturer = &fhirReference{
			Display: r.Manufacturer.String,
		}
	}

	// Performer (administering provider and ordering provider)
	if r.AdminProvider > 0 {
		imm.Performer = append(imm.Performer, fhirReference{
			Reference: fmt.Sprintf("Practitioner/%d", r.AdminProvider),
		})
	}
	if r.Provider > 0 && r.Provider != r.AdminProvider {
		imm.Performer = append(imm.Performer, fhirReference{
			Reference: fmt.Sprintf("Practitioner/%d", r.Provider),
		})
	}

	return imm
}

// ============================================================================
// Procedure resource (GET /api/fhir/Procedure, GET /api/fhir/Procedure/:id)
// ============================================================================

type fhirProcedure struct {
	ResourceType      string             `json:"resourceType"`
	ID                string             `json:"id"`
	Meta              fhirMeta           `json:"meta"`
	Text              *fhirNarrative     `json:"text,omitempty"`
	Status            string             `json:"status"`
	Code              fhirCodeableConcept `json:"code"`
	Subject           *fhirReference     `json:"subject,omitempty"`
	PerformedDateTime string             `json:"performedDateTime,omitempty"`
}

func fhirProcedureList(c *gin.Context) {
	patientParam := c.Query("patient")
	if patientParam == "" {
		fhirError(c, http.StatusBadRequest, "error", "required",
			"patient parameter is required")
		return
	}
	patientID := common.ParseInt(patientParam)
	if patientID == 0 {
		fhirError(c, http.StatusBadRequest, "error", "value",
			"Invalid patient parameter")
		return
	}

	rows, err := model.Queries.FhirProceduresByPatient(c.Request.Context(), patientID)
	if err != nil {
		log.Printf("fhirProcedureList: %v", err)
		fhirError(c, http.StatusInternalServerError, "error", "exception",
			"Internal server error querying procedures")
		return
	}
	if rows == nil {
		rows = make([]dbgen.FhirProceduresByPatientRow, 0)
	}

	entries := make([]fhirBundleEntry, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, fhirBundleEntry{
			FullURL:  fmt.Sprintf("urn:uuid:%d", r.ID),
			Resource: buildFhirProcedure(r),
		})
	}

	bundle := fhirBundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Total:        len(entries),
		Timestamp:    time.Now().Format(time.RFC3339),
		Meta: &fhirMeta{
			LastUpdated: time.Now().Format(time.RFC3339),
		},
		Entry: entries,
		Link: []fhirBundleLink{
			{
				Relation: "self",
				URL:      "http://" + c.Request.Host + "/api/fhir/Procedure?patient=" + patientParam,
			},
		},
	}

	fhirContentType(c)
	c.JSON(http.StatusOK, bundle)
}

func fhirProcedureGet(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		fhirError(c, http.StatusBadRequest, "error", "required",
			"Procedure ID is required")
		return
	}
	procedureID := common.ParseInt(id)
	if procedureID == 0 {
		fhirError(c, http.StatusBadRequest, "error", "value",
			"Invalid procedure ID")
		return
	}

	row, err := model.Queries.FhirProcedureById(c.Request.Context(), procedureID)
	if err != nil {
		if err == sql.ErrNoRows {
			fhirError(c, http.StatusNotFound, "error", "not-found",
				fmt.Sprintf("Procedure/%d not found", procedureID))
			return
		}
		log.Printf("fhirProcedureGet: %v", err)
		fhirError(c, http.StatusInternalServerError, "error", "exception",
			"Internal server error")
		return
	}

	fhirContentType(c)
	c.JSON(http.StatusOK, buildFhirProcedure(dbgen.FhirProceduresByPatientRow{
		ID:            row.ID,
		Patient:       row.Patient,
		OperationDate: row.OperationDate,
		Operation:     row.Operation,
		User:          row.User,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}))
}

func buildFhirProcedure(r dbgen.FhirProceduresByPatientRow) fhirProcedure {
	proc := fhirProcedure{
		ResourceType: "Procedure",
		ID:           strconv.FormatInt(r.ID, 10),
		Meta: fhirMeta{
			VersionID:   strconv.FormatInt(r.UpdatedAt.Unix(), 10),
			LastUpdated: r.UpdatedAt.Format(time.RFC3339),
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/Procedure"},
		},
		Status: "completed",
		Code: fhirCodeableConcept{
			Text: r.Operation,
		},
		Subject: &fhirReference{
			Reference: fmt.Sprintf("Patient/%d", r.Patient),
		},
		Text: &fhirNarrative{
			Status: "generated",
			Div: fmt.Sprintf(
				`<div xmlns="http://www.w3.org/1999/xhtml">Procedure: %s</div>`,
				r.Operation),
		},
	}

	// Performed date
	if r.OperationDate.Valid {
		proc.PerformedDateTime = r.OperationDate.Time.Format("2006-01-02")
	}

	return proc
}

// ============================================================================
// Encounter resource (GET /api/fhir/Encounter, GET /api/fhir/Encounter/:id)
// ============================================================================

type fhirPeriod struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}

type fhirEncounterParticipant struct {
	Individual *fhirReference `json:"individual,omitempty"`
}

type fhirEncounter struct {
	ResourceType string                     `json:"resourceType"`
	ID           string                     `json:"id"`
	Meta         fhirMeta                   `json:"meta"`
	Text         *fhirNarrative             `json:"text,omitempty"`
	Status       string                     `json:"status"`
	Class        fhirCoding                 `json:"class"`
	Type         []fhirCodeableConcept      `json:"type,omitempty"`
	Subject      *fhirReference             `json:"subject,omitempty"`
	Participant  []fhirEncounterParticipant `json:"participant,omitempty"`
	Period       *fhirPeriod                `json:"period,omitempty"`
}

func fhirEncounterList(c *gin.Context) {
	patientParam := c.Query("patient")
	if patientParam == "" {
		fhirError(c, http.StatusBadRequest, "error", "required",
			"patient parameter is required")
		return
	}
	patientID := common.ParseInt(patientParam)
	if patientID == 0 {
		fhirError(c, http.StatusBadRequest, "error", "value",
			"Invalid patient parameter")
		return
	}

	rows, err := model.Queries.FhirEncountersByPatient(c.Request.Context(), patientID)
	if err != nil {
		log.Printf("fhirEncounterList: %v", err)
		fhirError(c, http.StatusInternalServerError, "error", "exception",
			"Internal server error querying encounters")
		return
	}
	if rows == nil {
		rows = make([]dbgen.FhirEncountersByPatientRow, 0)
	}

	entries := make([]fhirBundleEntry, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, fhirBundleEntry{
			FullURL:  fmt.Sprintf("urn:uuid:%d", r.ID),
			Resource: buildFhirEncounter(r),
		})
	}

	bundle := fhirBundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Total:        len(entries),
		Timestamp:    time.Now().Format(time.RFC3339),
		Meta: &fhirMeta{
			LastUpdated: time.Now().Format(time.RFC3339),
		},
		Entry: entries,
		Link: []fhirBundleLink{
			{
				Relation: "self",
				URL:      "http://" + c.Request.Host + "/api/fhir/Encounter?patient=" + patientParam,
			},
		},
	}

	fhirContentType(c)
	c.JSON(http.StatusOK, bundle)
}

func fhirEncounterGet(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		fhirError(c, http.StatusBadRequest, "error", "required",
			"Encounter ID is required")
		return
	}
	encounterID := common.ParseInt(id)
	if encounterID == 0 {
		fhirError(c, http.StatusBadRequest, "error", "value",
			"Invalid encounter ID")
		return
	}

	row, err := model.Queries.FhirEncounterById(c.Request.Context(), encounterID)
	if err != nil {
		if err == sql.ErrNoRows {
			fhirError(c, http.StatusNotFound, "error", "not-found",
				fmt.Sprintf("Encounter/%d not found", encounterID))
			return
		}
		log.Printf("fhirEncounterGet: %v", err)
		fhirError(c, http.StatusInternalServerError, "error", "exception",
			"Internal server error")
		return
	}

	fhirContentType(c)
	c.JSON(http.StatusOK, buildFhirEncounter(dbgen.FhirEncountersByPatientRow{
		ID:               row.ID,
		Patient:          row.Patient,
		Eoc:              row.Eoc,
		Cpt:              row.Cpt,
		CptMod:           row.CptMod,
		Physician:        row.Physician,
		EncounterDate:    row.EncounterDate,
		EncounterEndDate: row.EncounterEndDate,
		PlaceOfService:   row.PlaceOfService,
		Comment:          row.Comment,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}))
}

func buildFhirEncounter(r dbgen.FhirEncountersByPatientRow) fhirEncounter {
	enc := fhirEncounter{
		ResourceType: "Encounter",
		ID:           strconv.FormatInt(r.ID, 10),
		Meta: fhirMeta{
			VersionID:   strconv.FormatInt(r.UpdatedAt.Unix(), 10),
			LastUpdated: r.UpdatedAt.Format(time.RFC3339),
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/Encounter"},
		},
		Status: "finished",
		Class: fhirCoding{
			System:  "http://terminology.hl7.org/CodeSystem/v3-ActCode",
			Code:    "AMB",
			Display: "ambulatory",
		},
		Subject: &fhirReference{
			Reference: fmt.Sprintf("Patient/%d", r.Patient),
		},
		Text: &fhirNarrative{
			Status: "generated",
			Div: fmt.Sprintf(
				`<div xmlns="http://www.w3.org/1999/xhtml">Encounter on %s</div>`,
				r.EncounterDate.Format("2006-01-02")),
		},
	}

	// Type from CPT
	if r.Cpt > 0 {
		enc.Type = []fhirCodeableConcept{{
			Coding: []fhirCoding{{
				System:  "http://www.ama-assn.org/go/cpt",
				Code:    strconv.FormatInt(r.Cpt, 10),
				Display: "CPT " + strconv.FormatInt(r.Cpt, 10),
			}},
		}}
	}

	// Period
	period := &fhirPeriod{
		Start: r.EncounterDate.Format(time.RFC3339),
	}
	if r.EncounterEndDate.Valid {
		period.End = r.EncounterEndDate.Time.Format(time.RFC3339)
	}
	enc.Period = period

	// Participant (physician)
	if r.Physician > 0 {
		enc.Participant = []fhirEncounterParticipant{
			{
				Individual: &fhirReference{
					Reference: fmt.Sprintf("Practitioner/%d", r.Physician),
				},
			},
		}
	}

	return enc
}

// ============================================================================
// FamilyMemberHistory resource (GET /api/fhir/FamilyMemberHistory)
// ============================================================================

type fhirFamilyMemberHistory struct {
	ResourceType    string              `json:"resourceType"`
	ID              string              `json:"id"`
	Meta            fhirMeta            `json:"meta"`
	Text            *fhirNarrative      `json:"text,omitempty"`
	Status          string              `json:"status"`
	Patient         fhirReference       `json:"patient"`
	Relationship    fhirCodeableConcept `json:"relationship"`
	Condition       []fhirCodeableConcept `json:"condition,omitempty"`
	OnsetAge        *fhirQuantity       `json:"onsetAge,omitempty"`
	DeceasedBoolean *bool               `json:"deceasedBoolean,omitempty"`
}

// mapRelationshipToSNOMED maps a relationship string to a SNOMED CT family member code.
func mapRelationshipToSNOMED(relationship string) fhirCoding {
	r := strings.ToLower(strings.TrimSpace(relationship))
	switch r {
	case "mother":
		return fhirCoding{System: "http://snomed.info/sct", Code: "72705000", Display: "Mother"}
	case "father":
		return fhirCoding{System: "http://snomed.info/sct", Code: "66839005", Display: "Father"}
	case "brother":
		return fhirCoding{System: "http://snomed.info/sct", Code: "70924004", Display: "Brother"}
	case "sister":
		return fhirCoding{System: "http://snomed.info/sct", Code: "27733009", Display: "Sister"}
	case "son":
		return fhirCoding{System: "http://snomed.info/sct", Code: "65616008", Display: "Son"}
	case "daughter":
		return fhirCoding{System: "http://snomed.info/sct", Code: "66089001", Display: "Daughter"}
	case "maternal grandmother", "grandmother (maternal)":
		return fhirCoding{System: "http://snomed.info/sct", Code: "17910006", Display: "Maternal grandmother"}
	case "maternal grandfather", "grandfather (maternal)":
		return fhirCoding{System: "http://snomed.info/sct", Code: "445448009", Display: "Maternal grandfather"}
	case "paternal grandmother", "grandmother (paternal)":
		return fhirCoding{System: "http://snomed.info/sct", Code: "17538003", Display: "Paternal grandmother"}
	case "paternal grandfather", "grandfather (paternal)":
		return fhirCoding{System: "http://snomed.info/sct", Code: "31611000", Display: "Paternal grandfather"}
	case "grandmother":
		return fhirCoding{System: "http://snomed.info/sct", Code: "410610009", Display: "Grandmother"}
	case "grandfather":
		return fhirCoding{System: "http://snomed.info/sct", Code: "410600008", Display: "Grandfather"}
	case "aunt":
		return fhirCoding{System: "http://snomed.info/sct", Code: "12683003", Display: "Aunt"}
	case "uncle":
		return fhirCoding{System: "http://snomed.info/sct", Code: "38048003", Display: "Uncle"}
	case "cousin":
		return fhirCoding{System: "http://snomed.info/sct", Code: "55814001", Display: "Cousin"}
	case "nephew":
		return fhirCoding{System: "http://snomed.info/sct", Code: "66904004", Display: "Nephew"}
	case "niece":
		return fhirCoding{System: "http://snomed.info/sct", Code: "65953000", Display: "Niece"}
	default:
		return fhirCoding{System: "http://snomed.info/sct", Code: "444295003", Display: "Family member"}
	}
}

func fhirFamilyMemberHistoryList(c *gin.Context) {
	patientParam := c.Query("patient")
	if patientParam == "" {
		fhirError(c, http.StatusBadRequest, "error", "required",
			"patient parameter is required")
		return
	}
	patientID := common.ParseInt(patientParam)
	if patientID == 0 {
		fhirError(c, http.StatusBadRequest, "error", "value",
			"Invalid patient parameter")
		return
	}

	rows, err := model.Queries.FhirFamilyHistoryByPatient(c.Request.Context(), patientID)
	if err != nil {
		log.Printf("fhirFamilyMemberHistoryList: %v", err)
		fhirError(c, http.StatusInternalServerError, "error", "exception",
			"Internal server error querying family history")
		return
	}
	if rows == nil {
		rows = make([]dbgen.FhirFamilyHistoryByPatientRow, 0)
	}

	entries := make([]fhirBundleEntry, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, fhirBundleEntry{
			FullURL:  fmt.Sprintf("urn:uuid:%d", r.ID),
			Resource: buildFhirFamilyMemberHistory(dbgen.FhirFamilyHistoryByPatientRow(r)),
		})
	}

	bundle := fhirBundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Total:        len(entries),
		Timestamp:    time.Now().Format(time.RFC3339),
		Meta: &fhirMeta{
			LastUpdated: time.Now().Format(time.RFC3339),
		},
		Entry: entries,
		Link: []fhirBundleLink{
			{
				Relation: "self",
				URL:      "http://" + c.Request.Host + "/api/fhir/FamilyMemberHistory?patient=" + patientParam,
			},
		},
	}

	fhirContentType(c)
	c.JSON(http.StatusOK, bundle)
}

func fhirFamilyMemberHistoryGet(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		fhirError(c, http.StatusBadRequest, "error", "required",
			"FamilyMemberHistory ID is required")
		return
	}
	historyID := common.ParseInt(id)
	if historyID == 0 {
		fhirError(c, http.StatusBadRequest, "error", "value",
			"Invalid family history ID")
		return
	}

	row, err := model.Queries.FhirFamilyHistoryById(c.Request.Context(), historyID)
	if err != nil {
		if err == sql.ErrNoRows {
			fhirError(c, http.StatusNotFound, "error", "not-found",
				fmt.Sprintf("FamilyMemberHistory/%d not found", historyID))
			return
		}
		log.Printf("fhirFamilyMemberHistoryGet: %v", err)
		fhirError(c, http.StatusInternalServerError, "error", "exception",
			"Internal server error")
		return
	}

	fhirContentType(c)
	c.JSON(http.StatusOK, buildFhirFamilyMemberHistory(dbgen.FhirFamilyHistoryByPatientRow{
		ID:            row.ID,
		Patient:       row.Patient,
		Relationship:  row.Relationship,
		ConditionName: row.ConditionName,
		Icd10Code:     row.Icd10Code,
		OnsetAge:      row.OnsetAge,
		Deceased:      row.Deceased,
		Notes:         row.Notes,
		User:          row.User,
		Active:        row.Active,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}))
}

func buildFhirFamilyMemberHistory(r dbgen.FhirFamilyHistoryByPatientRow) fhirFamilyMemberHistory {
	fmh := fhirFamilyMemberHistory{
		ResourceType: "FamilyMemberHistory",
		ID:           strconv.FormatInt(r.ID, 10),
		Meta: fhirMeta{
			VersionID:   strconv.FormatInt(r.UpdatedAt.Unix(), 10),
			LastUpdated: r.UpdatedAt.Format(time.RFC3339),
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/FamilyMemberHistory"},
		},
		Status: "completed",
		Patient: fhirReference{
			Reference: fmt.Sprintf("Patient/%d", r.Patient),
		},
		Relationship: fhirCodeableConcept{
			Coding: []fhirCoding{mapRelationshipToSNOMED(r.Relationship)},
			Text:   r.Relationship,
		},
		Text: &fhirNarrative{
			Status: "generated",
			Div: fmt.Sprintf(
				`<div xmlns="http://www.w3.org/1999/xhtml">Family history: %s - %s</div>`,
				r.Relationship, r.ConditionName),
		},
	}

	// Condition with ICD-10 code
	condition := fhirCodeableConcept{Text: r.ConditionName}
	if r.Icd10Code != "" {
		condition.Coding = []fhirCoding{
			{System: "http://hl7.org/fhir/sid/icd-10", Code: r.Icd10Code, Display: r.ConditionName},
		}
	}
	fmh.Condition = []fhirCodeableConcept{condition}

	// Onset age if > 0
	if r.OnsetAge > 0 {
		fmh.OnsetAge = &fhirQuantity{
			Value:  float64(r.OnsetAge),
			Unit:   "years",
			System: "http://unitsofmeasure.org",
			Code:   "a",
		}
	}

	// Deceased
	if r.Deceased {
		t := true
		fmh.DeceasedBoolean = &t
	}

	return fmh
}

// ============================================================================
// Patient $document operation (GET /api/fhir/Patient/:id/$document)
// ============================================================================

// fhirPatientDocument generates and returns a C-CDA CCD document for the patient.
func fhirPatientDocument(c *gin.Context) {
	id := c.Param("id")
	patientID := common.ParseInt(id)
	if patientID == 0 {
		fhirError(c, http.StatusBadRequest, "error", "value", "Invalid patient ID")
		return
	}

	xmlData, err := ccda.GenerateCCD(patientID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			fhirError(c, http.StatusNotFound, "error", "not-found",
				fmt.Sprintf("Patient/%d not found", patientID))
			return
		}
		log.Printf("fhirPatientDocument: error generating CCD for patient %d: %v", patientID, err)
		fhirError(c, http.StatusInternalServerError, "error", "exception",
			"Internal server error generating CCD document")
		return
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=ccd_patient_"+id+".xml")
	c.Data(http.StatusOK, "application/xml; charset=utf-8", xmlData)
}
