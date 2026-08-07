//go:build integration

package integration

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const dsn = "freemed:freemed@tcp(127.0.0.1:3306)/freemed?parseTime=true&loc=Local"

func openDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}
	return db
}

// cleanPatient removes rows inserted by a test to keep the DB clean.
func cleanPatient(t *testing.T, db *sql.DB, ids ...int64) {
	t.Helper()
	for _, id := range ids {
		if id == 0 {
			continue
		}
		_, _ = db.Exec("DELETE FROM patient WHERE id = ?", id)
	}
}

// TestCreatePatientSuccess inserts a patient with all required fields and
// verifies the row can be read back.
func TestCreatePatientSuccess(t *testing.T) {
	db := openDB(t)
	defer db.Close()

	now := time.Now()
	ptid := fmt.Sprintf("TEST-%d", now.UnixNano())

	result, err := db.Exec(`
		INSERT INTO patient (
			created_at, updated_at, ptdtadd, stamp,
			ptlname, ptfname, ptmname, ptsex, ptid,
			ptdob, ssn, pemail, ptprimarylanguage,
			user, provider, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now, now, now, now,
		"Smith", "John", "A", "M", ptid,
		"1980-01-15", "123-45-6789", "john@example.com", "en",
		1, 2, "active",
	)
	if err != nil {
		t.Fatalf("INSERT failed: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("LastInsertId failed: %v", err)
	}
	t.Cleanup(func() { cleanPatient(t, db, id) })

	if id == 0 {
		t.Fatal("expected non-zero id")
	}

	// Verify the row exists
	var gotFname, gotLname, gotPtid string
	err = db.QueryRow("SELECT ptfname, ptlname, ptid FROM patient WHERE id = ?", id).
		Scan(&gotFname, &gotLname, &gotPtid)
	if err != nil {
		t.Fatalf("SELECT failed: %v", err)
	}
	if gotFname != "John" {
		t.Errorf("ptfname = %q, want %q", gotFname, "John")
	}
	if gotLname != "Smith" {
		t.Errorf("ptlname = %q, want %q", gotLname, "Smith")
	}
	if gotPtid != ptid {
		t.Errorf("ptid = %q, want %q", gotPtid, ptid)
	}

	t.Logf("Created patient id=%d ptid=%s", id, ptid)
}

// TestCreatePatientDuplicate inserts the same patient data twice and verifies
// both rows exist (no unique constraint violation on patient identity fields).
func TestCreatePatientDuplicate(t *testing.T) {
	db := openDB(t)
	defer db.Close()

	now := time.Now()
	sharedPtid := fmt.Sprintf("DUP-%d", now.UnixNano())
	var ids []int64

	for i := 0; i < 2; i++ {
		result, err := db.Exec(`
			INSERT INTO patient (
				created_at, updated_at, ptdtadd, stamp,
				ptlname, ptfname, ptsex, ptid,
				user, provider, status
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			now, now, now, now,
			"DupTest", "Patient", "F", sharedPtid,
			1, 2, "active",
		)
		if err != nil {
			t.Fatalf("INSERT %d failed: %v", i+1, err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("LastInsertId %d failed: %v", i+1, err)
		}
		ids = append(ids, id)
	}

	t.Cleanup(func() { cleanPatient(t, db, ids...) })

	// Both rows should exist
	for _, id := range ids {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM patient WHERE id = ?", id).Scan(&count)
		if err != nil {
			t.Errorf("SELECT count for id=%d failed: %v", id, err)
		}
		if count != 1 {
			t.Errorf("expected 1 row for id=%d, got %d", id, count)
		}
	}

	// Both rows with the same ptid should exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM patient WHERE ptid = ?", sharedPtid).Scan(&count)
	if err != nil {
		t.Fatalf("SELECT count by ptid failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 rows with ptid=%q, got %d", sharedPtid, count)
	}

	t.Logf("Created %d duplicate patient rows with ptid=%s", count, sharedPtid)
}

// TestCreatePatientMissingFields attempts to insert a patient without required
// datetime columns (created_at, updated_at, ptdtadd, stamp) and expects an error
// because STRICT_TRANS_TABLES rejects missing NOT NULL columns without defaults.
func TestCreatePatientMissingFields(t *testing.T) {
	db := openDB(t)
	defer db.Close()

	// Omit created_at, updated_at, ptdtadd, and stamp — all are NOT NULL
	// without DEFAULT values. In strict mode this must fail.
	_, err := db.Exec(`
		INSERT INTO patient (
			ptlname, ptfname, ptsex, ptid,
			user, provider, status
		) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"Error", "Test", "M", "MISSING-FIELDS-TEST",
		1, 2, "active",
	)
	if err == nil {
		t.Fatal("expected INSERT error for missing NOT NULL fields, but got none")
	}

	t.Logf("Correctly rejected INSERT with missing fields: %v", err)
}
