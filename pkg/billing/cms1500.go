package billing

import (
	"database/sql"
	"fmt"
)

// CMS1500Data holds all fields needed for a CMS-1500 (02/12) claim form.
type CMS1500Data struct {
	// Box 1: Insurance type
	InsuranceType string // Medicare, Medicaid, etc.

	// Box 1a: Insured's ID number
	InsuredID string

	// Box 2: Patient's name (last, first, middle)
	PatientLastName  string
	PatientFirstName string
	PatientMiddleName string

	// Box 3: Patient DOB, Sex
	PatientDOB string // MM/DD/YYYY
	PatientSex string

	// Box 4: Insured's name
	InsuredLastName  string
	InsuredFirstName string
	InsuredMiddleName string

	// Box 5: Patient's address
	PatientAddress string
	PatientCity    string
	PatientState   string
	PatientZip     string
	PatientPhone   string

	// Box 6: Patient relationship to insured
	Relationship string

	// Box 7: Insured's address
	InsuredAddress string
	InsuredCity    string
	InsuredState   string
	InsuredZip     string
	InsuredPhone   string

	// Box 9: Other insured info
	OtherInsuredName   string
	OtherInsuredPolicy string

	// Box 11: Insured's policy/group number
	PolicyGroup string
	// Box 11a-c: Insured's DOB, Sex, Employer
	InsuredDOB  string

	// Box 17: Referring provider name
	ReferringProvider string

	// Box 17a: Referring provider ID (NPI)
	ReferringProviderNPI string

	// Box 21: Diagnosis codes (up to 12)
	DiagnosisCodes []string

	// Box 23: Prior authorization number
	PriorAuthNumber string

	// Box 24A-J: Service lines
	ServiceLines []CMS1500ServiceLine

	// Box 25: Federal tax ID
	FederalTaxID string

	// Box 26: Patient's account number
	PatientAccountNumber string

	// Box 28: Total charge
	TotalCharge float64

	// Box 30: Balance due (total charge - paid)
	BalanceDue float64

	// Box 31: Rendering provider signature info
	RenderingProvider string

	// Box 32: Service facility location
	FacilityName    string
	FacilityAddress string
	FacilityCity    string
	FacilityState   string
	FacilityZip     string
	FacilityNPI     string

	// Box 33: Billing provider info
	BillingProviderName    string
	BillingProviderAddress string
	BillingProviderCity    string
	BillingProviderState   string
	BillingProviderZip     string
	BillingProviderPhone   string
	BillingProviderNPI     string
}

// CMS1500ServiceLine represents one row in box 24 of the CMS-1500.
type CMS1500ServiceLine struct {
	DateFrom           string // MM/DD/YYYY
	DateTo             string
	PlaceOfService     string
	TypeOfService      string
	ProcedureCode      string
	Modifier1          string
	Modifier2          string
	Modifier3          string
	DiagnosisPointer1  string
	DiagnosisPointer2  string
	DiagnosisPointer3  string
	DiagnosisPointer4  string
	Charge             float64
	Units              int
	EPSDTFamilyPlan    string
	EMG                string
}

// AssembleCMS1500 queries the database and builds a CMS1500Data struct
// for the given procedure IDs and facility.
func AssembleCMS1500(db *sql.DB, procedureIDs []int64, facilityID int64) (*CMS1500Data, error) {
	if len(procedureIDs) == 0 {
		return nil, fmt.Errorf("billing: at least one procedure ID is required")
	}

	// Fetch procedure rows
	procRows, err := fetchProcRows(db, procedureIDs)
	if err != nil {
		return nil, fmt.Errorf("billing: fetching procedures: %w", err)
	}
	if len(procRows) == 0 {
		return nil, fmt.Errorf("billing: no procedures found for IDs %v", procedureIDs)
	}

	patientID := procRows[0].ProcPatient

	// Fetch patient
	patient, err := fetchPatientRow(db, patientID)
	if err != nil {
		return nil, fmt.Errorf("billing: fetching patient: %w", err)
	}

	// Fetch patient address
	patientAddr, _ := fetchPatientAddress(db, patientID)

	// Fetch primary coverage
	coverage, _ := fetchPrimaryCoverage(db, patientID)

	// Fetch facility
	facility, err := fetchFacility(db, facilityID)
	if err != nil {
		return nil, fmt.Errorf("billing: fetching facility: %w", err)
	}

	// Fetch rendering provider
	var renderingPhys *PhysicianRow
	if procRows[0].ProcPhysician != 0 {
		renderingPhys, _ = fetchPhysician(db, procRows[0].ProcPhysician)
	}

	// Fetch ICD codes
	icdCache := make(map[int64]*ICDRow)
	cptCache := make(map[int64]string)
	modCache := make(map[int64]string)

	for _, pr := range procRows {
		if pr.ProcCPT != 0 {
			if _, ok := cptCache[pr.ProcCPT]; !ok {
				cpt, _ := fetchCPTCode(db, pr.ProcCPT)
				if cpt != nil {
					cptCache[pr.ProcCPT] = cpt.Abbrev
				}
			}
		}
		for _, modID := range []int64{pr.ProcCPTMod, pr.ProcCPTMod2, pr.ProcCPTMod3} {
			if modID != 0 {
				if _, ok := modCache[modID]; !ok {
					mod, _ := fetchCPTMod(db, modID)
					if mod != nil {
						modCache[modID] = mod.CPTMod
					}
				}
			}
		}
		for _, diagID := range []int64{pr.ProcDiag1, pr.ProcDiag2, pr.ProcDiag3, pr.ProcDiag4} {
			if diagID != 0 {
				if _, ok := icdCache[diagID]; !ok {
					icd, _ := fetchICDCode(db, diagID)
					if icd != nil {
						icdCache[diagID] = icd
					}
				}
			}
		}
	}

	// Build diagnosis codes
	diagnosisCodes := make([]string, 0)
	diagIndexMap := make(map[int64]int)
	seenDiags := make(map[int64]bool)
	for _, pr := range procRows {
		for _, diagID := range []int64{pr.ProcDiag1, pr.ProcDiag2, pr.ProcDiag3, pr.ProcDiag4} {
			if diagID == 0 || seenDiags[diagID] {
				continue
			}
			seenDiags[diagID] = true
			if icd, ok := icdCache[diagID]; ok {
				code := icd.ICD9Code
				if icd.ICD10Code.Valid && icd.ICD10Code.String != "" {
					code = icd.ICD10Code.String
				}
				diagnosisCodes = append(diagnosisCodes, code)
				diagIndexMap[diagID] = len(diagnosisCodes)
			}
		}
	}

	// Build service lines
	serviceLines := make([]CMS1500ServiceLine, 0, len(procRows))
	var totalCharge float64

	for _, pr := range procRows {
		cptCode := cptCache[pr.ProcCPT]
		if cptCode == "" {
			cptCode = "ZZZZZ"
		}

		mod1 := modCache[pr.ProcCPTMod]
		mod2 := modCache[pr.ProcCPTMod2]
		mod3 := modCache[pr.ProcCPTMod3]

		// Build diagnosis pointers (as characters A-L per CMS-1500)
		var d1, d2, d3, d4 string
		ptrIdx := 0
		for _, diagID := range []int64{pr.ProcDiag1, pr.ProcDiag2, pr.ProcDiag3, pr.ProcDiag4} {
			if diagID == 0 {
				continue
			}
			if idx, ok := diagIndexMap[diagID]; ok && idx <= 12 {
				switch ptrIdx {
				case 0:
					d1 = string(rune('A' + idx - 1))
				case 1:
					d2 = string(rune('A' + idx - 1))
				case 2:
					d3 = string(rune('A' + idx - 1))
				case 3:
					d4 = string(rune('A' + idx - 1))
				}
				ptrIdx++
			}
		}

		units := int(pr.ProcUnits)
		if units <= 0 {
			units = 1
		}

		posCode := fmt.Sprintf("%02d", pr.ProcPos)
		if pr.ProcPos == 0 {
			posCode = "11"
		}

		serviceLines = append(serviceLines, CMS1500ServiceLine{
			DateFrom:          pr.ProcDt.Format("01/02/2006"),
			DateTo:            pr.ProcDt.Format("01/02/2006"),
			PlaceOfService:    posCode,
			ProcedureCode:     cptCode,
			Modifier1:         mod1,
			Modifier2:         mod2,
			Modifier3:         mod3,
			DiagnosisPointer1: d1,
			DiagnosisPointer2: d2,
			DiagnosisPointer3: d3,
			DiagnosisPointer4: d4,
			Charge:            pr.ProcCharges,
			Units:             units,
		})

		totalCharge += pr.ProcCharges
	}

	// Build CMS1500Data
	data := &CMS1500Data{
		InsuranceType:   "CI", // Commercial Insurance - default
		Relationship:    "18", // Self
		DiagnosisCodes:  diagnosisCodes,
		ServiceLines:    serviceLines,
		TotalCharge:     totalCharge,
		BalanceDue:      totalCharge, // Assuming no prior payments
		FederalTaxID:    nullStr(facility.PsrEIN),
		FacilityName:    facility.PsrName,
		FacilityAddress: nullStr(facility.PsrAddr1),
		FacilityCity:    nullStr(facility.PsrCity),
		FacilityState:   nullStr(facility.PsrState),
		FacilityZip:     nullStr(facility.PsrZip),
		FacilityNPI:     nullStr(facility.PsrNPI),
		BillingProviderName:    facility.PsrName,
		BillingProviderAddress: nullStr(facility.PsrAddr1),
		BillingProviderCity:    nullStr(facility.PsrCity),
		BillingProviderState:   nullStr(facility.PsrState),
		BillingProviderZip:     nullStr(facility.PsrZip),
		BillingProviderPhone:   nullStr(facility.PsrPhone),
		BillingProviderNPI:     nullStr(facility.PsrNPI),
	}

	// Patient info
	data.PatientLastName = patient.PtLName
	data.PatientFirstName = patient.PtFName
	if patient.PtMName.Valid {
		data.PatientMiddleName = patient.PtMName.String
	}
	if patient.PtDOB.Valid {
		data.PatientDOB = patient.PtDOB.Time.Format("01/02/2006")
	}
	data.PatientSex = patient.PtSex
	data.InsuredLastName = patient.PtLName
	data.InsuredFirstName = patient.PtFName
	data.PatientAccountNumber = patient.PtID

	// Patient address
	if patientAddr != nil {
		if patientAddr.Line1.Valid {
			data.PatientAddress = patientAddr.Line1.String
			data.InsuredAddress = patientAddr.Line1.String
		}
		if patientAddr.City.Valid {
			data.PatientCity = patientAddr.City.String
			data.InsuredCity = patientAddr.City.String
		}
		if patientAddr.State.Valid {
			data.PatientState = patientAddr.State.String
			data.InsuredState = patientAddr.State.String
		}
		if patientAddr.Zip.Valid {
			data.PatientZip = patientAddr.Zip.String
			data.InsuredZip = patientAddr.Zip.String
		} else if patientAddr.Postal.Valid {
			data.PatientZip = patientAddr.Postal.String
			data.InsuredZip = patientAddr.Postal.String
		}
	}

	// Coverage info
	if coverage != nil {
		data.InsuredID = coverage.PolicyNumber
		data.PolicyGroup = coverage.GroupNumber
		// Get payer info for box 9d
		insco, err := fetchInsco(db, coverage.InsuranceCompany)
		if err == nil && insco != nil {
			data.InsuranceType = insco.InscoName
		}
	}

	// Rendering provider
	if renderingPhys != nil {
		data.RenderingProvider = renderingPhys.PhyFName + " " + renderingPhys.PhyLName
		if data.RenderingProvider == " " {
			data.RenderingProvider = renderingPhys.PhyLName
		}
	}

	// Referring provider (PCP)
	if patient.PtPCP != 0 {
		pcp, err := fetchPhysician(db, patient.PtPCP)
		if err == nil && pcp != nil {
			data.ReferringProvider = pcp.PhyFName + " " + pcp.PhyLName
			if data.ReferringProvider == " " {
				data.ReferringProvider = pcp.PhyLName
			}
			data.ReferringProviderNPI = pcp.PhyNPI
		}
	}

	return data, nil
}
