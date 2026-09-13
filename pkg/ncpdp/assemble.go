package ncpdp

import (
	"database/sql"
	"fmt"
	"time"
)

// AssemblePrescriptionInput queries the FreeMED database for all data needed
// to construct a NewRx message for a given prescription ID. The NCPDP sender
// ID is supplied by the caller (from configuration) because it is assigned
// per-organization by NCPDP and has no safe default.
func AssemblePrescriptionInput(db *sql.DB, prescriptionID int64, senderNCPDPID string) (*PrescriptionInput, error) {
	if db == nil {
		return nil, fmt.Errorf("ncpdp: database connection is nil")
	}

	in := &PrescriptionInput{}

	// Query the prescriptions table
	var (
		drugName, dosage, frequency, quantity, sig, status, pharmacyName string
		refills                                                           int64
		writtenDate                                                       time.Time
		patientID, prescriberID                                           int64
		notes                                                            sql.NullString
	)
	err := db.QueryRow(`
		SELECT drug_name, dosage, frequency, quantity, refills,
			date_written, prescribing_provider, pharmacy, status, notes, patient
		FROM prescriptions
		WHERE id = ? AND deleted_at IS NULL
	`, prescriptionID).Scan(&drugName, &dosage, &frequency, &quantity,
		&refills, &writtenDate, &prescriberID, &pharmacyName, &status, &notes, &patientID)
	if err != nil {
		return nil, fmt.Errorf("ncpdp: prescription %d not found: %w", prescriptionID, err)
	}

	in.DrugName = drugName
	in.Dosage = dosage
	in.Frequency = frequency
	in.Quantity = quantity
	in.QuantityUnit = "EA" // default unit
	in.Refills = fmt.Sprintf("%d", refills)
	in.Substitution = "1" // default: allow generic substitution
	in.WrittenDate = writtenDate.Format("2006-01-02")

	// SIG: combine dosage + frequency + sig
	if sig != "" {
		in.SIG = sig
	} else {
		in.SIG = fmt.Sprintf("%s %s", dosage, frequency)
	}

	// Query patient demographics
	var (
		ptFName, ptLName, ptID, ptSex string
		ptDOB                          sql.NullTime
	)
	err = db.QueryRow(`
		SELECT ptfname, ptlname, ptid, ptsex, ptdob
		FROM patient WHERE id = ? AND deleted_at IS NULL
	`, patientID).Scan(&ptFName, &ptLName, &ptID, &ptSex, &ptDOB)
	if err != nil {
		return nil, fmt.Errorf("ncpdp: patient %d not found: %w", patientID, err)
	}

	in.PatientFName = ptFName
	in.PatientLName = ptLName
	in.PatientMRN = ptID
	in.PatientGender = ptSex
	if ptDOB.Valid {
		in.PatientDOB = ptDOB.Time.Format("2006-01-02")
	}

	// Query patient address
	var (
		addr1, city, state, zip sql.NullString
	)
	err = db.QueryRow(`
		SELECT addr_line1, city, state, postal_code
		FROM patient_address WHERE patient = ? AND active = 'active'
		LIMIT 1
	`, patientID).Scan(&addr1, &city, &state, &zip)
	if err == nil {
		if addr1.Valid {
			in.PatientAddr1 = addr1.String
		}
		if city.Valid {
			in.PatientCity = city.String
		}
		if state.Valid {
			in.PatientState = state.String
		}
		if zip.Valid {
			in.PatientZip = zip.String
		}
	}

	// Query patient phone
	var phoneNum sql.NullString
	err = db.QueryRow(`
		SELECT number FROM phone WHERE patient = ? AND active = 'active'
		LIMIT 1
	`, patientID).Scan(&phoneNum)
	if err == nil && phoneNum.Valid {
		in.PatientPhone = phoneNum.String
	}

	// Query prescriber
	var (
		phyNPI, phyFName, phyLName, phyDEA sql.NullString
	)
	err = db.QueryRow(`
		SELECT phynpi, phyfname, phylname, phydea
		FROM physician WHERE id = ?
	`, prescriberID).Scan(&phyNPI, &phyFName, &phyLName, &phyDEA)
	if err == nil {
		if phyNPI.Valid {
			in.PrescriberNPI = phyNPI.String
		}
		if phyDEA.Valid {
			in.PrescriberDEA = phyDEA.String
		}
		if phyFName.Valid {
			in.PrescriberFName = phyFName.String
		}
		if phyLName.Valid {
			in.PrescriberLName = phyLName.String
		}
	}

	// Query pharmacy (by name matching, or by pharmacy table)
	if pharmacyName != "" {
		var (
			phNCPDPID, phNPI, phPhone, phFax sql.NullString
		)
		err = db.QueryRow(`
			SELECT ncpdp_id, phstate, phcity
			FROM pharmacy WHERE phname = ? LIMIT 1
		`, pharmacyName).Scan(&phNCPDPID, &phNPI, &phPhone)
		if err == nil {
			in.PharmacyName = pharmacyName
			if phNCPDPID.Valid {
				in.PharmacyNCPDPID = phNCPDPID.String
			}
			if phNPI.Valid {
				in.PharmacyNPI = phNPI.String
			}
			if phPhone.Valid {
				in.PharmacyPhone = phPhone.String
			}
			if phFax.Valid {
				in.PharmacyFax = phFax.String
			}
		} else {
			in.PharmacyName = pharmacyName
		}
	}

	// Sender NCPDPID identifies this organization to the pharmacy/Surescripts.
	// Supplied by the caller from configuration — never defaulted.
	in.SenderNCPDPID = senderNCPDPID

	return in, nil
}
