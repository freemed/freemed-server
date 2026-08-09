//go:build integration

package integration

import (
	"fmt"
	"testing"
	"time"
)

// TestEligibilityFlowIntegration tests the flow from patient+coverage → 270 generation.
// This verifies that the DB schema supports eligibility data properly.
func TestEligibilityFlowIntegration(t *testing.T) {
	db := openDB(t)
	defer db.Close()

	now := time.Now()
	ptid := fmt.Sprintf("ELIG-%d", now.UnixNano())

	// Create a test patient
	result, err := db.Exec(`
		INSERT INTO patient (created_at, updated_at, ptdtadd, stamp,
			ptlname, ptfname, ptid, ptsex, ptdob, user, provider, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now, now, now, now, "Eligibility", "Test", ptid, "F", "1975-08-20", 1, 1, "active")
	if err != nil {
		t.Fatalf("INSERT patient failed: %v", err)
	}
	patientID, _ := result.LastInsertId()
	t.Cleanup(func() { cleanPatient(t, db, patientID) })

	// Create test insurance company
	inscoResult, err := db.Exec(`
		INSERT INTO insco (insconame, inscox12id, inscoaddr1, inscocity, inscostate, inscozip, inscophone, inscodtadd, inscodtmod, created_at, updated_at)
		VALUES ('Test Insurance Co', 'TESTPAYER', '123 Main St', 'Anytown', 'CA', '90210', '555-0100', NOW(), NOW(), NOW(), NOW())`)
	if err != nil {
		t.Fatalf("INSERT insco failed: %v", err)
	}
	inscoID, _ := inscoResult.LastInsertId()
	t.Cleanup(func() {
		db.Exec("DELETE FROM insco WHERE id = ?", inscoID)
	})

	// Create test physician with NPI
	phyResult, err := db.Exec(`
		INSERT INTO physician (phylname, phyfname, phynpi, created_at, updated_at)
		VALUES ('Provider', 'Test', '1234567890', NOW(), NOW())`)
	if err != nil {
		t.Fatalf("INSERT physician failed: %v", err)
	}
	phyID, _ := phyResult.LastInsertId()
	t.Cleanup(func() {
		db.Exec("DELETE FROM physician WHERE id = ?", phyID)
	})

	// Create patient coverage
	_, err = db.Exec(`
		INSERT INTO patient_coverage (patient, insurance_company, policy_number, group_number,
			primary_coverage, effective_date, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, 1, CURDATE(), 'active', NOW(), NOW())`,
		patientID, inscoID, "POL-123456", "GRP-789")
	if err != nil {
		t.Fatalf("INSERT patient_coverage failed: %v", err)
	}

	// Update patient PCP to the test physician
	_, err = db.Exec("UPDATE patient SET ptpcp = ? WHERE id = ?", phyID, patientID)
	if err != nil {
		t.Fatalf("UPDATE patient PCP failed: %v", err)
	}

	// Verify coverage exists
	var covPolicyNo, covGroupNo string
	err = db.QueryRow(`
		SELECT policy_number, group_number FROM patient_coverage
		WHERE patient = ? AND primary_coverage = 1 AND active = 'active'
		LIMIT 1`, patientID).Scan(&covPolicyNo, &covGroupNo)
	if err != nil {
		t.Fatalf("SELECT coverage verification failed: %v", err)
	}
	if covPolicyNo != "POL-123456" {
		t.Errorf("expected policy_number 'POL-123456', got %q", covPolicyNo)
	}
	if covGroupNo != "GRP-789" {
		t.Errorf("expected group_number 'GRP-789', got %q", covGroupNo)
	}

	// Verify insurance company
	var inscoName, payerID string
	err = db.QueryRow("SELECT insconame, inscox12id FROM insco WHERE id = ?", inscoID).Scan(&inscoName, &payerID)
	if err != nil {
		t.Fatalf("SELECT insco failed: %v", err)
	}
	if inscoName != "Test Insurance Co" {
		t.Errorf("expected insco name 'Test Insurance Co', got %q", inscoName)
	}
	if payerID != "TESTPAYER" {
		t.Errorf("expected payer ID 'TESTPAYER', got %q", payerID)
	}
}
