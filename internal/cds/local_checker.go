package cds

import (
	"fmt"
	"strings"
)

// LocalChecker implements InteractionChecker using hardcoded drug class and
// interaction maps. This is suitable for offline/embedded use and can be
// extended with RxNorm data in the future.
type LocalChecker struct{}

// Ensure LocalChecker satisfies the InteractionChecker interface.
var _ InteractionChecker = (*LocalChecker)(nil)

// drugClassMap maps lowercase drug names to their pharmacologic classes.
var drugClassMap = map[string]string{
	// ACE inhibitors
	"lisinopril":    "ACE inhibitor",
	"enalapril":     "ACE inhibitor",
	"ramipril":      "ACE inhibitor",
	"captopril":     "ACE inhibitor",
	"benazepril":    "ACE inhibitor",
	"quinapril":     "ACE inhibitor",
	"fosinopril":    "ACE inhibitor",
	"moexipril":     "ACE inhibitor",
	"perindopril":   "ACE inhibitor",
	"trandolapril":  "ACE inhibitor",

	// ARBs
	"losartan":      "ARB",
	"valsartan":     "ARB",
	"irbesartan":    "ARB",
	"candesartan":   "ARB",
	"telmisartan":   "ARB",
	"olmesartan":    "ARB",

	// NSAIDs
	"ibuprofen":     "NSAID",
	"naproxen":      "NSAID",
	"diclofenac":    "NSAID",
	"celecoxib":     "NSAID",
	"meloxicam":     "NSAID",
	"indomethacin":  "NSAID",
	"ketorolac":     "NSAID",
	"piroxicam":     "NSAID",
	"aspirin":       "NSAID",

	// Anticoagulants
	"warfarin":              "Anticoagulant",
	"apixaban":              "Anticoagulant",
	"rivaroxaban":           "Anticoagulant",
	"dabigatran":            "Anticoagulant",
	"edoxaban":              "Anticoagulant",
	"heparin":               "Anticoagulant",
	"enoxaparin":            "Anticoagulant",

	// Antiplatelet
	"clopidogrel":           "Antiplatelet",
	"ticagrelor":            "Antiplatelet",
	"prasugrel":             "Antiplatelet",

	// Opioids
	"morphine":              "Opioid",
	"oxycodone":             "Opioid",
	"hydrocodone":           "Opioid",
	"fentanyl":              "Opioid",
	"hydromorphone":         "Opioid",
	"codeine":               "Opioid",
	"tramadol":              "Opioid",
	"methadone":             "Opioid",
	"buprenorphine":         "Opioid",
	"meperidine":            "Opioid",

	// Benzodiazepines
	"diazepam":              "Benzodiazepine",
	"lorazepam":             "Benzodiazepine",
	"alprazolam":            "Benzodiazepine",
	"clonazepam":            "Benzodiazepine",
	"temazepam":             "Benzodiazepine",
	"midazolam":             "Benzodiazepine",
	"chlordiazepoxide":      "Benzodiazepine",

	// SSRIs
	"fluoxetine":            "SSRI",
	"sertraline":            "SSRI",
	"paroxetine":            "SSRI",
	"citalopram":            "SSRI",
	"escitalopram":          "SSRI",
	"fluvoxamine":           "SSRI",

	// SNRIs
	"venlafaxine":           "SNRI",
	"duloxetine":            "SNRI",
	"desvenlafaxine":        "SNRI",

	// MAOIs
	"phenelzine":            "MAOI",
	"tranylcypromine":       "MAOI",
	"isocarboxazid":         "MAOI",
	"selegiline":            "MAOI",

	// Statins
	"atorvastatin":          "Statin",
	"simvastatin":           "Statin",
	"rosuvastatin":          "Statin",
	"pravastatin":           "Statin",
	"lovastatin":            "Statin",
	"pitavastatin":          "Statin",
	"fluvastatin":           "Statin",

	// Macrolide antibiotics
	"erythromycin":          "Macrolide antibiotic",
	"clarithromycin":        "Macrolide antibiotic",
	"azithromycin":          "Macrolide antibiotic",

	// Fluoroquinolone antibiotics
	"ciprofloxacin":         "Fluoroquinolone antibiotic",
	"levofloxacin":          "Fluoroquinolone antibiotic",
	"moxifloxacin":          "Fluoroquinolone antibiotic",

	// Penicillin antibiotics
	"amoxicillin":           "Penicillin antibiotic",
	"penicillin":            "Penicillin antibiotic",
	"ampicillin":            "Penicillin antibiotic",
	"amoxicillin-clavulanate": "Penicillin antibiotic",

	// Sulfonamide antibiotics
	"sulfamethoxazole":      "Sulfonamide antibiotic",
	"trimethoprim-sulfamethoxazole": "Sulfonamide antibiotic",

	// Potassium supplements
	"potassium chloride":    "Potassium supplement",
	"potassium citrate":     "Potassium supplement",
	"potassium gluconate":   "Potassium supplement",

	// Diuretics
	"furosemide":            "Loop diuretic",
	"bumetanide":            "Loop diuretic",
	"torsemide":             "Loop diuretic",
	"hydrochlorothiazide":   "Thiazide diuretic",
	"chlorthalidone":        "Thiazide diuretic",
	"spironolactone":        "Potassium-sparing diuretic",
	"eplerenone":            "Potassium-sparing diuretic",

	// Beta blockers
	"metoprolol":            "Beta blocker",
	"atenolol":              "Beta blocker",
	"propranolol":           "Beta blocker",
	"carvedilol":            "Beta blocker",
	"bisoprolol":            "Beta blocker",

	// Calcium channel blockers
	"amlodipine":            "Calcium channel blocker",
	"nifedipine":            "Calcium channel blocker",
	"diltiazem":             "Calcium channel blocker",
	"verapamil":             "Calcium channel blocker",

	// Antidiabetics
	"metformin":              "Biguanide",
	"glipizide":              "Sulfonylurea",
	"glyburide":              "Sulfonylurea",
	"glimepiride":            "Sulfonylurea",
	"insulin":                "Insulin",
	"pioglitazone":           "Thiazolidinedione",
	"empagliflozin":          "SGLT2 inhibitor",
	"dapagliflozin":          "SGLT2 inhibitor",
	"semaglutide":            "GLP-1 agonist",
	"liraglutide":            "GLP-1 agonist",

	// Thyroid
	"levothyroxine":          "Thyroid hormone",
	"methimazole":            "Antithyroid",
	"propylthiouracil":       "Antithyroid",

	// Anticonvulsants
	"phenytoin":              "Anticonvulsant",
	"carbamazepine":          "Anticonvulsant",
	"valproic acid":          "Anticonvulsant",
	"lamotrigine":            "Anticonvulsant",
	"levetiracetam":          "Anticonvulsant",
	"gabapentin":             "Anticonvulsant",
	"pregabalin":             "Anticonvulsant",

	// Corticosteroids
	"prednisone":             "Corticosteroid",
	"methylprednisolone":     "Corticosteroid",
	"dexamethasone":          "Corticosteroid",
	"hydrocortisone":         "Corticosteroid",

	// Methotrexate
	"methotrexate":           "Antimetabolite",

	// Lithium
	"lithium":                "Mood stabilizer",

	// Digoxin
	"digoxin":                "Cardiac glycoside",

	// Theophylline
	"theophylline":           "Xanthine derivative",

	// Proton pump inhibitors
	"omeprazole":             "Proton pump inhibitor",
	"esomeprazole":           "Proton pump inhibitor",
	"lansoprazole":           "Proton pump inhibitor",
	"pantoprazole":           "Proton pump inhibitor",

	// H2 blockers
	"famotidine":             "H2 blocker",
	"ranitidine":             "H2 blocker",
	"cimetidine":             "H2 blocker",

	// Antifungals
	"fluconazole":            "Azole antifungal",
	"itraconazole":           "Azole antifungal",
	"ketoconazole":           "Azole antifungal",
	"voriconazole":           "Azole antifungal",
	"terbinafine":            "Antifungal",

	// HIV antivirals
	"ritonavir":              "Protease inhibitor",
	"lopinavir":              "Protease inhibitor",

	// Immunosuppressants
	"cyclosporine":           "Calcineurin inhibitor",
	"tacrolimus":             "Calcineurin inhibitor",

	// Bisphosphonates
	"alendronate":            "Bisphosphonate",
	"risedronate":            "Bisphosphonate",

	// Muscle relaxants
	"cyclobenzaprine":         "Muscle relaxant",
	"baclofen":                "Muscle relaxant",
	"tizanidine":             "Muscle relaxant",
}

// interactionKey returns a canonical key for a pair of drug classes.
func interactionKey(a, b string) string {
	if a < b {
		return a + "|" + b
	}
	return b + "|" + a
}

// interactionRule describes a drug-drug interaction between two classes.
type interactionRule struct {
	Severity    Severity
	Description string
}

// interactionMap holds pairwise drug-class interaction rules.
// Keys are `classA|classB` where A < B alphabetically.
var interactionMap = map[string]interactionRule{
	// Cardiovascular
	interactionKey("ACE inhibitor", "Potassium supplement"): {
		SevMajor, "Concurrent use increases risk of hyperkalemia. Monitor serum potassium closely.",
	},
	interactionKey("ACE inhibitor", "Potassium-sparing diuretic"): {
		SevMajor, "Additive hyperkalemic effect. Avoid combination or monitor potassium rigorously.",
	},
	interactionKey("ARB", "Potassium supplement"): {
		SevMajor, "Concurrent use increases risk of hyperkalemia. Monitor serum potassium closely.",
	},
	interactionKey("ARB", "Potassium-sparing diuretic"): {
		SevMajor, "Additive hyperkalemic effect. Avoid combination or monitor potassium rigorously.",
	},
	interactionKey("ACE inhibitor", "NSAID"): {
		SevModerate, "NSAIDs may reduce antihypertensive effect of ACE inhibitors and increase risk of renal impairment.",
	},
	interactionKey("ARB", "NSAID"): {
		SevModerate, "NSAIDs may reduce antihypertensive effect of ARBs and increase risk of renal impairment.",
	},

	// Bleeding risk
	interactionKey("NSAID", "Anticoagulant"): {
		SevContraindicated, "Combined use significantly increases bleeding risk. Avoid unless benefits clearly outweigh risks with close monitoring.",
	},
	interactionKey("NSAID", "Antiplatelet"): {
		SevContraindicated, "Combined use significantly increases bleeding risk. Avoid unless benefits clearly outweigh risks with close monitoring.",
	},
	interactionKey("Anticoagulant", "Antiplatelet"): {
		SevMajor, "Dual antithrombotic therapy increases bleeding risk. Use only with strong indication and monitoring.",
	},
	interactionKey("NSAID", "Corticosteroid"): {
		SevMajor, "Additive GI ulceration and bleeding risk. Consider gastroprotection.",
	},
	interactionKey("SSRI", "NSAID"): {
		SevModerate, "SSRIs may impair platelet function. Concurrent NSAID use increases GI bleeding risk.",
	},
	interactionKey("SSRI", "Anticoagulant"): {
		SevModerate, "SSRIs may impair platelet function. Increased bleeding risk with anticoagulants.",
	},

	// CNS depression
	interactionKey("Opioid", "Benzodiazepine"): {
		SevContraindicated, "Combined use causes profound sedation, respiratory depression, coma, and death. Avoid unless no alternatives exist.",
	},
	interactionKey("Opioid", "Muscle relaxant"): {
		SevMajor, "Additive CNS and respiratory depression. Use with extreme caution.",
	},
	interactionKey("Benzodiazepine", "Muscle relaxant"): {
		SevMajor, "Additive CNS depression. Increased risk of sedation and respiratory compromise.",
	},
	interactionKey("Opioid", "Anticonvulsant"): {
		SevModerate, "Additive CNS depression. Monitor for excessive sedation.",
	},

	// Serotonin syndrome
	interactionKey("SSRI", "MAOI"): {
		SevContraindicated, "Risk of serotonin syndrome — potentially fatal. Allow at least 14-day washout between drugs.",
	},
	interactionKey("SNRI", "MAOI"): {
		SevContraindicated, "Risk of serotonin syndrome — potentially fatal. Allow at least 14-day washout between drugs.",
	},
	interactionKey("SSRI", "SNRI"): {
		SevMajor, "Additive serotonergic effects increase risk of serotonin syndrome. Use with caution.",
	},
	interactionKey("Opioid", "MAOI"): {
		SevContraindicated, "Risk of serotonin syndrome with serotonergic opioids (tramadol, meperidine, methadone).",
	},
	interactionKey("Opioid", "SSRI"): {
		SevModerate, "Tramadol and methadone with SSRIs increase serotonin syndrome risk. Monitor closely.",
	},

	// Statin interactions
	interactionKey("Statin", "Macrolide antibiotic"): {
		SevMajor, "Macrolides inhibit statin metabolism, increasing rhabdomyolysis risk. Avoid simvastatin; consider alternative statin or antibiotic.",
	},
	interactionKey("Statin", "Azole antifungal"): {
		SevMajor, "Azole antifungals inhibit statin metabolism, increasing rhabdomyolysis risk. Avoid simvastatin/lovastatin.",
	},
	interactionKey("Statin", "Protease inhibitor"): {
		SevContraindicated, "Protease inhibitors severely increase statin levels. Avoid simvastatin/lovastatin; use atorvastatin/rosuvastatin with caution.",
	},
	interactionKey("Statin", "Calcineurin inhibitor"): {
		SevMajor, "Increased statin exposure and myopathy risk. Use lowest effective statin dose.",
	},

	// QT prolongation
	interactionKey("Fluoroquinolone antibiotic", "Antiarrhythmic"): {
		SevMajor, "Additive QT prolongation risk. Avoid combination if possible; monitor ECG.",
	},
	interactionKey("Macrolide antibiotic", "Antiarrhythmic"): {
		SevMajor, "Additive QT prolongation risk. Avoid combination if possible; monitor ECG.",
	},
	interactionKey("Azole antifungal", "Statin"): {
		SevMajor, "Azole antifungals inhibit statin metabolism, increasing rhabdomyolysis risk.",
	},

	// Lithium
	interactionKey("Mood stabilizer", "NSAID"): {
		SevMajor, "NSAIDs decrease lithium clearance, increasing lithium levels and toxicity risk. Monitor lithium levels.",
	},
	interactionKey("Mood stabilizer", "Thiazide diuretic"): {
		SevMajor, "Thiazides decrease lithium clearance, increasing lithium toxicity risk. Monitor lithium levels.",
	},
	interactionKey("Mood stabilizer", "ACE inhibitor"): {
		SevModerate, "ACE inhibitors may increase lithium levels. Monitor lithium levels.",
	},

	// Methotrexate
	interactionKey("Antimetabolite", "NSAID"): {
		SevContraindicated, "NSAIDs reduce methotrexate clearance, causing severe bone marrow suppression and GI toxicity. Avoid concurrent use.",
	},
	interactionKey("Antimetabolite", "Penicillin antibiotic"): {
		SevModerate, "Penicillins may decrease methotrexate clearance, increasing toxicity risk.",
	},
	interactionKey("Antimetabolite", "Sulfonamide antibiotic"): {
		SevContraindicated, "Trimethoprim-sulfamethoxazole with methotrexate may cause severe bone marrow suppression.",
	},

	// Digoxin
	interactionKey("Cardiac glycoside", "Macrolide antibiotic"): {
		SevMajor, "Macrolides may increase digoxin absorption, raising digoxin toxicity risk. Monitor digoxin levels.",
	},
	interactionKey("Cardiac glycoside", "Loop diuretic"): {
		SevModerate, "Diuretic-induced hypokalemia increases digoxin toxicity risk. Monitor potassium and digoxin levels.",
	},

	// Potassium
	interactionKey("Potassium supplement", "Potassium-sparing diuretic"): {
		SevMajor, "Additive hyperkalemic effect. Avoid combination or monitor potassium rigorously.",
	},

	// Theophylline
	interactionKey("Xanthine derivative", "Fluoroquinolone antibiotic"): {
		SevMajor, "Fluoroquinolones (especially ciprofloxacin) inhibit theophylline metabolism, increasing toxicity risk. Monitor theophylline levels.",
	},

	// Antidiabetics
	interactionKey("Sulfonylurea", "Sulfonamide antibiotic"): {
		SevModerate, "Sulfonamides may potentiate sulfonylurea hypoglycemic effect. Monitor blood glucose.",
	},
	interactionKey("Thiazolidinedione", "Insulin"): {
		SevMajor, "Combined use increases fluid retention and heart failure risk. Monitor for edema and CHF symptoms.",
	},

	// Thyroid
	interactionKey("Thyroid hormone", "Calcium channel blocker"): {
		SevMinor, "Calcium may reduce levothyroxine absorption. Separate doses by at least 4 hours.",
	},
	interactionKey("Thyroid hormone", "Proton pump inhibitor"): {
		SevModerate, "PPIs reduce gastric acid, decreasing levothyroxine absorption. Monitor TSH.",
	},

	// Anticonvulsants
	interactionKey("Anticonvulsant", "Anticoagulant"): {
		SevMajor, "Several anticonvulsants (phenytoin, carbamazepine) induce warfarin metabolism, reducing anticoagulant effect.",
	},
}

// allergyDrugClassMap maps common allergy substances to drug classes for cross-referencing.
// These are the allergy NAMES a patient might report (e.g., "penicillin").
var allergyDrugClassMap = map[string]string{
	"penicillin":         "Penicillin antibiotic",
	"amoxicillin":        "Penicillin antibiotic",
	"ampicillin":         "Penicillin antibiotic",
	"sulfa":              "Sulfonamide antibiotic",
	"sulfonamide":        "Sulfonamide antibiotic",
	"nsaid":              "NSAID",
	"aspirin":            "NSAID",
	"ibuprofen":          "NSAID",
	"naproxen":           "NSAID",
	"codeine":            "Opioid",
	"morphine":           "Opioid",
	"opioid":             "Opioid",
	"statin":             "Statin",
	"contrast dye":       "Radiocontrast media",
	"latex":              "Latex",
	"egg":                "Egg-derived product",
	"soy":                "Soy-derived product",
	"cephalosporin":      "Cephalosporin antibiotic",
	"cefazolin":          "Cephalosporin antibiotic",
	"ceftriaxone":        "Cephalosporin antibiotic",
	"macrolide":          "Macrolide antibiotic",
	"erythromycin":       "Macrolide antibiotic",
	"tetracycline":       "Tetracycline antibiotic",
	"doxycycline":        "Tetracycline antibiotic",
	"quinolone":          "Fluoroquinolone antibiotic",
	"ciprofloxacin":      "Fluoroquinolone antibiotic",
	"acetaminophen":      "Analgesic",
	"paracetamol":        "Analgesic",
	"valproic acid":      "Anticonvulsant",
	"phenytoin":          "Anticonvulsant",
	"carbamazepine":      "Anticonvulsant",
	"insulin":            "Insulin",
	"metformin":          "Biguanide",
}

// getDrugClass returns the class for a drug name (case-insensitive), or empty string.
func getDrugClass(name string) string {
	return drugClassMap[strings.ToLower(strings.TrimSpace(name))]
}

// getClassPair returns sorted (classA, classB) pair as a canonical key for lookup.
func (lc *LocalChecker) getInteraction(drugA, drugB string) *InteractionResult {
	classA := getDrugClass(drugA)
	classB := getDrugClass(drugB)
	if classA == "" || classB == "" {
		return nil
	}
	if classA == classB {
		return nil // same class, no interaction with itself
	}

	key := interactionKey(classA, classB)
	rule, ok := interactionMap[key]
	if !ok {
		return nil
	}
	return &InteractionResult{
		Severity:    rule.Severity,
		Description: rule.Description,
		Drugs:       []string{drugA, drugB},
	}
}

// CheckDrugDrug evaluates all unique pairwise drug interactions.
func (lc *LocalChecker) CheckDrugDrug(drugs []string) ([]InteractionResult, error) {
	if len(drugs) < 2 {
		return nil, nil
	}

	seen := make(map[string]bool)
	var results []InteractionResult

	for i := 0; i < len(drugs); i++ {
		for j := i + 1; j < len(drugs); j++ {
			a := strings.TrimSpace(drugs[i])
			b := strings.TrimSpace(drugs[j])
			if a == "" || b == "" {
				continue
			}

			// Deduplicate by canonical drug pair key
			dpKey := pairKey(a, b)
			if seen[dpKey] {
				continue
			}
			seen[dpKey] = true

			if result := lc.getInteraction(a, b); result != nil {
				results = append(results, *result)
			}
		}
	}
	return results, nil
}

// CheckDrugAllergy checks a drug against a patient's allergy list.
// Each allergy substance is matched against the drug's class.
func (lc *LocalChecker) CheckDrugAllergy(drug string, allergies []string) ([]InteractionResult, error) {
	if drug == "" || len(allergies) == 0 {
		return nil, nil
	}

	drugName := strings.TrimSpace(drug)
	drugClass := getDrugClass(drugName)
	if drugClass == "" {
		return nil, nil
	}

	seen := make(map[string]bool)
	var results []InteractionResult

	for _, allergy := range allergies {
		allergyText := strings.ToLower(strings.TrimSpace(allergy))
		if allergyText == "" || allergyText == "active" {
			continue
		}

		// Check direct allergy-to-drug-class match
		for allergySubstance, allergyClass := range allergyDrugClassMap {
			if strings.Contains(allergyText, allergySubstance) {
				if match := checkAllergyDrugMatch(drugClass, allergyClass, drugName, allergy); match != nil {
					if !seen[match.Description] {
						seen[match.Description] = true
						results = append(results, *match)
					}
				}
			}
		}

		// Also check if the allergy text itself matches a drug class
		allergyClass := drugClassMap[allergyText]
		if allergyClass != "" {
			if match := checkAllergyDrugMatch(drugClass, allergyClass, drugName, allergy); match != nil {
				if !seen[match.Description] {
					seen[match.Description] = true
					results = append(results, *match)
				}
			}
		}
	}
	return results, nil
}

// checkAllergyDrugMatch returns an InteractionResult if the drug class matches an allergy.
func checkAllergyDrugMatch(drugClass, allergyClass, drugName, allergy string) *InteractionResult {
	if drugClass == allergyClass {
		return &InteractionResult{
			Severity:    SevContraindicated,
			Description: fmt.Sprintf("Patient has reported allergy to %s (%s class). %s is also in the %s class and should be avoided.", allergy, allergyClass, drugName, drugClass),
			Drugs:       []string{drugName},
		}
	}

	// Check cross-reactivity: penicillin allergy → cephalosporin caution
	if allergyClass == "Penicillin antibiotic" && drugClass == "Cephalosporin antibiotic" {
		return &InteractionResult{
			Severity:    SevModerate,
			Description: fmt.Sprintf("Patient has penicillin allergy. %s (cephalosporin) has ~1-10%% cross-reactivity risk. Use with caution.", drugName),
			Drugs:       []string{drugName},
		}
	}

	// Sulfonamide antibiotic allergy → sulfonylurea caution
	if allergyClass == "Sulfonamide antibiotic" && drugClass == "Sulfonylurea" {
		return &InteractionResult{
			Severity:    SevModerate,
			Description: fmt.Sprintf("Patient has sulfonamide allergy. %s (sulfonylurea) has potential cross-reactivity. Monitor closely.", drugName),
			Drugs:       []string{drugName},
		}
	}

	return nil
}

// pairKey returns a canonical key for a drug pair (alphabetically sorted, lowercase).
func pairKey(a, b string) string {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))
	if a < b {
		return a + "|" + b
	}
	return b + "|" + a
}

// NewLocalChecker creates a new LocalChecker instance.
func NewLocalChecker() *LocalChecker {
	return &LocalChecker{}
}
