//go:build integration

package integration

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/freemed/freemed-server/dbgen"
)

func TestGenerateDailySchedule(t *testing.T) {
	// Connect to MySQL
	db, err := sql.Open("mysql", "freemed:freemed@tcp(127.0.0.1:3306)/freemed?parseTime=true")
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	ctx := context.Background()
	now := time.Now()

	// Step 1: Insert a test provider into the physician table
	result, err := db.ExecContext(ctx, `
		INSERT INTO physician (created_at, updated_at, phylname, phyfname, phynpi)
		VALUES (?, ?, ?, ?, ?)
	`, now, now, "TestDoctor", "John", "1234567890")
	if err != nil {
		t.Fatalf("Failed to insert test provider: %v", err)
	}

	providerID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get provider ID: %v", err)
	}
	t.Logf("Inserted provider with ID: %d", providerID)

	// Cleanup: remove scheduler rows and the test provider on exit
	defer func() {
		db.ExecContext(ctx, "DELETE FROM scheduler WHERE calphysician = ?", providerID)
		db.ExecContext(ctx, "DELETE FROM physician WHERE id = ?", providerID)
	}()

	// Step 2: Call GenerateDailySchedule via dbgen Queries
	queries := dbgen.New(db)

	testDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	rows, err := queries.SchedulerDailyApptScheduler(ctx, dbgen.SchedulerDailyApptSchedulerParams{
		ReqDate:         testDate,
		StartHour:       9,
		EndHour:         16,
		IntervalMinutes: 15,
		ProviderID:      providerID,
	})
	if err != nil {
		t.Fatalf("SchedulerDailyApptScheduler failed: %v", err)
	}
	t.Logf("Stored procedure returned %d rows", len(rows))

	// Step 3: Verify scheduler table has expected rows
	// 7 hours (9-15) * 4 slots/hour (60/15) = 28 slots
	var count int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM scheduler
		WHERE caldateof = ?
		  AND calphysician = ?
		  AND caltype = 'open'
		  AND calstatus = 'open'
	`, testDate, providerID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count scheduler rows: %v", err)
	}

	expected := 28
	if count != expected {
		t.Errorf("Expected %d scheduler rows, got %d", expected, count)
	} else {
		t.Logf("Verified: %d scheduler rows created", count)
	}

	// Verify slot details - first slot should be 9:00
	var calhour, calminute, calduration int
	err = db.QueryRowContext(ctx, `
		SELECT calhour, calminute, calduration
		FROM scheduler
		WHERE caldateof = ?
		  AND calphysician = ?
		  AND caltype = 'open'
		  AND calstatus = 'open'
		ORDER BY calhour, calminute
		LIMIT 1
	`, testDate, providerID).Scan(&calhour, &calminute, &calduration)
	if err != nil {
		t.Fatalf("Failed to query first slot details: %v", err)
	}

	if calhour != 9 || calminute != 0 || calduration != 15 {
		t.Errorf("First slot expected (hour=9, minute=0, duration=15), got (hour=%d, minute=%d, duration=%d)",
			calhour, calminute, calduration)
	} else {
		t.Logf("Verified: first slot at %d:%02d, duration=%d", calhour, calminute, calduration)
	}

	// Verify last slot should be 15:45
	err = db.QueryRowContext(ctx, `
		SELECT calhour, calminute
		FROM scheduler
		WHERE caldateof = ?
		  AND calphysician = ?
		  AND caltype = 'open'
		  AND calstatus = 'open'
		ORDER BY calhour DESC, calminute DESC
		LIMIT 1
	`, testDate, providerID).Scan(&calhour, &calminute)
	if err != nil {
		t.Fatalf("Failed to query last slot details: %v", err)
	}

	if calhour != 15 || calminute != 45 {
		t.Errorf("Last slot expected (hour=15, minute=45), got (hour=%d, minute=%d)",
			calhour, calminute)
	} else {
		t.Logf("Verified: last slot at %d:%02d", calhour, calminute)
	}
}
