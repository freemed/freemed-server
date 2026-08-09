package billing

import (
	"database/sql"
	"fmt"
	"time"

	x12_270 "github.com/freemed/freemed-server/pkg/x12/270"
)

// EligibilityRequestInput holds the assembled data from DB queries.
type EligibilityRequestInput struct {
	// Provider info
	ProviderNPI   string
	ProviderTaxID string
	ProviderName  string

	// Facility info
	FacilityNPI string

	// Payer info
	PayerID   string
	PayerName string

	// Subscriber (the insured party)
	SubscriberLastName  string
	SubscriberFirstName string
	SubscriberMiddleName string
	SubscriberDOB       string // YYYYMMDD
	SubscriberGender    string // M/F/U
	SubscriberMemberID  string
	SubscriberGroupNo   string
	SubscriberAddr1     string
	SubscriberCity      string
	SubscriberState     string
	SubscriberZip       string

	// Dependent (if patient is not subscriber)
	DependentLastName  string
	DependentFirstName string
	DependentDOB       string // YYYYMMDD
	DependentGender    string // M/F/U
	DependentRel       string // 01=Spouse, 19=Child

	IsSelf bool // true if subscriber is the patient themselves

	// Service types to check
	ServiceTypes []string
}

// AssembleEligibilityRequest builds an X12 270 eligibility inquiry from database records.
// It queries patient demographics, coverage, and insurance company info.
func AssembleEligibilityRequest(db *sql.DB, patientID int64) (*x12_270.EligibilityInquiry270, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	input := &EligibilityRequestInput{}

	// Query patient demographics
	var (
		ptLName, ptFName, ptID, ptSex string
		ptDOB                         sql.NullTime
		ptPCP                         int64
	)
	err := db.QueryRow(`
		SELECT ptlname, ptfname, ptid, ptsex, ptdob, ptpcp
		FROM patient WHERE id = ? AND deleted_at IS NULL
	`, patientID).Scan(&ptLName, &ptFName, &ptID, &ptSex, &ptDOB, &ptPCP)
	if err != nil {
		return nil, fmt.Errorf("patient %d not found: %w", patientID, err)
	}

	// Query primary coverage
	var (
		covInsCo    int64
		covPolicyNo string
		covGroupNo  string
	)
	err = db.QueryRow(`
		SELECT insurance_company, policy_number, group_number
		FROM patient_coverage
		WHERE patient = ? AND primary_coverage = 1 AND active = 'active'
		LIMIT 1
	`, patientID).Scan(&covInsCo, &covPolicyNo, &covGroupNo)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("coverage lookup failed: %w", err)
	}

	// Query insurance company
	var (
		inscoName   string
		inscoPayerID string
	)
	if covInsCo > 0 {
		err = db.QueryRow(`
			SELECT iname, ipayerid FROM insco WHERE id = ?
		`, covInsCo).Scan(&inscoName, &inscoPayerID)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("insurance company lookup failed: %w", err)
		}
	}

	// Query PCP (provider)
	var (
		phyNPI   string
		phyLName string
		phyFName string
		phyUPIN  sql.NullString
	)
	if ptPCP > 0 {
		err = db.QueryRow(`
			SELECT phynpi, phylname, phyfname, phyupin
			FROM physician WHERE id = ?
		`, ptPCP).Scan(&phyNPI, &phyLName, &phyFName, &phyUPIN)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("physician lookup failed: %w", err)
		}
	}

	// Query patient's address
	var (
		addr1, addrCity, addrState sql.NullString
		addrZip                     sql.NullString
	)
	err = db.QueryRow(`
		SELECT addr_line1, city, state, postal_code
		FROM patient_address WHERE patient = ? AND active = 'active'
		LIMIT 1
	`, patientID).Scan(&addr1, &addrCity, &addrState, &addrZip)
	if err != nil && err != sql.ErrNoRows {
		// Non-fatal — address is optional
	}

	// Populate input
	input.SubscriberLastName = ptLName
	input.SubscriberFirstName = ptFName
	input.SubscriberMemberID = covPolicyNo
	input.SubscriberGroupNo = covGroupNo
	if ptDOB.Valid {
		input.SubscriberDOB = ptDOB.Time.Format("20060102")
	}
	input.SubscriberGender = normalizeGender(ptSex)
	input.IsSelf = true // Subscriber is the patient
	if addr1.Valid {
		input.SubscriberAddr1 = addr1.String
	}
	if addrCity.Valid {
		input.SubscriberCity = addrCity.String
	}
	if addrState.Valid {
		input.SubscriberState = addrState.String
	}
	if addrZip.Valid {
		input.SubscriberZip = addrZip.String
	}

	input.ProviderNPI = phyNPI
	input.ProviderName = fmt.Sprintf("%s %s", phyFName, phyLName)
	if phyUPIN.Valid {
		input.ProviderTaxID = phyUPIN.String
	}

	input.PayerName = inscoName
	input.PayerID = inscoPayerID

	// Default service types
	input.ServiceTypes = []string{"30"} // Health Benefit Plan Coverage

	return assemble270FromInput(input), nil
}

// assemble270FromInput builds the 270 inquiry from assembled input data.
func assemble270FromInput(input *EligibilityRequestInput) *x12_270.EligibilityInquiry270 {
	now := time.Now()
	dateStr := now.Format("20060102")

	inq := &x12_270.EligibilityInquiry270{
		SenderID:        padOrTrunc(input.ProviderTaxID, 15),
		ReceiverID:      padOrTrunc(input.PayerID, 15),
		TransactionDate: dateStr,
		TransactionNo:   fmt.Sprintf("%d", now.UnixNano()%1000000),
		InfoSource: x12_270.InfoSource{
			EntityID:     padOrTrunc(input.PayerID, 15),
			EntityIDQual: "PI",
			Name:         truncate(input.PayerName, 60),
		},
		InfoReceiver: x12_270.InfoReceiver{
			EntityID:     padOrTrunc(input.ProviderNPI, 15),
			EntityIDQual: "XX",
			Name:         truncate(input.ProviderName, 60),
		},
		Subscriber: x12_270.SubscriberInfo{
			LastName:    truncate(input.SubscriberLastName, 35),
			FirstName:   truncate(input.SubscriberFirstName, 25),
			MiddleName:  truncate(input.SubscriberMiddleName, 25),
			IDQualifier: "MI",
			MemberID:    truncate(input.SubscriberMemberID, 20),
			GroupNumber: truncate(input.SubscriberGroupNo, 30),
			DOB:         input.SubscriberDOB,
			Gender:      input.SubscriberGender,
			Address1:    truncate(input.SubscriberAddr1, 55),
			City:        truncate(input.SubscriberCity, 30),
			State:       truncate(input.SubscriberState, 2),
			Zip:         truncate(input.SubscriberZip, 15),
		},
		ServiceTypes: input.ServiceTypes,
	}

	if !input.IsSelf && input.DependentLastName != "" {
		inq.Dependent = &x12_270.DependentInfo{
			LastName:     truncate(input.DependentLastName, 35),
			FirstName:    truncate(input.DependentFirstName, 25),
			DOB:          input.DependentDOB,
			Gender:       input.DependentGender,
			Relationship: input.DependentRel,
		}
	}

	return inq
}

// normalizeGender converts FreeMED sex codes to X12 gender codes.
func normalizeGender(sex string) string {
	switch sex {
	case "m", "M":
		return "M"
	case "f", "F":
		return "F"
	default:
		return "U"
	}
}

// truncate truncates a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// padOrTrunc pads or truncates a string to the given length.
func padOrTrunc(s string, length int) string {
	if len(s) >= length {
		return s[:length]
	}
	pad := make([]byte, length-len(s))
	for i := range pad {
		pad[i] = ' '
	}
	return s + string(pad)
}
