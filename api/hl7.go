package api

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/freemed/freemed-server/pkg/hl7"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["hl7"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.POST("/adt", hl7ADTIngest)
			r.POST("/oru", hl7ORUIngest)
			r.POST("/parse", hl7Parse)
		},
	}
}

// hl7ADTIngest handles POST /api/hl7/adt
// Accepts a raw HL7v2 ADT message, extracts demographics, and creates or
// updates the patient record.
func hl7ADTIngest(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	msg, err := hl7.Unmarshal(data)
	if err != nil {
		log.Printf("hl7ADTIngest: parse error: %v", err)
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	pt, eventType, err := hl7.ParseADT(msg)
	if err != nil {
		log.Printf("hl7ADTIngest: ADT parse error: %v", err)
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	sess, err := common.GetSession(c)
	if err != nil {
		common.ErrorResponseFromError(c, http.StatusUnauthorized, err)
		return
	}

	// Look up patient by MRN (stored in ptid)
	patientID, found, err := findPatientByMRN(c, pt.MRN)
	if err != nil {
		log.Printf("hl7ADTIngest: find patient error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	status := "created"
	if found {
		status = "updated"
		if err := updatePatientFromHL7(c, patientID, pt); err != nil {
			log.Printf("hl7ADTIngest: update patient error: %v", err)
			common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
			return
		}
	} else {
		patientID, err = createPatientFromHL7(c, pt, sess.UserId)
		if err != nil {
			log.Printf("hl7ADTIngest: create patient error: %v", err)
			common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"patient_id": patientID,
		"event_type": eventType,
		"status":     status,
	})
}

// hl7ORUIngest handles POST /api/hl7/oru
// Accepts a raw HL7v2 ORU^R01 lab result message, parses observations,
// and inserts them into the labs table.
func hl7ORUIngest(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	msg, err := hl7.Unmarshal(data)
	if err != nil {
		log.Printf("hl7ORUIngest: parse error: %v", err)
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	lab, err := hl7.ParseORU(msg)
	if err != nil {
		log.Printf("hl7ORUIngest: ORU parse error: %v", err)
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	// Find patient by MRN
	patientID, _, err := findPatientByMRN(c, lab.PatientMRN)
	if err != nil {
		log.Printf("hl7ORUIngest: find patient error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	sess, err := common.GetSession(c)
	if err != nil {
		common.ErrorResponseFromError(c, http.StatusUnauthorized, err)
		return
	}

	// Parse result date
	labDate := parseHL7DateTime(lab.ResultDate)

	var firstLabID int64
	for i, obs := range lab.Observations {
		notes := sql.NullString{}
		if obs.AbnormalFlag != "" {
			notes = sql.NullString{String: "Abnormal flag: " + obs.AbnormalFlag, Valid: true}
		}

		result, err := model.Queries.CreateLab(c.Request.Context(), dbgen.CreateLabParams{
			PatientID:      patientID,
			LabName:        obs.TestName,
			LabDate:        labDate,
			Result:         obs.Value,
			Unit:           obs.Units,
			ReferenceRange: obs.RefRange,
			Status:         "final",
			Notes:          notes,
			UserID:         sess.UserId,
		})
		if err != nil {
			log.Printf("hl7ORUIngest: create lab error: %v", err)
			common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
			return
		}

		insertedID, _ := result.LastInsertId()
		if i == 0 {
			firstLabID = insertedID
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"lab_id":     firstLabID,
		"test_count": len(lab.Observations),
		"status":     "imported",
	})
}

// hl7Parse handles POST /api/hl7/parse
// Parses a raw HL7 message and returns its structured representation as JSON.
// Useful for debugging and troubleshooting HL7 integrations.
func hl7Parse(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	msg, err := hl7.Unmarshal(data)
	if err != nil {
		log.Printf("hl7Parse: parse error: %v", err)
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	// Convert segments to a JSON-friendly representation
	type segmentJSON struct {
		Name   string     `json:"name"`
		Fields [][]string `json:"fields"`
	}

	segments := make([]segmentJSON, 0, len(msg.Segments))
	for _, seg := range msg.Segments {
		segments = append(segments, segmentJSON{
			Name:   seg.Name,
			Fields: seg.Fields,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"segments": segments,
	})
}

// findPatientByMRN looks up a patient by the MRN value stored in ptid.
// Returns (patientID, found, error).
func findPatientByMRN(c *gin.Context, mrn string) (int64, bool, error) {
	if mrn == "" {
		return 0, false, nil
	}

	var id int64
	err := model.SqlDb.QueryRowContext(
		c.Request.Context(),
		"SELECT id FROM patient WHERE ptid = ? AND ptarchive = 0 LIMIT 1",
		mrn,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return id, true, nil
}

// createPatientFromHL7 inserts a new patient record from HL7 demographics.
func createPatientFromHL7(c *gin.Context, pt *hl7.PatientDemographics, userID int64) (int64, error) {
	ptdob := parseHL7Date(pt.DOB)

	result, err := model.Queries.PatientCreate(c.Request.Context(), dbgen.PatientCreateParams{
		Ptlname:  pt.LastName,
		Ptfname:  pt.FirstName,
		Ptmname:  sql.NullString{},
		Ptsuffix: "",
		Ptsex:    pt.Sex,
		Ptid:     pt.MRN,
		Ptdob:    ptdob,
		User:     userID,
	})
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// updatePatientFromHL7 updates an existing patient record from HL7 demographics.
func updatePatientFromHL7(c *gin.Context, patientID int64, pt *hl7.PatientDemographics) error {
	ptdob := parseHL7Date(pt.DOB)

	_, err := model.SqlDb.ExecContext(
		c.Request.Context(),
		`UPDATE patient SET
			ptlname = ?,
			ptfname = ?,
			ptsex = ?,
			ptdob = ?,
			ssn = ?,
			updated_at = NOW()
		WHERE id = ?`,
		pt.LastName, pt.FirstName, pt.Sex, ptdob, pt.SSN, patientID,
	)
	return err
}

// parseHL7Date parses an HL7 date string (YYYYMMDD or YYYYMMDDHHMMSS) to sql.NullTime.
func parseHL7Date(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{}
	}

	// Try full datetime first
	if len(s) >= 14 {
		t, err := time.Parse("20060102150405", s[:14])
		if err == nil {
			return sql.NullTime{Time: t, Valid: true}
		}
	}

	// Try date only
	if len(s) >= 8 {
		t, err := time.Parse("20060102", s[:8])
		if err == nil {
			return sql.NullTime{Time: t, Valid: true}
		}
	}

	return sql.NullTime{}
}

// parseHL7DateTime parses an HL7 date/time string (YYYYMMDDHHMMSS) to time.Time.
// Returns zero time if unparseable.
func parseHL7DateTime(s string) time.Time {
	if s == "" {
		return time.Now()
	}

	if len(s) >= 14 {
		t, err := time.Parse("20060102150405", s[:14])
		if err == nil {
			return t
		}
	}
	if len(s) >= 8 {
		t, err := time.Parse("20060102", s[:8])
		if err == nil {
			return t
		}
	}

	return time.Now()
}
