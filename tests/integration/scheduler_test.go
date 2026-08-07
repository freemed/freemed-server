//go:build integration

package integration

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const schedulerDSN = "freemed:freemed@tcp(127.0.0.1:3306)/freemed?parseTime=true"

// openSchedulerDB opens a MySQL connection for scheduler integration tests.
func openSchedulerDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("mysql", schedulerDSN)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}
	return db
}

// insertTestProvider inserts a test physician and returns the new ID.
func insertTestProvider(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	now := time.Now()
	result, err := db.Exec(
		"INSERT INTO physician (created_at, updated_at, phylname, phyfname) VALUES (?, ?, ?, ?)",
		now, now, "TestDoctor", "Integration",
	)
	if err != nil {
		t.Fatalf("failed to insert test provider: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get provider LastInsertId: %v", err)
	}
	t.Logf("Inserted test provider with ID: %d", id)
	return id
}

// TestSchedulerCreate inserts a test provider, creates an appointment via direct
// INSERT, and verifies the row exists.
func TestSchedulerCreate(t *testing.T) {
	db := openSchedulerDB(t)
	defer db.Close()

	providerID := insertTestProvider(t, db)

	// Cleanup on exit
	t.Cleanup(func() {
		db.Exec("DELETE FROM scheduler WHERE calphysician = ?", providerID)
		db.Exec("DELETE FROM physician WHERE id = ?", providerID)
	})

	now := time.Now()
	apptDate := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)

	result, err := db.Exec(`
		INSERT INTO scheduler (
			created_at, updated_at,
			caldateof, calcreated,
			caltype, calhour, calminute, calduration,
			calphysician, calpatient,
			calstatus, calprenote,
			user
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now, now,
		apptDate, now,
		"appointment", 10, 0, 30,
		providerID, 1,
		"scheduled", "Initial consultation",
		1,
	)
	if err != nil {
		t.Fatalf("INSERT scheduler failed: %v", err)
	}

	apptID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("LastInsertId failed: %v", err)
	}
	t.Logf("Created appointment with ID: %d", apptID)

	if apptID == 0 {
		t.Fatal("expected non-zero appointment ID")
	}

	// Verify the row exists
	var (
		gotPhysician int64
		gotStatus    string
		gotType      string
		gotHour      int
		gotDuration  int
	)
	err = db.QueryRow(
		"SELECT calphysician, calstatus, caltype, calhour, calduration FROM scheduler WHERE id = ?",
		apptID,
	).Scan(&gotPhysician, &gotStatus, &gotType, &gotHour, &gotDuration)
	if err != nil {
		t.Fatalf("SELECT failed: %v", err)
	}

	if gotPhysician != providerID {
		t.Errorf("calphysician = %d, want %d", gotPhysician, providerID)
	}
	if gotStatus != "scheduled" {
		t.Errorf("calstatus = %q, want %q", gotStatus, "scheduled")
	}
	if gotType != "appointment" {
		t.Errorf("caltype = %q, want %q", gotType, "appointment")
	}
	if gotHour != 10 {
		t.Errorf("calhour = %d, want 10", gotHour)
	}
	if gotDuration != 30 {
		t.Errorf("calduration = %d, want 30", gotDuration)
	}

	t.Logf("Verified appointment: id=%d, provider=%d, status=%s, type=%s, hour=%d, duration=%d",
		apptID, gotPhysician, gotStatus, gotType, gotHour, gotDuration)
}

// TestSchedulerCancel creates an appointment, updates calstatus to 'cancelled',
// and verifies the status changed.
func TestSchedulerCancel(t *testing.T) {
	db := openSchedulerDB(t)
	defer db.Close()

	providerID := insertTestProvider(t, db)

	t.Cleanup(func() {
		db.Exec("DELETE FROM scheduler WHERE calphysician = ?", providerID)
		db.Exec("DELETE FROM physician WHERE id = ?", providerID)
	})

	now := time.Now()
	apptDate := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)

	// Create the appointment
	result, err := db.Exec(`
		INSERT INTO scheduler (
			created_at, updated_at, caldateof, calcreated,
			caltype, calhour, calminute, calduration,
			calphysician, calpatient, calstatus, user
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now, now, apptDate, now,
		"appointment", 14, 0, 30,
		providerID, 1, "scheduled", 1,
	)
	if err != nil {
		t.Fatalf("INSERT scheduler failed: %v", err)
	}

	apptID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("LastInsertId failed: %v", err)
	}
	t.Logf("Created appointment with ID: %d, status=scheduled", apptID)

	// Cancel the appointment
	_, err = db.Exec(
		"UPDATE scheduler SET calstatus = ?, updated_at = ? WHERE id = ?",
		"cancelled", time.Now(), apptID,
	)
	if err != nil {
		t.Fatalf("UPDATE calstatus failed: %v", err)
	}

	// Verify status changed
	var gotStatus string
	err = db.QueryRow("SELECT calstatus FROM scheduler WHERE id = ?", apptID).Scan(&gotStatus)
	if err != nil {
		t.Fatalf("SELECT failed: %v", err)
	}

	if gotStatus != "cancelled" {
		t.Errorf("calstatus = %q, want %q", gotStatus, "cancelled")
	}

	t.Logf("Verified appointment id=%d cancelled: status=%s", apptID, gotStatus)
}

// TestSchedulerReschedule creates an appointment, updates caldateof to a new
// date, and verifies the change.
func TestSchedulerReschedule(t *testing.T) {
	db := openSchedulerDB(t)
	defer db.Close()

	providerID := insertTestProvider(t, db)

	t.Cleanup(func() {
		db.Exec("DELETE FROM scheduler WHERE calphysician = ?", providerID)
		db.Exec("DELETE FROM physician WHERE id = ?", providerID)
	})

	now := time.Now()
	origDate := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)
	newDate := time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC)

	// Create the appointment
	result, err := db.Exec(`
		INSERT INTO scheduler (
			created_at, updated_at, caldateof, calcreated,
			caltype, calhour, calminute, calduration,
			calphysician, calpatient, calstatus, user
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now, now, origDate, now,
		"appointment", 9, 30, 20,
		providerID, 1, "scheduled", 1,
	)
	if err != nil {
		t.Fatalf("INSERT scheduler failed: %v", err)
	}

	apptID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("LastInsertId failed: %v", err)
	}
	t.Logf("Created appointment with ID: %d, date=%s", apptID, origDate.Format("2006-01-02"))

	// Reschedule to new date
	_, err = db.Exec(
		"UPDATE scheduler SET caldateof = ?, updated_at = ? WHERE id = ?",
		newDate, time.Now(), apptID,
	)
	if err != nil {
		t.Fatalf("UPDATE caldateof failed: %v", err)
	}

	// Verify date changed
	var gotDate time.Time
	err = db.QueryRow("SELECT caldateof FROM scheduler WHERE id = ?", apptID).Scan(&gotDate)
	if err != nil {
		t.Fatalf("SELECT failed: %v", err)
	}

	if gotDate.Format("2006-01-02") != newDate.Format("2006-01-02") {
		t.Errorf("caldateof = %s, want %s", gotDate.Format("2006-01-02"), newDate.Format("2006-01-02"))
	}

	t.Logf("Verified appointment id=%d rescheduled: date=%s", apptID, gotDate.Format("2006-01-02"))
}

// TestSchedulerDoubleBook inserts two appointments for the same provider at the
// same time and verifies both exist (no unique constraint on provider+time).
func TestSchedulerDoubleBook(t *testing.T) {
	db := openSchedulerDB(t)
	defer db.Close()

	providerID := insertTestProvider(t, db)

	t.Cleanup(func() {
		db.Exec("DELETE FROM scheduler WHERE calphysician = ?", providerID)
		db.Exec("DELETE FROM physician WHERE id = ?", providerID)
	})

	now := time.Now()
	apptDate := time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC)

	// Insert first appointment
	result1, err := db.Exec(`
		INSERT INTO scheduler (
			created_at, updated_at, caldateof, calcreated,
			caltype, calhour, calminute, calduration,
			calphysician, calpatient, calstatus, user
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now, now, apptDate, now,
		"appointment", 11, 0, 30,
		providerID, 1, "scheduled", 1,
	)
	if err != nil {
		t.Fatalf("first INSERT scheduler failed: %v", err)
	}

	id1, err := result1.LastInsertId()
	if err != nil {
		t.Fatalf("first LastInsertId failed: %v", err)
	}
	t.Logf("Created first appointment with ID: %d", id1)

	// Insert second appointment — same provider, same date, same time
	result2, err := db.Exec(`
		INSERT INTO scheduler (
			created_at, updated_at, caldateof, calcreated,
			caltype, calhour, calminute, calduration,
			calphysician, calpatient, calstatus, user
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now, now, apptDate, now,
		"appointment", 11, 0, 30,
		providerID, 2, "scheduled", 1,
	)
	if err != nil {
		t.Fatalf("second INSERT scheduler failed: %v", err)
	}

	id2, err := result2.LastInsertId()
	if err != nil {
		t.Fatalf("second LastInsertId failed: %v", err)
	}
	t.Logf("Created second appointment with ID: %d", id2)

	// Verify both rows exist
	var count int
	err = db.QueryRow(
		"SELECT COUNT(*) FROM scheduler WHERE calphysician = ? AND caldateof = ? AND calhour = 11 AND calminute = 0",
		providerID, apptDate,
	).Scan(&count)
	if err != nil {
		t.Fatalf("SELECT COUNT failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 appointments for provider=%d at date=%s hour=11, got %d",
			providerID, apptDate.Format("2006-01-02"), count)
	}

	t.Logf("Verified double-booking: %d appointments exist for provider=%d at date=%s hour=11",
		count, providerID, apptDate.Format("2006-01-02"))
}
