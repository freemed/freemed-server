//go:build integration

package integration

import (
	"fmt"
	"testing"
	"time"
)

// TestPatientRollback tests transaction behaviour for patient INSERT:
//  1. BEGIN → INSERT → ROLLBACK — row must NOT exist.
//  2. BEGIN → INSERT → COMMIT   — row must persist.
func TestPatientRollback(t *testing.T) {
	t.Run("RollbackDoesNotPersist", func(t *testing.T) {
		db := openDB(t)
		defer db.Close()

		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("BEGIN failed: %v", err)
		}

		now := time.Now()
		ptid := fmt.Sprintf("RB-%d", now.UnixNano())

		result, err := tx.Exec(`
			INSERT INTO patient (
				created_at, updated_at, ptdtadd, stamp,
				ptlname, ptfname, ptmname, ptsex, ptid,
				ptdob, ssn, pemail, ptprimarylanguage,
				user, provider, status
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			now, now, now, now,
			"Rollback", "Test", "X", "M", ptid,
			"1990-06-15", "987-65-4321", "rollback@example.com", "en",
			1, 2, "active",
		)
		if err != nil {
			tx.Rollback()
			t.Fatalf("INSERT inside transaction failed: %v", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			tx.Rollback()
			t.Fatalf("LastInsertId inside transaction failed: %v", err)
		}
		t.Logf("Inserted id=%d inside transaction (about to rollback)", id)

		// Intentionally rollback
		if err := tx.Rollback(); err != nil {
			t.Fatalf("ROLLBACK failed: %v", err)
		}

		// Verify the row does NOT exist
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM patient WHERE id = ?", id).Scan(&count)
		if err != nil {
			t.Fatalf("SELECT COUNT after rollback failed: %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0 rows after rollback, got %d (id=%d)", count, id)
		}

		// Also verify by ptid
		err = db.QueryRow("SELECT COUNT(*) FROM patient WHERE ptid = ?", ptid).Scan(&count)
		if err != nil {
			t.Fatalf("SELECT COUNT by ptid after rollback failed: %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0 rows with ptid=%q after rollback, got %d", ptid, count)
		}

		t.Logf("ROLLBACK verified: no rows for ptid=%s", ptid)
	})

	t.Run("CommitPersists", func(t *testing.T) {
		db := openDB(t)
		defer db.Close()

		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("BEGIN failed: %v", err)
		}

		now := time.Now()
		ptid := fmt.Sprintf("CMT-%d", now.UnixNano())

		result, err := tx.Exec(`
			INSERT INTO patient (
				created_at, updated_at, ptdtadd, stamp,
				ptlname, ptfname, ptmname, ptsex, ptid,
				ptdob, ssn, pemail, ptprimarylanguage,
				user, provider, status
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			now, now, now, now,
			"Commit", "Test", "Y", "F", ptid,
			"1985-03-22", "456-78-9012", "commit@example.com", "en",
			1, 2, "active",
		)
		if err != nil {
			tx.Rollback()
			t.Fatalf("INSERT inside transaction failed: %v", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			tx.Rollback()
			t.Fatalf("LastInsertId inside transaction failed: %v", err)
		}
		t.Logf("Inserted id=%d inside transaction (about to commit)", id)

		if err := tx.Commit(); err != nil {
			t.Fatalf("COMMIT failed: %v", err)
		}

		// Cleanup after test succeeds
		t.Cleanup(func() { cleanPatient(t, db, id) })

		// Verify the row exists
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM patient WHERE id = ?", id).Scan(&count)
		if err != nil {
			t.Fatalf("SELECT COUNT after commit failed: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1 row after commit, got %d (id=%d)", count, id)
		}

		// Verify field values
		var gotFname, gotLname, gotPtid string
		err = db.QueryRow("SELECT ptfname, ptlname, ptid FROM patient WHERE id = ?", id).
			Scan(&gotFname, &gotLname, &gotPtid)
		if err != nil {
			t.Fatalf("SELECT after commit failed: %v", err)
		}
		if gotFname != "Test" {
			t.Errorf("ptfname = %q, want %q", gotFname, "Test")
		}
		if gotLname != "Commit" {
			t.Errorf("ptlname = %q, want %q", gotLname, "Commit")
		}
		if gotPtid != ptid {
			t.Errorf("ptid = %q, want %q", gotPtid, ptid)
		}

		t.Logf("COMMIT verified: row id=%d ptid=%s persisted", id, ptid)
	})
}
