//go:build integration

package integration

import (
	"database/sql"
	"fmt"
	"testing"
	"time"
)

// cleanFamilyHistory removes inserted family history rows.
func cleanFamilyHistory(t *testing.T, db *sql.DB, ids ...int64) {
	t.Helper()
	for _, id := range ids {
		if id == 0 {
			continue
		}
		_, _ = db.Exec("DELETE FROM family_history WHERE id = ?", id)
	}
}

// TestFamilyHistoryCRUD tests full CRUD for family_history records.
func TestFamilyHistoryCRUD(t *testing.T) {
	db := openDB(t)
	defer db.Close()

	// First, create a test patient
	now := time.Now()
	ptid := fmt.Sprintf("FH-%d", now.UnixNano())

	result, err := db.Exec(`
		INSERT INTO patient (created_at, updated_at, ptdtadd, stamp,
			ptlname, ptfname, ptid, ptsex, ptdob, user, provider, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now, now, now, now, "Doe", "Jane", ptid, "F", "1990-06-15", 1, 1, "active")
	if err != nil {
		t.Fatalf("INSERT patient failed: %v", err)
	}
	patientID, _ := result.LastInsertId()
	t.Cleanup(func() { cleanPatient(t, db, patientID) })

	// CREATE family history record
	fhResult, err := db.Exec(`
		INSERT INTO family_history (patient, relationship, condition_name, icd10_code,
			onset_age, deceased, notes, user, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'active', NOW(), NOW())`,
		patientID, "Mother", "Type 2 Diabetes", "E11.9", 45, 0, "Diagnosed at age 45", 1)
	if err != nil {
		t.Fatalf("INSERT family_history failed: %v", err)
	}
	fhID, _ := fhResult.LastInsertId()
	t.Cleanup(func() { cleanFamilyHistory(t, db, fhID) })

	if fhID == 0 {
		t.Fatal("expected non-zero family_history id")
	}

	// READ: verify the row exists
	var (
		gotRel, gotCond, gotICD10 string
		gotOnsetAge               int32
		gotDeceased               bool
	)
	err = db.QueryRow("SELECT relationship, condition_name, icd10_code, onset_age, deceased FROM family_history WHERE id = ? AND active = 'active'",
		fhID).Scan(&gotRel, &gotCond, &gotICD10, &gotOnsetAge, &gotDeceased)
	if err != nil {
		t.Fatalf("SELECT family_history failed: %v", err)
	}
	if gotRel != "Mother" {
		t.Errorf("expected relationship 'Mother', got %q", gotRel)
	}
	if gotCond != "Type 2 Diabetes" {
		t.Errorf("expected condition 'Type 2 Diabetes', got %q", gotCond)
	}
	if gotICD10 != "E11.9" {
		t.Errorf("expected ICD10 'E11.9', got %q", gotICD10)
	}
	if gotOnsetAge != 45 {
		t.Errorf("expected onset_age 45, got %d", gotOnsetAge)
	}
	if gotDeceased {
		t.Error("expected deceased to be false")
	}

	// UPDATE
	_, err = db.Exec(`
		UPDATE family_history SET condition_name = ?, onset_age = ?, updated_at = NOW()
		WHERE id = ? AND patient = ? AND active = 'active'`,
		"Type 2 Diabetes with complications", 50, fhID, patientID)
	if err != nil {
		t.Fatalf("UPDATE family_history failed: %v", err)
	}

	// Verify update
	var updatedCond string
	var updatedAge int32
	err = db.QueryRow("SELECT condition_name, onset_age FROM family_history WHERE id = ? AND active = 'active'",
		fhID).Scan(&updatedCond, &updatedAge)
	if err != nil {
		t.Fatalf("SELECT after update failed: %v", err)
	}
	if updatedCond != "Type 2 Diabetes with complications" {
		t.Errorf("expected updated condition, got %q", updatedCond)
	}
	if updatedAge != 50 {
		t.Errorf("expected updated onset_age 50, got %d", updatedAge)
	}

	// DELETE (soft delete)
	_, err = db.Exec(`
		UPDATE family_history SET active = 'inactive', deleted_at = NOW(), updated_at = NOW()
		WHERE id = ? AND patient = ?`, fhID, patientID)
	if err != nil {
		t.Fatalf("soft DELETE family_history failed: %v", err)
	}

	// Verify soft delete — should not return
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM family_history WHERE id = ? AND active = 'active'", fhID).Scan(&count)
	if err != nil {
		t.Fatalf("COUNT after delete failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 active rows after soft delete, got %d", count)
	}
}

// TestFamilyHistoryListByPatient tests listing family history for a patient.
func TestFamilyHistoryListByPatient(t *testing.T) {
	db := openDB(t)
	defer db.Close()

	now := time.Now()
	ptid := fmt.Sprintf("FHL-%d", now.UnixNano())

	result, err := db.Exec(`
		INSERT INTO patient (created_at, updated_at, ptdtadd, stamp,
			ptlname, ptfname, ptid, ptsex, ptdob, user, provider, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now, now, now, now, "List", "Test", ptid, "M", "1985-01-01", 1, 1, "active")
	if err != nil {
		t.Fatalf("INSERT patient failed: %v", err)
	}
	patientID, _ := result.LastInsertId()
	t.Cleanup(func() { cleanPatient(t, db, patientID) })

	// Insert 2 family history records
	var ids []int64
	for _, rel := range []string{"Father", "Brother"} {
		r, err := db.Exec(`
			INSERT INTO family_history (patient, relationship, condition_name, icd10_code,
				onset_age, deceased, notes, user, active, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'active', NOW(), NOW())`,
			patientID, rel, "Hypertension", "I10", 55, 0, "", 1)
		if err != nil {
			t.Fatalf("INSERT family_history (%s) failed: %v", rel, err)
		}
		id, _ := r.LastInsertId()
		ids = append(ids, id)
	}
	t.Cleanup(func() {
		for _, id := range ids {
			cleanFamilyHistory(t, db, id)
		}
	})

	// List by patient
	rows, err := db.Query("SELECT id, relationship, condition_name FROM family_history WHERE patient = ? AND active = 'active' ORDER BY created_at DESC", patientID)
	if err != nil {
		t.Fatalf("SELECT family_history list failed: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	if count != 2 {
		t.Errorf("expected 2 family history records, got %d", count)
	}
}
