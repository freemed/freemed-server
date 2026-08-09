package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestFhirContentType verifies the Content-Type header is set correctly.
func TestFhirContentType(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	fhirContentType(c)

	ct := w.Header().Get("Content-Type")
	expected := "application/fhir+json; charset=utf-8"
	if ct != expected {
		t.Errorf("fhirContentType = %q, want %q", ct, expected)
	}
}

// TestFhirError_OperationOutcome verifies fhirError produces the correct
// OperationOutcome structure with correct severity and status code.
func TestFhirError_OperationOutcome(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	fhirError(c, http.StatusBadRequest, "error", "required", "Patient ID is required")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/fhir+json; charset=utf-8" {
		t.Errorf("expected FHIR content type, got %q", ct)
	}

	var outcome fhirOperationOutcome
	if err := json.Unmarshal(w.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to unmarshal OperationOutcome: %v", err)
	}

	if outcome.ResourceType != "OperationOutcome" {
		t.Errorf("ResourceType = %q, want %q", outcome.ResourceType, "OperationOutcome")
	}
	if len(outcome.Issue) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(outcome.Issue))
	}
	if outcome.Issue[0].Severity != "error" {
		t.Errorf("Issue severity = %q, want %q", outcome.Issue[0].Severity, "error")
	}
	if outcome.Issue[0].Code != "required" {
		t.Errorf("Issue code = %q, want %q", outcome.Issue[0].Code, "required")
	}
	if outcome.Issue[0].Diagnostics != "Patient ID is required" {
		t.Errorf("Issue diagnostics = %q, want %q", outcome.Issue[0].Diagnostics, "Patient ID is required")
	}
}

// TestFhirError_NotFound verifies a 404 OperationOutcome.
func TestFhirError_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	fhirError(c, http.StatusNotFound, "error", "not-found", "Patient/999 not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var outcome fhirOperationOutcome
	if err := json.Unmarshal(w.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if outcome.Issue[0].Code != "not-found" {
		t.Errorf("expected code 'not-found', got %q", outcome.Issue[0].Code)
	}
}

// TestMapPtsexToFhirGender tests all gender mapping paths.
func TestMapPtsexToFhirGender(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"m", "male"},
		{"M", "male"},
		{"Male", "male"},
		{"male", "male"},
		{"MALE", "male"},
		{"f", "female"},
		{"F", "female"},
		{"Female", "female"},
		{"female", "female"},
		{"o", "other"},
		{"O", "other"},
		{"Other", "other"},
		{"", "unknown"},
		{"X", "unknown"},
		{"unknown", "unknown"},
		{"  m  ", "male"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := mapPtsexToFhirGender(tt.input)
			if got != tt.want {
				t.Errorf("mapPtsexToFhirGender(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestFhirResourceStructs_JSON_Marshal verifies FHIR resource structs marshal
// to valid JSON.
func TestFhirResourceStructs_JSON_Marshal(t *testing.T) {
	t.Run("fhirPatient", func(t *testing.T) {
		p := fhirPatient{
			ResourceType: "Patient",
			ID:           "123",
			Meta:         fhirMeta{VersionID: "1", LastUpdated: "2026-01-01T00:00:00Z"},
			Name: []fhirHumanName{
				{Use: "official", Family: "Doe", Given: []string{"John"}},
			},
			Gender:    "male",
			BirthDate: "1980-01-01",
		}
		data, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("marshal fhirPatient: %v", err)
		}
		if !strings.Contains(string(data), `"resourceType":"Patient"`) {
			t.Error("missing resourceType in fhirPatient JSON")
		}
		if !strings.Contains(string(data), `"family":"Doe"`) {
			t.Error("missing family name in fhirPatient JSON")
		}
	})

	t.Run("fhirObservation", func(t *testing.T) {
		obs := fhirObservation{
			ResourceType: "Observation",
			ID:           "456",
			Status:       "final",
			Category: []fhirCodeableConcept{{
				Coding: []fhirCoding{{System: "http://terminology.hl7.org/CodeSystem/observation-category", Code: "vital-signs"}},
			}},
			Code: fhirCodeableConcept{
				Coding: []fhirCoding{{System: "http://loinc.org", Code: "85353-1"}},
			},
			Subject: &fhirReference{Reference: "Patient/1"},
		}
		data, err := json.Marshal(obs)
		if err != nil {
			t.Fatalf("marshal fhirObservation: %v", err)
		}
		if !strings.Contains(string(data), `"resourceType":"Observation"`) {
			t.Error("missing resourceType in fhirObservation JSON")
		}
		if !strings.Contains(string(data), `"status":"final"`) {
			t.Error("missing status in fhirObservation JSON")
		}
	})

	t.Run("fhirCondition", func(t *testing.T) {
		cond := fhirCondition{
			ResourceType: "Condition",
			ID:           "789",
			Subject:      &fhirReference{Reference: "Patient/1"},
			Code: &fhirCodeableConcept{
				Coding: []fhirCoding{{System: "http://snomed.info/sct", Code: "38341003", Display: "Hypertension"}},
			},
		}
		data, err := json.Marshal(cond)
		if err != nil {
			t.Fatalf("marshal fhirCondition: %v", err)
		}
		if !strings.Contains(string(data), `"resourceType":"Condition"`) {
			t.Error("missing resourceType in fhirCondition JSON")
		}
	})

	t.Run("fhirBundle", func(t *testing.T) {
		bundle := fhirBundle{
			ResourceType: "Bundle",
			Type:         "searchset",
			Total:        1,
			Entry: []fhirBundleEntry{
				{FullURL: "urn:uuid:123", Resource: fhirPatient{ResourceType: "Patient", ID: "1"}},
			},
		}
		data, err := json.Marshal(bundle)
		if err != nil {
			t.Fatalf("marshal fhirBundle: %v", err)
		}
		if !strings.Contains(string(data), `"resourceType":"Bundle"`) {
			t.Error("missing resourceType in fhirBundle JSON")
		}
		if !strings.Contains(string(data), `"type":"searchset"`) {
			t.Error("missing type in fhirBundle JSON")
		}
	})

	t.Run("fhirOperationOutcome", func(t *testing.T) {
		oo := fhirOperationOutcome{
			ResourceType: "OperationOutcome",
			Issue: []fhirOperationOutcomeIssue{
				{Severity: "error", Code: "required", Diagnostics: "Missing parameter"},
			},
		}
		data, err := json.Marshal(oo)
		if err != nil {
			t.Fatalf("marshal fhirOperationOutcome: %v", err)
		}
		if !strings.Contains(string(data), `"resourceType":"OperationOutcome"`) {
			t.Error("missing resourceType")
		}
		if !strings.Contains(string(data), `"severity":"error"`) {
			t.Error("missing severity in issue")
		}
	})
}

// TestFhirCodingAndIdentifier verifies the basic FHIR building blocks marshal correctly.
func TestFhirCodingAndIdentifier(t *testing.T) {
	t.Run("fhirCoding", func(t *testing.T) {
		coding := fhirCoding{System: "http://loinc.org", Code: "8480-6", Display: "Systolic BP"}
		data, err := json.Marshal(coding)
		if err != nil {
			t.Fatalf("marshal fhirCoding: %v", err)
		}
		if !strings.Contains(string(data), `"system":"http://loinc.org"`) {
			t.Error("coding missing system")
		}
	})

	t.Run("fhirIdentifier", func(t *testing.T) {
		id := fhirIdentifier{Use: "usual", System: "http://freemed.local/fhir/identifier/mrn", Value: "MRN-001"}
		data, err := json.Marshal(id)
		if err != nil {
			t.Fatalf("marshal fhirIdentifier: %v", err)
		}
		if !strings.Contains(string(data), `"value":"MRN-001"`) {
			t.Error("identifier missing value")
		}
	})

	t.Run("fhirReference", func(t *testing.T) {
		ref := fhirReference{Reference: "Patient/42", Display: "John Doe"}
		data, err := json.Marshal(ref)
		if err != nil {
			t.Fatalf("marshal fhirReference: %v", err)
		}
		if !strings.Contains(string(data), `"reference":"Patient/42"`) {
			t.Error("reference missing reference field")
		}
	})

	t.Run("fhirQuantity", func(t *testing.T) {
		q := fhirQuantity{Value: 120.5, Unit: "mm[Hg]", System: "http://unitsofmeasure.org", Code: "mm[Hg]"}
		data, err := json.Marshal(q)
		if err != nil {
			t.Fatalf("marshal fhirQuantity: %v", err)
		}
		if !strings.Contains(string(data), `"value":120.5`) {
			t.Error("quantity missing value")
		}
	})

	t.Run("fhirAddress", func(t *testing.T) {
		addr := fhirAddress{Use: "home", Line: []string{"123 Main St"}, City: "Hartford", State: "CT", PostalCode: "06101"}
		data, err := json.Marshal(addr)
		if err != nil {
			t.Fatalf("marshal fhirAddress: %v", err)
		}
		if !strings.Contains(string(data), `"city":"Hartford"`) {
			t.Error("address missing city")
		}
	})

	t.Run("fhirHumanName", func(t *testing.T) {
		name := fhirHumanName{Use: "official", Family: "Smith", Given: []string{"Jane", "Marie"}}
		data, err := json.Marshal(name)
		if err != nil {
			t.Fatalf("marshal fhirHumanName: %v", err)
		}
		if !strings.Contains(string(data), `"family":"Smith"`) {
			t.Error("name missing family")
		}
		if !strings.Contains(string(data), `"given"`) {
			t.Error("name missing given array")
		}
	})
}
