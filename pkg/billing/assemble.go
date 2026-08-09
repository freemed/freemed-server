// Package billing provides claim assembly and encoding for the FreeMED EMR server.
// It bridges the X12 837 encoder (internal/billing/x12) with database queries
// so that claims can be generated from encounter (procrec) data.
package billing

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/freemed/freemed-server/internal/billing/x12"
)

// ProcrecRow holds a row from the procrec table.
type ProcRecRow struct {
	ID            int64
	ProcPatient   int64
	ProcCPT       int64
	ProcCPTMod    int64
	ProcCPTMod2   int64
	ProcCPTMod3   int64
	ProcDiag1     int64
	ProcDiag2     int64
	ProcDiag3     int64
	ProcDiag4     int64
	ProcDiagSet   string
	ProcCharges   float64
	ProcUnits     float64
	ProcVoucher   sql.NullString
	ProcPhysician int64
	ProcDt        time.Time
	ProcPos       int64
	ProcBilled    int64
}

// PatientRow holds patient demographic data.
type PatientRow struct {
	ID           int64
	PtLName      string
	PtFName      string
	PtMName      sql.NullString
	PtDOB        sql.NullTime
	PtSex        string
	PtID         string
	PtPCP        int64
	SSN          sql.NullString
	PtPrimaryFacility int64
}

// PatientAddressRow holds patient address data.
type PatientAddressRow struct {
	Line1  sql.NullString
	City   sql.NullString
	State  sql.NullString
	Postal sql.NullString
	Zip    sql.NullString
}

// CoverageRow holds patient coverage (insurance) data.
type CoverageRow struct {
	InsuranceCompany int64
	CoverageType     int64
	PolicyNumber     string
	GroupNumber      string
	PrimaryCoverage  bool
}

// PhysicianRow holds physician/provider data.
type PhysicianRow struct {
	ID       int64
	PhyLName string
	PhyFName string
	PhyMName string
	PhyNPI   string
	PhyUPIN  string
}

// FacilityRow holds facility data.
type FacilityRow struct {
	ID       int64
	PsrName  string
	PsrAddr1 sql.NullString
	PsrCity  sql.NullString
	PsrState sql.NullString
	PsrZip   sql.NullString
	PsrPhone sql.NullString
	PsrEIN   sql.NullString
	PsrNPI   sql.NullString
	PsrX12ID sql.NullString
}

// InscoRow holds insurance company data.
type InscoRow struct {
	ID         int64
	InscoName  string
	InscoAddr1 string
	InscoCity  string
	InscoState string
	InscoZip   string
	InscoPhone string
	InscoX12ID string
}

// CPTRow holds CPT code data.
type CPTRow struct {
	ID        int64
	Abbrev    string
	CPTNameInt sql.NullString
}

// ICDRow holds ICD-9/ICD-10 code data.
type ICDRow struct {
	ID         int64
	ICD9Code   string
	ICD10Code  sql.NullString
	ICD9Descrip string
}

// CPTModRow holds CPT modifier data.
type CPTModRow struct {
	ID    int64
	CPTMod string
}

// AssembleAndEncodeProfessionalClaim queries the database for encounter data
// identified by procedureIDs, builds a Claim837P, and returns the encoded X12 837P string.
//
//   - db: active database/sql connection pool
//   - procedureIDs: list of procrec.id values to include in the claim
//   - facilityID: id from the facility table for the billing provider
func AssembleAndEncodeProfessionalClaim(db *sql.DB, procedureIDs []int64, facilityID int64) (string, error) {
	if len(procedureIDs) == 0 {
		return "", fmt.Errorf("billing: at least one procedure ID is required")
	}

	ctx := db // for brevity; all queries use the same db

	// ---- Step 1: Fetch all procedure rows ----
	procRows, err := fetchProcRows(db, procedureIDs)
	if err != nil {
		return "", fmt.Errorf("billing: fetching procedures: %w", err)
	}
	if len(procRows) == 0 {
		return "", fmt.Errorf("billing: no procedures found for IDs %v", procedureIDs)
	}

	// Determine the patient ID from the first procedure
	patientID := procRows[0].ProcPatient
	// Determine the voucher (claim tracking number)
	voucher := ""
	for _, pr := range procRows {
		if pr.ProcVoucher.Valid && pr.ProcVoucher.String != "" {
			voucher = pr.ProcVoucher.String
			break
		}
	}
	if voucher == "" {
		voucher = fmt.Sprintf("CLM-%d", time.Now().Unix())
	}

	// Validate all procedures belong to the same patient
	for _, pr := range procRows {
		if pr.ProcPatient != patientID {
			return "", fmt.Errorf("billing: all procedures must belong to the same patient (found %d and %d)", patientID, pr.ProcPatient)
		}
	}

	// ---- Step 2: Fetch patient data ----
	patient, err := fetchPatientRow(db, patientID)
	if err != nil {
		return "", fmt.Errorf("billing: fetching patient: %w", err)
	}

	// Fetch patient address
	patientAddr, err := fetchPatientAddress(db, patientID)
	if err != nil {
		return "", fmt.Errorf("billing: fetching patient address: %w", err)
	}

	// ---- Step 3: Fetch primary coverage ----
	coverage, err := fetchPrimaryCoverage(db, patientID)
	if err != nil {
		return "", fmt.Errorf("billing: fetching coverage: %w", err)
	}

	// ---- Step 4: Fetch facility (billing provider / submitter) ----
	facility, err := fetchFacility(db, facilityID)
	if err != nil {
		return "", fmt.Errorf("billing: fetching facility: %w", err)
	}

	// ---- Step 5: Fetch rendering provider (from first procedure's physician) ----
	renderingPhysicianID := procRows[0].ProcPhysician
	renderingPhys, err := fetchPhysician(db, renderingPhysicianID)
	if err != nil {
		return "", fmt.Errorf("billing: fetching rendering physician: %w", err)
	}

	// ---- Step 6: Fetch insurance company (receiver) ----
	var receiverName, receiverID string
	if coverage != nil {
		insco, err := fetchInsco(db, coverage.InsuranceCompany)
		if err != nil {
			return "", fmt.Errorf("billing: fetching insurance company: %w", err)
		}
		receiverName = insco.InscoName
		if insco.InscoX12ID != "" {
			receiverID = insco.InscoX12ID
		} else {
			receiverID = fmt.Sprintf("%d", insco.ID) // fallback
		}
	} else {
		receiverName = "UNKNOWN PAYER"
		receiverID = "99999"
	}

	// ---- Step 7: Fetch CPT codes, modifiers, ICD codes ----
	cptCache := make(map[int64]string)
	modCache := make(map[int64]string)
	icdCache := make(map[int64]*ICDRow)

	needCPTs := make(map[int64]bool)
	needMods := make(map[int64]bool)
	needICDs := make(map[int64]bool)

	for _, pr := range procRows {
		if pr.ProcCPT != 0 {
			needCPTs[pr.ProcCPT] = true
		}
		if pr.ProcCPTMod != 0 {
			needMods[pr.ProcCPTMod] = true
		}
		if pr.ProcCPTMod2 != 0 {
			needMods[pr.ProcCPTMod2] = true
		}
		if pr.ProcCPTMod3 != 0 {
			needMods[pr.ProcCPTMod3] = true
		}
		if pr.ProcDiag1 != 0 {
			needICDs[pr.ProcDiag1] = true
		}
		if pr.ProcDiag2 != 0 {
			needICDs[pr.ProcDiag2] = true
		}
		if pr.ProcDiag3 != 0 {
			needICDs[pr.ProcDiag3] = true
		}
		if pr.ProcDiag4 != 0 {
			needICDs[pr.ProcDiag4] = true
		}
	}

	for cptID := range needCPTs {
		cpt, err := fetchCPTCode(db, cptID)
		if err != nil {
			return "", fmt.Errorf("billing: fetching CPT %d: %w", cptID, err)
		}
		if cpt != nil {
			cptCache[cptID] = cpt.Abbrev
		}
	}

	for modID := range needMods {
		mod, err := fetchCPTMod(db, modID)
		if err != nil {
			return "", fmt.Errorf("billing: fetching modifier %d: %w", modID, err)
		}
		if mod != nil {
			modCache[modID] = mod.CPTMod
		}
	}

	for icdID := range needICDs {
		icd, err := fetchICDCode(db, icdID)
		if err != nil {
			return "", fmt.Errorf("billing: fetching ICD %d: %w", icdID, err)
		}
		if icd != nil {
			icdCache[icdID] = icd
		}
	}

	// ---- Step 8: Resolve ICD codes into display strings (prefer ICD-10, fallback to ICD-9) ----
	diagnosisCodes := make([]string, 0)
	diagIndexMap := make(map[int64]int) // maps icd9.id → 1-based position in diagnosisCodes

	// We collect unique diagnoses from procdiag1-4 across all procedures
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
				diagIndexMap[diagID] = len(diagnosisCodes) // 1-based
			}
		}
	}

	// If no diagnoses found, add a placeholder
	if len(diagnosisCodes) == 0 {
		diagnosisCodes = append(diagnosisCodes, "Z00.00")
	}

	// ---- Step 9: Build service lines ----
	serviceLines := make([]x12.ServiceLineInfo, 0, len(procRows))
	var totalCharges float64
	var minDate, maxDate time.Time
	firstDate := true

	for _, pr := range procRows {
		// Resolve CPT code
		cptCode := cptCache[pr.ProcCPT]
		if cptCode == "" {
			cptCode = "ZZZZZ" // placeholder for missing code
		}

		// Resolve modifiers
		mod1 := ""
		mod2 := ""
		mod3 := ""
		if pr.ProcCPTMod != 0 {
			mod1 = modCache[pr.ProcCPTMod]
		}
		if pr.ProcCPTMod2 != 0 {
			mod2 = modCache[pr.ProcCPTMod2]
		}
		if pr.ProcCPTMod3 != 0 {
			mod3 = modCache[pr.ProcCPTMod3]
		}

		// Build diagnosis pointers (each diag pointer is 1-based index in diagnosisCodes)
		diagPointers := make([]int, 0)
		for _, diagID := range []int64{pr.ProcDiag1, pr.ProcDiag2, pr.ProcDiag3, pr.ProcDiag4} {
			if diagID == 0 {
				continue
			}
			if idx, ok := diagIndexMap[diagID]; ok {
				diagPointers = append(diagPointers, idx)
			}
		}
		// If no diagnosis pointers, default to first diagnosis
		if len(diagPointers) == 0 && len(diagnosisCodes) > 0 {
			diagPointers = []int{1}
		}

		units := int(pr.ProcUnits)
		if units <= 0 {
			units = 1
		}

		serviceLines = append(serviceLines, x12.ServiceLineInfo{
			LineNumber:        len(serviceLines) + 1,
			ProcedureCode:     cptCode,
			Modifier1:         mod1,
			Modifier2:         mod2,
			Modifier3:         mod3,
			Charge:            pr.ProcCharges,
			Units:             units,
			DiagnosisPointers: diagPointers,
			ServiceDate:       pr.ProcDt.Format("20060102"),
		})

		totalCharges += pr.ProcCharges

		if firstDate || pr.ProcDt.Before(minDate) {
			minDate = pr.ProcDt
		}
		if firstDate || pr.ProcDt.After(maxDate) {
			maxDate = pr.ProcDt
		}
		firstDate = false
	}

	// ---- Step 10: Determine relationship to subscriber ----
	// If patient == subscriber (self) or no coverage, use "18" (self)
	relToSub := "18"
	if coverage != nil && coverage.PolicyNumber != "" {
		relToSub = "18" // default to self; could be overridden for dependents
	}

	// ---- Step 11: Build place of service code ----
	pos := "11" // default: office
	if len(procRows) > 0 {
		posCode := procRows[0].ProcPos
		// Simple mapping of common POS codes
		posMap := map[int64]string{
			0:  "11", // default office
			11: "11", // office
			21: "21", // inpatient hospital
			22: "22", // outpatient hospital
			23: "23", // emergency room
			24: "24", // ambulatory surgical center
			12: "12", // home
		}
		if mapped, ok := posMap[posCode]; ok {
			pos = mapped
		} else {
			pos = fmt.Sprintf("%02d", posCode)
		}
	}

	// ---- Step 12: Build subscriber and patient info ----
	subscriberName := patient.PtLName
	subscriberMemberID := patient.PtID // MRN as subscriber member ID fallback
	if coverage != nil && coverage.PolicyNumber != "" {
		subscriberMemberID = coverage.PolicyNumber
	}

	subscriberDOB := ""
	if patient.PtDOB.Valid {
		subscriberDOB = patient.PtDOB.Time.Format("20060102")
	}

	subscriberAddress1 := ""
	subscriberCity := ""
	subscriberState := ""
	subscriberZip := ""
	if patientAddr != nil {
		if patientAddr.Line1.Valid {
			subscriberAddress1 = patientAddr.Line1.String
		}
		if patientAddr.City.Valid {
			subscriberCity = patientAddr.City.String
		}
		if patientAddr.State.Valid {
			subscriberState = patientAddr.State.String
		}
		if patientAddr.Zip.Valid {
			subscriberZip = patientAddr.Zip.String
		} else if patientAddr.Postal.Valid {
			subscriberZip = patientAddr.Postal.String
		}
	}

	subscriberSSN := ""
	if patient.SSN.Valid {
		subscriberSSN = patient.SSN.String
	}

	patientMiddleName := ""
	if patient.PtMName.Valid {
		patientMiddleName = patient.PtMName.String
	}

	// Payer info
	payerName := receiverName
	payerID := receiverID
	if coverage != nil {
		// Use the insurance company name as payer
		payerName = receiverName
		payerID = receiverID
	}

	// ---- Step 13: Build the Claim837P ----
	claim := &x12.Claim837P{
		SubmitterName: facility.PsrName,
		SubmitterID:   facilityIDStr(facility),
		ReceiverName:  receiverName,
		ReceiverID:    receiverID,
		BillingProvider: x12.ProviderInfo{
			Name:    facility.PsrName,
			NPI:     nullStr(facility.PsrNPI),
			TaxID:   nullStr(facility.PsrEIN),
			Address1: nullStr(facility.PsrAddr1),
			City:    nullStr(facility.PsrCity),
			State:   nullStr(facility.PsrState),
			Zip:     nullStr(facility.PsrZip),
			Phone:   nullStr(facility.PsrPhone),
		},
		Subscriber: x12.SubscriberInfo{
			LastName:   subscriberName,
			FirstName:  patient.PtFName,
			MiddleName: patientMiddleName,
			MemberID:   subscriberMemberID,
			SSN:        subscriberSSN,
			DOB:        subscriberDOB,
			Address1:   subscriberAddress1,
			City:       subscriberCity,
			State:      subscriberState,
			Zip:        subscriberZip,
			PayerName:  payerName,
			PayerID:    payerID,
		},
		Patient: x12.PatientInfo{
			LastName:                 patient.PtLName,
			FirstName:                patient.PtFName,
			DOB:                      subscriberDOB,
			RelationshipToSubscriber: relToSub,
			SSN:                      subscriberSSN,
			Address1:                 subscriberAddress1,
			City:                     subscriberCity,
			State:                    subscriberState,
			Zip:                      subscriberZip,
		},
		Claim: x12.ClaimInfo{
			ClaimID:        voucher,
			TotalCharges:   totalCharges,
			PlaceOfService: pos,
			DiagnosisCodes: diagnosisCodes,
			StatementFrom:  minDate.Format("20060102"),
			StatementTo:    maxDate.Format("20060102"),
		},
		ServiceLines: serviceLines,
	}

	// Add rendering provider if we have valid provider data
	if renderingPhys != nil && renderingPhys.PhyNPI != "" {
		renderingMiddle := renderingPhys.PhyMName
		claim.RenderingProvider = &x12.ProviderRefInfo{
			LastName:   renderingPhys.PhyLName,
			FirstName:  renderingPhys.PhyFName,
			MiddleName: renderingMiddle,
			NPI:        renderingPhys.PhyNPI,
		}
	}

	// ---- Step 14: Validate and encode ----
	if err := x12.ValidateClaim(claim); err != nil {
		return "", fmt.Errorf("billing: claim validation failed: %w", err)
	}

	encoded, err := x12.Encode837Professional(claim)
	if err != nil {
		return "", fmt.Errorf("billing: encoding 837P: %w", err)
	}

	_ = ctx
	return string(encoded), nil
}

// ---- Database query helpers ----

func fetchProcRows(db *sql.DB, ids []int64) ([]ProcRecRow, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	query := fmt.Sprintf(`SELECT id, procpatient, proccpt, proccptmod, proccptmod2, proccptmod3,
		procdiag1, procdiag2, procdiag3, procdiag4, procdiagset,
		proccharges, procunits, procvoucher, procphysician, procdt, procpos, procbilled
		FROM procrec WHERE id IN (%s)`, strings.Join(placeholders, ","))
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []ProcRecRow
	for rows.Next() {
		var r ProcRecRow
		err := rows.Scan(&r.ID, &r.ProcPatient, &r.ProcCPT, &r.ProcCPTMod, &r.ProcCPTMod2,
			&r.ProcCPTMod3, &r.ProcDiag1, &r.ProcDiag2, &r.ProcDiag3, &r.ProcDiag4,
			&r.ProcDiagSet, &r.ProcCharges, &r.ProcUnits, &r.ProcVoucher,
			&r.ProcPhysician, &r.ProcDt, &r.ProcPos, &r.ProcBilled)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

func fetchPatientRow(db *sql.DB, id int64) (*PatientRow, error) {
	var r PatientRow
	err := db.QueryRow(`SELECT id, ptlname, ptfname, ptmname, ptdob, ptsex, ptid, ptpcp, ssn, ptprimaryfacility
		FROM patient WHERE id = ?`, id).Scan(
		&r.ID, &r.PtLName, &r.PtFName, &r.PtMName, &r.PtDOB,
		&r.PtSex, &r.PtID, &r.PtPCP, &r.SSN, &r.PtPrimaryFacility)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func fetchPatientAddress(db *sql.DB, patientID int64) (*PatientAddressRow, error) {
	var r PatientAddressRow
	err := db.QueryRow(`SELECT line1, city, stpr, postal, zip
		FROM patient_address WHERE patient = ? AND active = 1 LIMIT 1`, patientID).Scan(
		&r.Line1, &r.City, &r.State, &r.Postal, &r.Zip)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func fetchPrimaryCoverage(db *sql.DB, patientID int64) (*CoverageRow, error) {
	var r CoverageRow
	err := db.QueryRow(`SELECT insurance_company, coverage_type, policy_number, group_number, primary_coverage
		FROM patient_coverage WHERE patient = ? AND primary_coverage = 1 LIMIT 1`, patientID).Scan(
		&r.InsuranceCompany, &r.CoverageType, &r.PolicyNumber, &r.GroupNumber, &r.PrimaryCoverage)
	if err == sql.ErrNoRows {
		// Fallback: get any coverage for this patient
		err = db.QueryRow(`SELECT insurance_company, coverage_type, policy_number, group_number, primary_coverage
			FROM patient_coverage WHERE patient = ? LIMIT 1`, patientID).Scan(
			&r.InsuranceCompany, &r.CoverageType, &r.PolicyNumber, &r.GroupNumber, &r.PrimaryCoverage)
		if err == sql.ErrNoRows {
			return nil, nil // No coverage found
		}
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func fetchFacility(db *sql.DB, id int64) (*FacilityRow, error) {
	var r FacilityRow
	err := db.QueryRow(`SELECT id, psrname, psraddr1, psrcity, psrstate, psrzip, psrphone,
		psrein, psrnpi, psrx12id FROM facility WHERE id = ?`, id).Scan(
		&r.ID, &r.PsrName, &r.PsrAddr1, &r.PsrCity, &r.PsrState, &r.PsrZip,
		&r.PsrPhone, &r.PsrEIN, &r.PsrNPI, &r.PsrX12ID)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func fetchPhysician(db *sql.DB, id int64) (*PhysicianRow, error) {
	if id == 0 {
		return nil, nil
	}
	var r PhysicianRow
	err := db.QueryRow(`SELECT id, phylname, phyfname, phymname, phynpi, phyupin
		FROM physician WHERE id = ?`, id).Scan(
		&r.ID, &r.PhyLName, &r.PhyFName, &r.PhyMName, &r.PhyNPI, &r.PhyUPIN)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func fetchInsco(db *sql.DB, id int64) (*InscoRow, error) {
	var r InscoRow
	err := db.QueryRow(`SELECT id, insconame, inscoaddr1, inscocity, inscostate, inscozip,
		inscophone, inscox12id FROM insco WHERE id = ?`, id).Scan(
		&r.ID, &r.InscoName, &r.InscoAddr1, &r.InscoCity, &r.InscoState,
		&r.InscoZip, &r.InscoPhone, &r.InscoX12ID)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func fetchCPTCode(db *sql.DB, id int64) (*CPTRow, error) {
	var r CPTRow
	err := db.QueryRow(`SELECT id, abbrev, cptnameint FROM cpt WHERE id = ?`, id).Scan(
		&r.ID, &r.Abbrev, &r.CPTNameInt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func fetchCPTMod(db *sql.DB, id int64) (*CPTModRow, error) {
	var r CPTModRow
	err := db.QueryRow(`SELECT id, cptmod FROM cptmod WHERE id = ?`, id).Scan(
		&r.ID, &r.CPTMod)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func fetchICDCode(db *sql.DB, id int64) (*ICDRow, error) {
	var r ICDRow
	err := db.QueryRow(`SELECT id, icd9code, icd10code, icd9descrip FROM icd9 WHERE id = ?`, id).Scan(
		&r.ID, &r.ICD9Code, &r.ICD10Code, &r.ICD9Descrip)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// ---- Helpers ----

func nullStr(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func facilityIDStr(f *FacilityRow) string {
	if f.PsrX12ID.Valid && f.PsrX12ID.String != "" {
		return f.PsrX12ID.String
	}
	if f.PsrEIN.Valid && f.PsrEIN.String != "" {
		return f.PsrEIN.String
	}
	return fmt.Sprintf("%d", f.ID)
}
