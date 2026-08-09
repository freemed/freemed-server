package cds

import "testing"

func TestLocalChecker_DrugDrug_Contraindicated(t *testing.T) {
	checker := NewLocalChecker()
	results, err := checker.CheckDrugDrug([]string{"warfarin", "ibuprofen"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Error("expected interactions for warfarin+ibuprofen")
	}
	found := false
	for _, r := range results {
		if r.Severity == SevContraindicated {
			found = true
		}
	}
	if !found {
		t.Error("expected contraindicated interaction")
	}
}

func TestLocalChecker_DrugDrug_Major(t *testing.T) {
	checker := NewLocalChecker()
	results, err := checker.CheckDrugDrug([]string{"lisinopril", "potassium chloride"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Error("expected interactions for lisinopril+potassium")
	}
	found := false
	for _, r := range results {
		if r.Severity == SevMajor {
			found = true
		}
	}
	if !found {
		t.Error("expected major severity interaction")
	}
}

func TestLocalChecker_DrugDrug_NoInteraction(t *testing.T) {
	checker := NewLocalChecker()
	results, err := checker.CheckDrugDrug([]string{"amoxicillin", "acetaminophen"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected no interactions for amoxicillin+acetaminophen, got %d", len(results))
	}
}

func TestLocalChecker_DrugDrug_Empty(t *testing.T) {
	checker := NewLocalChecker()
	results, err := checker.CheckDrugDrug([]string{})
	if err != nil {
		t.Fatal(err)
	}
	if results != nil {
		t.Errorf("expected nil results for empty drug list, got %v", results)
	}
}

func TestLocalChecker_DrugDrug_Single(t *testing.T) {
	checker := NewLocalChecker()
	results, err := checker.CheckDrugDrug([]string{"warfarin"})
	if err != nil {
		t.Fatal(err)
	}
	if results != nil {
		t.Errorf("expected nil results for single drug, got %v", results)
	}
}

func TestLocalChecker_DrugAllergy_Penicillin(t *testing.T) {
	checker := NewLocalChecker()
	results, err := checker.CheckDrugAllergy("amoxicillin", []string{"penicillin"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Error("expected allergy interaction for penicillin allergy + amoxicillin")
	}
	found := false
	for _, r := range results {
		if r.Severity == SevContraindicated {
			found = true
		}
	}
	if !found {
		t.Error("expected contraindicated severity for penicillin allergy")
	}
}

func TestLocalChecker_DrugAllergy_NoMatch(t *testing.T) {
	checker := NewLocalChecker()
	results, err := checker.CheckDrugAllergy("amoxicillin", []string{"latex", "egg"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected no allergy interactions for non-matching allergies, got %d", len(results))
	}
}

func TestLocalChecker_DrugAllergy_Empty(t *testing.T) {
	checker := NewLocalChecker()
	results, err := checker.CheckDrugAllergy("amoxicillin", []string{})
	if err != nil {
		t.Fatal(err)
	}
	if results != nil {
		t.Errorf("expected nil results for empty allergy list, got %v", results)
	}
}

func TestLocalChecker_DrugAllergy_EmptyDrug(t *testing.T) {
	checker := NewLocalChecker()
	results, err := checker.CheckDrugAllergy("", []string{"penicillin"})
	if err != nil {
		t.Fatal(err)
	}
	if results != nil {
		t.Errorf("expected nil results for empty drug, got %v", results)
	}
}

func TestLocalChecker_MultipleInteractions(t *testing.T) {
	// warfarin+ibuprofen (contraindicated), lisinopril+ibuprofen (moderate), tramadol+diazepam (contraindicated)
	checker := NewLocalChecker()
	results, err := checker.CheckDrugDrug([]string{"warfarin", "ibuprofen", "lisinopril", "tramadol", "diazepam"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) < 2 {
		t.Errorf("expected multiple interactions, got %d: %v", len(results), results)
	}
}
