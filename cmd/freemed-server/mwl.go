package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/freemed/freemed-server/config"
	"github.com/freemed/freemed-server/model"
	"github.com/freemed/freemed-server/pkg/dicomnet"
)

// mwlSource is the database backed dicomnet.WorklistSource. It reads the
// `scheduler` (appointment) table joined to patient / physician / facility and
// projects each row as one Modality Worklist scheduled procedure step.
//
// All SQL values are bound parameters; no value from the network is ever
// concatenated into a statement.
type mwlSource struct {
	db        *sql.DB
	stationAE string
	maxRows   int
	logf      func(format string, args ...any)
}

// mwlQuerySQL fetches the appointment rows (and their patient, performing
// physician and facility) that fall inside the requested date window. Matching
// beyond the window is performed in Go by the SCP, which keeps the statement
// free of any client supplied text.
//
// The patient join is an INNER JOIN on purpose: a worklist entry without a
// patient identity would let a modality image an unidentified person, so
// appointments whose patient row is missing (or soft deleted) are omitted.
const mwlQuerySQL = `
SELECT
    s.id,
    s.caldateof,
    s.calhour,
    s.calminute,
    s.calduration,
    COALESCE(s.calstatus, ''),
    COALESCE(s.caltype, ''),
    COALESCE(s.calprenote, ''),
    COALESCE(p.ptid, ''),
    COALESCE(p.ptlname, ''),
    COALESCE(p.ptfname, ''),
    COALESCE(p.ptmname, ''),
    p.ptdob,
    COALESCE(p.ptsex, ''),
    COALESCE(ph.phylname, ''),
    COALESCE(ph.phyfname, ''),
    COALESCE(ph.phymname, ''),
    COALESCE(f.psrname, '')
FROM scheduler s
JOIN patient p        ON p.id = s.calpatient    AND p.deleted_at IS NULL
LEFT JOIN physician ph ON ph.id = s.calphysician AND ph.deleted_at IS NULL
LEFT JOIN facility f  ON f.id = s.calfacility  AND f.deleted_at IS NULL
WHERE s.deleted_at IS NULL
  AND s.calpatient > 0
  AND p.ptid <> ''
  AND s.caldateof >= ?
  AND s.caldateof < ?
ORDER BY s.caldateof ASC, s.calhour ASC, s.calminute ASC
LIMIT ?`

// mwlDefaultWindowDays is the worklist window used when the SCU does not ask
// for an explicit date range.
const mwlDefaultWindowDays = 30

// mwlMaxWindowDays bounds the window a single query can pull, so a hostile or
// buggy SCU cannot make the server scan the entire appointment history.
const mwlMaxWindowDays = 366

// mwlDefaultMaxRows bounds the number of appointment rows fetched per query.
const mwlDefaultMaxRows = 2000

// newMwlSource builds a database backed worklist source. stationAE is the AE
// title advertised in ScheduledStationAETitle.
func newMwlSource(db *sql.DB, stationAE string) *mwlSource {
	return &mwlSource{
		db:        db,
		stationAE: stationAE,
		maxRows:   mwlDefaultMaxRows,
		logf:      log.Printf,
	}
}

// Find returns the worklist entries inside the query's date window.
func (s *mwlSource) Find(ctx context.Context, q *dicomnet.MWLQuery) ([]dicomnet.MWLItem, error) {
	if s.db == nil {
		return nil, fmt.Errorf("mwl: database is not configured")
	}

	start, end := mwlWindow(q)
	rows, err := s.db.QueryContext(ctx, mwlQuerySQL, start, end, s.maxRows)
	if err != nil {
		return nil, fmt.Errorf("mwl: querying appointments: %w", err)
	}
	defer rows.Close()

	// Cancelled appointments are not work: they are withheld unless the SCU
	// explicitly asks for CANCELLED steps.
	wantCancelled := false
	if v, ok := q.SPSMatchingValue(dicomnet.TagSPSStatus); ok && strings.Contains(strings.ToUpper(v), "CANCEL") {
		wantCancelled = true
	}

	var items []dicomnet.MWLItem
	for rows.Next() {
		var (
			id                           int64
			calDate                      time.Time
			calHour, calMinute           int64
			calDuration                  int64
			status, apptType, prenote    string
			ptid, ptlname, ptfname       string
			ptmname, ptsex               string
			ptdob                        sql.NullTime
			phylname, phyfname, phymname string
			facilityName                 string
		)
		if err := rows.Scan(
			&id, &calDate, &calHour, &calMinute, &calDuration,
			&status, &apptType, &prenote,
			&ptid, &ptlname, &ptfname, &ptmname, &ptdob, &ptsex,
			&phylname, &phyfname, &phymname,
			&facilityName,
		); err != nil {
			return nil, fmt.Errorf("mwl: scanning appointment row: %w", err)
		}
		item := s.item(id, calDate, calHour, calMinute, calDuration,
			status, apptType, prenote,
			ptid, ptlname, ptfname, ptmname, ptdob, ptsex,
			phylname, phyfname, phymname, facilityName)
		if item.ScheduledProcedureStepStatus == "CANCELLED" && !wantCancelled {
			continue
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("mwl: reading appointment rows: %w", err)
	}
	return items, nil
}

// mwlWindow computes the [start, end) datetime window for a query. An explicit
// date range is honoured (capped at mwlMaxWindowDays); otherwise a default
// window starting today is used.
func mwlWindow(q *dicomnet.MWLQuery) (time.Time, time.Time) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 0, mwlDefaultWindowDays)

	if q == nil {
		return start, end
	}
	dr, err := q.DateRange()
	if err != nil || dr.Universal {
		return start, end
	}
	if t, ok := parseMWLDay(dr.Start); ok {
		start = t
	}
	if t, ok := parseMWLDay(dr.End); ok {
		end = t.AddDate(0, 0, 1) // the end date is inclusive
	}
	if end.Before(start) || end.Equal(start) {
		end = start.AddDate(0, 0, 1)
	}
	if end.Sub(start) > mwlMaxWindowDays*24*time.Hour {
		end = start.AddDate(0, 0, mwlMaxWindowDays)
	}
	return start, end
}

// parseMWLDay parses a DICOM DA value into local midnight.
func parseMWLDay(v string) (time.Time, bool) {
	if v == "" {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation("20060102", v, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// item projects one scheduler row onto a Modality Worklist entry.
func (s *mwlSource) item(
	id int64,
	calDate time.Time,
	calHour, calMinute, calDuration int64,
	status, apptType, prenote string,
	ptid, ptlname, ptfname, ptmname string,
	ptdob sql.NullTime,
	ptsex string,
	phylname, phyfname, phymname string,
	facilityName string,
) dicomnet.MWLItem {
	// The scheduler table stores the appointment date in caldateof and the
	// time of day in calhour / calminute (the same convention the scheduler
	// API uses: CONCAT(LPAD(calhour,2,'0'), ':', LPAD(calminute,2,'0'))).
	dateOnly := time.Date(calDate.Year(), calDate.Month(), calDate.Day(), 0, 0, 0, 0, time.Local)
	startAt := dateOnly.Add(time.Duration(calHour)*time.Hour + time.Duration(calMinute)*time.Minute)
	duration := time.Duration(calDuration) * time.Minute
	if duration < 0 {
		duration = 0
	}
	endAt := startAt.Add(duration)

	desc := strings.TrimSpace(apptType)
	if prenote != "" {
		if desc != "" {
			desc += " - "
		}
		desc += strings.TrimSpace(prenote)
	}

	item := dicomnet.MWLItem{
		PatientName:                       mwlPersonName(ptlname, ptfname, ptmname),
		PatientID:                         mwlText(ptid, 64),
		PatientSex:                        mwlSex(ptsex),
		AccessionNumber:                   fmt.Sprintf("A%010d", id),
		StudyInstanceUID:                  mwlUID("study", id),
		StudyDescription:                  mwlText(desc, 64),
		InstitutionName:                   mwlText(facilityName, 64),
		RequestedProcedureID:              mwlText(fmt.Sprintf("RP%d", id), 16),
		RequestedProcedureDescription:     mwlText(desc, 64),
		Modality:                          mwlModality(apptType, prenote),
		ScheduledStationAETitle:           s.stationAE,
		ScheduledProcedureStepStartDate:   startAt.Format("20060102"),
		ScheduledProcedureStepStartTime:   startAt.Format("150405"),
		ScheduledProcedureStepEndDate:     endAt.Format("20060102"),
		ScheduledProcedureStepEndTime:     endAt.Format("150405"),
		ScheduledPerformingPhysicianName:  mwlPersonName(phylname, phyfname, phymname),
		ScheduledProcedureStepID:          mwlText(fmt.Sprintf("SPS%d", id), 16),
		ScheduledProcedureStepDescription: mwlText(desc, 64),
		ScheduledProcedureStepStatus:      mwlStepStatus(status),
	}
	if ptdob.Valid {
		item.PatientBirthDate = ptdob.Time.Format("20060102")
	}
	return item
}

// mwlUID builds a deterministic, valid DICOM UID for a database row. The
// 2.25.<decimal> form is the UUID/OID root recommended by PS3.5 Annex B.
func mwlUID(kind string, id int64) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("freemed-mwl-%s-%d", kind, id)))
	n := new(big.Int).SetBytes(sum[:16])
	return "2.25." + n.String()
}

// mwlPersonName formats a DICOM PN as Family^Given^Middle. Component
// separators found in the source data are replaced so that a value containing
// '^', '=' or '\' cannot forge extra PN components or a multi-valued
// attribute.
func mwlPersonName(last, first, middle string) string {
	parts := make([]string, 0, 3)
	for _, p := range []string{last, first, middle} {
		if p = sanitizePNComponent(p); p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, "^")
}

func sanitizePNComponent(s string) string {
	s = sanitizeDICOMText(s)
	if s == "" {
		return ""
	}
	replacer := strings.NewReplacer("^", " ", "=", " ")
	return strings.Join(strings.Fields(replacer.Replace(s)), " ")
}

// sanitizeDICOMText removes characters that would corrupt a DICOM string VR.
// The backslash separates the values of a multi-valued attribute, so it must
// never reach the wire from source data: it and other control characters
// become spaces. NUL is dropped.
func sanitizeDICOMText(s string) string {
	s = strings.Map(func(r rune) rune {
		switch {
		case r == 0x00:
			return -1
		case r == '\\' || r < 0x20 || r == 0x7f:
			return ' '
		}
		return r
	}, s)
	return strings.TrimSpace(s)
}

// mwlText sanitises and truncates a value for a DICOM short-string VR.
func mwlText(s string, max int) string {
	s = strings.Join(strings.Fields(sanitizeDICOMText(s)), " ")
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max]
}

// mwlSex maps the FreeMED patient sex field onto the DICOM CS values.
func mwlSex(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "m", "male":
		return "M"
	case "f", "female":
		return "F"
	case "o", "other":
		return "O"
	}
	return ""
}

// mwlStepStatus maps the scheduler status text onto the DICOM Scheduled
// Procedure Step Status values of (0040,0020).
func mwlStepStatus(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "":
		return "SCHEDULED"
	case "cancelled", "canceled", "cancel", "no show", "noshow", "no-show", "missed":
		return "CANCELLED"
	case "completed", "complete", "done", "seen", "checked out", "checkout":
		return "COMPLETED"
	case "arrived", "checked in", "checkin", "checked-in":
		return "ARRIVED"
	case "in progress", "started", "start", "wip", "roomed":
		return "STARTED"
	case "scheduled", "booked", "pending":
		return "SCHEDULED"
	}
	return "SCHEDULED"
}

// mwlModality derives a DICOM modality from the free-text appointment type and
// pre-note. The scheduler schema has no modality column, so this is a
// documented heuristic: unmatched appointments are advertised as OT (other).
func mwlModality(apptType, prenote string) string {
	text := strings.ToLower(apptType + " " + prenote)
	type rule struct {
		modality string
		terms    []string
	}
	// Order matters: the more specific terms come first.
	rules := []rule{
		{"PT", []string{"pet scan", "pet/ct", "pet-ct"}},
		{"MG", []string{"mammo"}},
		{"MR", []string{"mri", "magnetic resonance"}},
		{"CT", []string{"ct scan", "ctscan", "cat scan", "catscan", " ct ", "ct-", "ct /", "ct,"}},
		{"US", []string{"ultrasound", "ultrasono", "sonograph", "sono ", "echo ", "echocardio", "doppler"}},
		{"BMD", []string{"dexa", "bone density", "densitometry"}},
		{"XA", []string{"angio", "cath ", "catheterization", "fluoroscop", "fluoro "}},
		{"NM", []string{"nuclear", "spect", "scintigra"}},
		{"ECG", []string{"ekg", "ecg", "electrocardio", "stress test", "holter"}},
		{"DX", []string{"x-ray", "xray", "x ray", "radiograph", "chest film", " cxr", "cxr "}},
	}
	for _, r := range rules {
		for _, term := range r.terms {
			if strings.Contains(" "+text+" ", term) {
				return r.modality
			}
		}
	}
	return "OT"
}

// startMwlServer starts the Modality Worklist C-FIND SCP when it is enabled in
// the configuration (Mwl.Port > 0). It returns (nil, nil) when disabled.
func startMwlServer(ctx context.Context) (*dicomnet.Server, error) {
	cfg := config.Config.Mwl
	if cfg.Port <= 0 {
		return nil, nil
	}
	if cfg.AETitle == "" {
		return nil, fmt.Errorf("mwl: AE title is not configured")
	}

	source := newMwlSource(model.SqlDb, cfg.AETitle)
	srv, err := dicomnet.NewServer(dicomnet.Config{
		ListenAddr:      fmt.Sprintf(":%d", cfg.Port),
		AETitle:         cfg.AETitle,
		MaxAssociations: cfg.MaxAssociations,
		Source:          source,
		Logf:            log.Printf,
	})
	if err != nil {
		return nil, fmt.Errorf("mwl: %w", err)
	}

	go func() {
		if err := srv.ListenAndServe(ctx); err != nil {
			log.Printf("mwl: DICOM worklist SCP stopped: %v", err)
		}
	}()
	return srv, nil
}
