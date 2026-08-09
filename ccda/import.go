package ccda

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// ImportResult holds counts of records imported per category.
type ImportResult struct {
	Allergies     int
	Medications   int
	Problems      int
	Vitals        int
	Immunizations int
	Procedures    int
	Results       int
	SocialHistory int
	Encounters    int
	Total         int
}

// ImportCCD imports parsed CCD data into the FreeMED database using direct SQL.
// It performs deduplication by checking for existing matching records.
func ImportCCD(db *sql.DB, parsed *ParsedCCD, patientID int64, userID int64) (*ImportResult, error) {
	if db == nil {
		return nil, fmt.Errorf("ccda: database connection is nil")
	}

	ctx := context.Background()
	result := &ImportResult{}

	// --- Allergies ---
	// NOTE: The allergies table is sparse (only patient + active).
	// Full allergy substance/reaction import requires allergies_atomic table (see gap analysis).
	// For now, mark that we have allergy data but can't import substance details.
	for _, a := range parsed.Allergies {
		if a.Substance == "" {
			continue
		}
		exists := false
		_ = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM allergies WHERE patient = ? AND active = 'active'", patientID).Scan(&exists)
		// Create at least one allergy record to indicate patient has allergies
		if !exists {
			_, err := db.ExecContext(ctx,
				`INSERT INTO allergies (created_at, updated_at, patient, active) VALUES (NOW(), NOW(), ?, 'active')`,
				patientID)
			if err != nil {
				log.Printf("ccda.ImportCCD: allergy insert failed: %v", err)
			} else {
				result.Allergies++
			}
		}
	}
	// NOTE: allergy substance/reaction details not importable until allergies_atomic table exists

	// --- Medications ---
	for _, m := range parsed.Medications {
		if m.DrugName == "" {
			continue
		}
		var count int
		err := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM medications WHERE patient = ? AND drug_name = ? AND active = 'active'",
			patientID, m.DrugName).Scan(&count)
		if err != nil || count > 0 {
			continue
		}

		startDate := parseCCDADateNull(m.StartDate)
		endDate := parseCCDADateNull(m.EndDate)

		_, err = db.ExecContext(ctx, `
			INSERT INTO medications (patient, drug_name, dosage, frequency, start_date, end_date,
				prescribing_provider, active, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, 'active', NOW(), NOW())
		`, patientID, m.DrugName, m.Dosage, m.Frequency, startDate, endDate, userID)
		if err != nil {
			log.Printf("ccda.ImportCCD: medication insert failed for %s: %v", m.DrugName, err)
		} else {
			result.Medications++
		}
	}

	// --- Problems ---
	for _, p := range parsed.Problems {
		if p.Condition == "" {
			continue
		}
		var count int
		err := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM current_problems WHERE patient = ? AND problem = ? AND active = 'active'",
			patientID, p.Condition).Scan(&count)
		if err != nil || count > 0 {
			continue
		}

		probDate := time.Now()
		if d := parseCCDADateNull(p.OnsetDate); d.Valid {
			probDate = d.Time
		}

		_, err = db.ExecContext(ctx, `
			INSERT INTO current_problems (patient, date, problem, user, active, created_at, updated_at)
			VALUES (?, ?, ?, ?, 'active', NOW(), NOW())
		`, patientID, probDate, p.Condition, userID)
		if err != nil {
			log.Printf("ccda.ImportCCD: problem insert failed for %s: %v", p.Condition, err)
		} else {
			result.Problems++
		}
	}

	// --- Vitals ---
	for _, v := range parsed.Vitals {
		tempStr := sql.NullString{}
		if v.TempF != nil {
			tempStr = sql.NullString{String: fmt.Sprintf("%.1f", *v.TempF), Valid: true}
		}
		heightStr := sql.NullString{}
		if v.HeightCm != nil {
			heightStr = sql.NullString{String: fmt.Sprintf("%.1f", *v.HeightCm), Valid: true}
		}
		weightStr := sql.NullString{}
		if v.WeightKg != nil {
			weightStr = sql.NullString{String: fmt.Sprintf("%.1f", *v.WeightKg), Valid: true}
		}
		bmiStr := sql.NullString{}
		if v.BMI != nil {
			bmiStr = sql.NullString{String: fmt.Sprintf("%.1f", *v.BMI), Valid: true}
		}

		_, err := db.ExecContext(ctx, `
			INSERT INTO vitals (patient, date_taken, systolic, diastolic, heart_rate,
				respiratory_rate, temperature, oxygen_saturation, height_cm, weight_kg, bmi,
				user, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		`, patientID, timeOrNow(v.Date),
			nullFloat64ToNullInt32(v.Systolic), nullFloat64ToNullInt32(v.Diastolic),
			nullFloat64ToNullInt32(v.HeartRate), nullFloat64ToNullInt32(v.RespRate),
			tempStr, nullFloat64ToNullInt32(v.O2Sat),
			heightStr, weightStr, bmiStr, userID)
		if err != nil {
			log.Printf("ccda.ImportCCD: vitals insert failed: %v", err)
		} else {
			result.Vitals++
		}
	}

	// --- Immunizations ---
	// NOTE: The immunization table uses reference IDs (immunization, route, body_site are
	// int64 foreign keys to immunization reference tables), not string vaccine names.
	// For C-CDA import, we can only record that immunizations exist, not specific vaccine names.
	for _, i := range parsed.Immunizations {
		if i.Vaccine == "" {
			continue
		}
		immDate := time.Now()
		if d := parseCCDADateNull(i.DateGiven); d.Valid {
			immDate = d.Time
		}

		_, err := db.ExecContext(ctx, `
			INSERT INTO immunization (dateof, patient, provider, admin_provider, immunization, 
				route, body_site, manufacturer, lot_number, previous_doses, recovered,
				user, active, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'active', NOW(), NOW())
		`, immDate, patientID, userID, userID, 0, 0, 0,
			sql.NullString{String: i.Manufacturer, Valid: i.Manufacturer != ""},
			sql.NullString{String: i.LotNumber, Valid: i.LotNumber != ""},
			0, false, userID)
		if err != nil {
			log.Printf("ccda.ImportCCD: immunization insert failed for %s: %v", i.Vaccine, err)
		} else {
			result.Immunizations++
		}
	}

	// --- Procedures (previous_operations) ---
	for _, p := range parsed.Procedures {
		if p.Procedure == "" {
			continue
		}
		opDate := sql.NullTime{}
		if d := parseCCDADateNull(p.Date); d.Valid {
			opDate = d
		}

		_, err := db.ExecContext(ctx, `
			INSERT INTO previous_operations (patient, operation_date, operation, user, created_at, updated_at)
			VALUES (?, ?, ?, ?, NOW(), NOW())
		`, patientID, opDate, p.Procedure, userID)
		if err != nil {
			log.Printf("ccda.ImportCCD: procedure insert failed for %s: %v", p.Procedure, err)
		} else {
			result.Procedures++
		}
	}

	// --- Social History ---
	for _, s := range parsed.SocialHistory {
		recDate := time.Now()
		if d := parseCCDADateNull(s.EffectiveDate); d.Valid {
			recDate = d.Time
		}

		_, err := db.ExecContext(ctx, `
			INSERT INTO social_history (patient, smoking_status, smoking_detail, alcohol_use, 
				alcohol_detail, drug_use, drug_detail, recorded_date, user, active, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'active', NOW(), NOW())
		`, patientID,
			socialMap(s, "smoking", "value"), socialMap(s, "smoking", "detail"),
			socialMap(s, "alcohol", "value"), socialMap(s, "alcohol", "detail"),
			socialMap(s, "drug_use", "value"), socialMap(s, "drug_use", "detail"),
			recDate, userID)
		if err != nil {
			log.Printf("ccda.ImportCCD: social_history insert failed: %v", err)
		} else {
			result.SocialHistory++
		}
	}

	result.Total = result.Allergies + result.Medications + result.Problems +
		result.Vitals + result.Immunizations + result.Procedures +
		result.Results + result.SocialHistory + result.Encounters

	return result, nil
}

// =============================================================================
// Helper functions
// =============================================================================

// parseCCDADateNull parses a CCDA date string (YYYYMMDD or YYYYMMDDHHMMSS) into sql.NullTime.
func parseCCDADateNull(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{}
	}
	// Try YYYYMMDD
	if len(s) >= 8 {
		t, err := time.Parse("20060102", s[:8])
		if err == nil {
			return sql.NullTime{Time: t, Valid: true}
		}
	}
	// Try YYYYMMDDHHMMSS
	if len(s) >= 14 {
		t, err := time.Parse("20060102150405", s[:14])
		if err == nil {
			return sql.NullTime{Time: t, Valid: true}
		}
	}
	return sql.NullTime{}
}

func timeOrNow(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now()
	}
	return t
}

func nullFloat64ToNullInt32(v *float64) sql.NullInt32 {
	if v == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: int32(*v), Valid: true}
}

// socialMap maps a parsed social history entry to the correct column based on category.
func socialMap(s ParsedSocialHistory, category, field string) string {
	if s.Category != category {
		return ""
	}
	if field == "value" {
		return s.Value
	}
	return s.Detail
}
