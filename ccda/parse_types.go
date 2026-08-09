package ccda

import "time"

// =============================================================================
// Parsed C-CDA types — representing extracted clinical data after parsing
// =============================================================================

// ParsedCCD holds all clinical data extracted from a C-CDA Continuity of Care Document.
type ParsedCCD struct {
	Patient       ParsedPatient
	Allergies     []ParsedAllergy
	Medications   []ParsedMedication
	Problems      []ParsedProblem
	Vitals        []ParsedVitalSign
	Immunizations []ParsedImmunization
	Procedures    []ParsedProcedure
	Results       []ParsedResult
	SocialHistory []ParsedSocialHistory
	FamilyHistory []ParsedFamilyHistory
	Encounters    []ParsedEncounter
}

// ParsedPatient holds patient-level demographics extracted from the CCD header.
type ParsedPatient struct {
	FirstName    string
	LastName     string
	MiddleName   string
	DOB          string // YYYYMMDD
	Gender       string // M/F/U
	MRN          string
	AddressLine1 string
	AddressLine2 string
	City         string
	State        string
	Zip          string
	Phone        string
}

// ParsedAllergy holds an extracted allergy entry.
type ParsedAllergy struct {
	Substance     string
	Reaction      string
	Severity      string // mild/moderate/severe/life threatening
	Status        string // active/inactive
	OnsetDate     string // YYYYMMDD
	Code          string // SNOMED/NDC/RxNorm code
	CodeSystem    string
}

// ParsedMedication holds an extracted medication entry.
type ParsedMedication struct {
	DrugName     string
	Dosage       string
	Frequency    string
	SIG          string
	StartDate    string // YYYYMMDD
	EndDate      string // YYYYMMDD
	Status       string // active/completed
	RxNormCode   string
	NDCCode      string
}

// ParsedProblem holds an extracted problem/condition entry.
type ParsedProblem struct {
	Condition    string
	ICD10Code    string
	OnsetDate    string // YYYYMMDD
	Status       string // active/resolved/inactive
	IsChronic    bool
}

// ParsedVitalSign holds an extracted vital sign organization entry.
type ParsedVitalSign struct {
	Date      time.Time
	Systolic  *float64
	Diastolic *float64
	HeartRate *float64
	RespRate  *float64
	TempF     *float64
	O2Sat     *float64
	HeightCm  *float64
	WeightKg  *float64
	BMI       *float64
}

// ParsedImmunization holds an extracted immunization entry.
type ParsedImmunization struct {
	Vaccine     string
	CVXCode     string
	DateGiven   string // YYYYMMDD
	LotNumber   string
	Manufacturer string
}

// ParsedProcedure holds an extracted surgical/medical procedure entry.
type ParsedProcedure struct {
	Procedure    string
	Date         string // YYYYMMDD
	CPTCode      string
	SNOMEDCode   string
}

// ParsedResult holds an extracted lab/diagnostic result entry.
type ParsedResult struct {
	TestName       string
	LOINCCode      string
	ResultValue    string
	ResultUnit     string
	ReferenceRange string
	ResultDate     string // YYYYMMDD
	AbnormalFlag   string // H/L/N
}

// ParsedSocialHistory holds an extracted social history observation.
type ParsedSocialHistory struct {
	Category      string // smoking/alcohol/drug/occupation/etc.
	Value         string
	Detail        string
	EffectiveDate string // YYYYMMDD
}

// ParsedFamilyHistory holds an extracted family health history entry.
type ParsedFamilyHistory struct {
	Relationship   string
	ConditionName  string
	ICD10Code      string
	OnsetAge       int32
	Deceased       bool
}

// ParsedEncounter holds an extracted encounter/visit entry.
type ParsedEncounter struct {
	Date        string // YYYYMMDD
	EncounterType string
	ProviderName  string
	FacilityName  string
}
