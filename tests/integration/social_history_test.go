//go:build integration

package integration

import (
	"database/sql"
	"fmt"
	"testing"
	"time"
)

// cleanSocialHistory removes inserted social history rows.
func cleanSocialHistory(t *testing.T, db *sql.DB, ids ...int64) {
	t.Helper()
	for _, id := range ids {
		if id == 0 {
			continue
		}
		_, _ = db.Exec("DELETE FROM social_history WHERE id = ?", id)
	}
}

// TestSocialHistoryCRUD tests full CRUD for social_history records.
func TestSocialHistoryCRUD(t *testing.T) {
	db := openDB(t)
	defer db.Close()

	now := time.Now()
	ptid := fmt.Sprintf("SH-%d", now.UnixNano())

	result, err := db.Exec(`
		INSERT INTO patient (created_at, updated_at, ptdtadd, stamp,
			ptlname, ptfname, ptid, ptsex, ptdob, user, provider, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now, now, now, now, "Social", "Test", ptid, "F", "1995-03-01", 1, 1, "active")
	if err != nil {
		t.Fatalf("INSERT patient failed: %v", err)
	}
	patientID, _ := result.LastInsertId()
	t.Cleanup(func() { cleanPatient(t, db, patientID) })

	// CREATE social history record
	recDate := time.Now().Truncate(24 * time.Hour)
	shResult, err := db.Exec(`
		INSERT INTO social_history (patient, smoking_status, smoking_detail,
			alcohol_use, alcohol_detail, drug_use, drug_detail,
			exercise_frequency, occupation, living_situation,
			food_insecurity, transportation_access, notes, recorded_date,
			user, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'active', NOW(), NOW())`,
		patientID, "current-smoker", "1 pack/day", "moderate", "2-3 drinks/week",
		"none", "", "regular", "Software Engineer", "with-family",
		false, true, "No concerns", recDate, 1)
	if err != nil {
		t.Fatalf("INSERT social_history failed: %v", err)
	}
	shID, _ := shResult.LastInsertId()
	t.Cleanup(func() { cleanSocialHistory(t, db, shID) })

	if shID == 0 {
		t.Fatal("expected non-zero social_history id")
	}

	// READ
	var (
		gotSmoking, gotAlcohol, gotDrugs, gotExercise string
		gotOccupation, gotLiving                      string
		gotFood, gotTransport                         bool
	)
	err = db.QueryRow(`
		SELECT smoking_status, alcohol_use, drug_use, exercise_frequency,
			occupation, living_situation, food_insecurity, transportation_access
		FROM social_history WHERE id = ? AND active = 'active'`,
		shID).Scan(&gotSmoking, &gotAlcohol, &gotDrugs, &gotExercise,
		&gotOccupation, &gotLiving, &gotFood, &gotTransport)
	if err != nil {
		t.Fatalf("SELECT social_history failed: %v", err)
	}

	if gotSmoking != "current-smoker" {
		t.Errorf("expected smoking 'current-smoker', got %q", gotSmoking)
	}
	if gotAlcohol != "moderate" {
		t.Errorf("expected alcohol 'moderate', got %q", gotAlcohol)
	}
	if gotOccupation != "Software Engineer" {
		t.Errorf("expected occupation 'Software Engineer', got %q", gotOccupation)
	}

	// Get latest
	var latestID int64
	err = db.QueryRow(`
		SELECT id FROM social_history
		WHERE patient = ? AND active = 'active'
		ORDER BY recorded_date DESC LIMIT 1`, patientID).Scan(&latestID)
	if err != nil {
		t.Fatalf("GetLatestSocialHistory failed: %v", err)
	}
	if latestID != shID {
		t.Errorf("expected latest social_history ID %d, got %d", shID, latestID)
	}

	// Soft delete
	_, err = db.Exec(`
		UPDATE social_history SET active = 'inactive', deleted_at = NOW(), updated_at = NOW()
		WHERE id = ? AND patient = ?`, shID, patientID)
	if err != nil {
		t.Fatalf("soft DELETE social_history failed: %v", err)
	}

	var activeCount int
	err = db.QueryRow("SELECT COUNT(*) FROM social_history WHERE id = ? AND active = 'active'", shID).Scan(&activeCount)
	if err != nil {
		t.Fatalf("COUNT after delete failed: %v", err)
	}
	if activeCount != 0 {
		t.Errorf("expected 0 active rows after soft delete, got %d", activeCount)
	}
}

// TestSocialHistorySmokingStatuses tests different smoking status values.
func TestSocialHistorySmokingStatuses(t *testing.T) {
	db := openDB(t)
	defer db.Close()

	now := time.Now()
	ptid := fmt.Sprintf("SHS-%d", now.UnixNano())

	result, err := db.Exec(`
		INSERT INTO patient (created_at, updated_at, ptdtadd, stamp,
			ptlname, ptfname, ptid, ptsex, ptdob, user, provider, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now, now, now, now, "Smoke", "Test", ptid, "M", "1980-01-01", 1, 1, "active")
	if err != nil {
		t.Fatalf("INSERT patient failed: %v", err)
	}
	patientID, _ := result.LastInsertId()
	t.Cleanup(func() { cleanPatient(t, db, patientID) })

	var ids []int64
	for _, status := range []string{"never-smoker", "former-smoker", "current-smoker", "unknown"} {
		r, err := db.Exec(`
			INSERT INTO social_history (patient, smoking_status, recorded_date, user, active, created_at, updated_at)
			VALUES (?, ?, NOW(), ?, 'active', NOW(), NOW())`,
			patientID, status, 1)
		if err != nil {
			t.Fatalf("INSERT social_history (%s) failed: %v", status, err)
		}
		id, _ := r.LastInsertId()
		ids = append(ids, id)
	}
	t.Cleanup(func() {
		for _, id := range ids {
			cleanSocialHistory(t, db, id)
		}
	})

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM social_history WHERE patient = ? AND active = 'active'", patientID).Scan(&count)
	if err != nil {
		t.Fatalf("COUNT failed: %v", err)
	}
	if count != 4 {
		t.Errorf("expected 4 social history records, got %d", count)
	}
}
