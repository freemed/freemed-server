package cds

// Severity levels for clinical decision support interactions.
type Severity string

const (
	SevContraindicated Severity = "contraindicated"
	SevMajor           Severity = "major"
	SevModerate        Severity = "moderate"
	SevMinor           Severity = "minor"
)

// InteractionResult describes a detected drug interaction or warning.
type InteractionResult struct {
	Severity    Severity `json:"severity"`
	Description string   `json:"description"`
	Drugs       []string `json:"drugs"`
}

// InteractionChecker is the interface for drug interaction checking.
type InteractionChecker interface {
	// CheckDrugDrug evaluates all pairwise drug interactions for a list of drug names.
	CheckDrugDrug(drugs []string) ([]InteractionResult, error)

	// CheckDrugAllergy checks a single drug against a list of patient allergies.
	CheckDrugAllergy(drug string, allergies []string) ([]InteractionResult, error)
}
