package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["portal"] = common.ApiMapping{
		Authenticated: false, // portal uses its own JWT auth, verified per-request
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/me", portalMe)
			r.GET("/appointments", portalAppointments)
			r.POST("/appointments/request", PortalAppointmentRequest)
			r.GET("/medications", portalMedications)
			r.GET("/allergies", portalAllergies)
			r.GET("/vitals", portalVitals)
			r.GET("/problems", portalProblems)
			r.GET("/labs", portalLabs)
			r.GET("/documents", portalDocuments)
		},
	}
}

// ============================================================================
// GET /api/portal/me — Patient demographics
// ============================================================================

func portalMe(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	row, err := model.Queries.GetPortalPatientDemographics(c.Request.Context(), patientID)
	if err != nil {
		if err == sql.ErrNoRows {
			common.ErrorResponse(c, http.StatusNotFound, "patient not found")
			return
		}
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	resp := gin.H{
		"id":             row.ID,
		"first_name":     row.FirstName,
		"last_name":      row.LastName,
		"patient_id":     row.PatientIDDisplay,
		"gender":         row.Gender,
		"language":       row.Language,
		"portal_enabled": row.PortalEnabled,
	}
	if row.MiddleName.Valid {
		resp["middle_name"] = row.MiddleName.String
	}
	if row.Suffix != "" {
		resp["suffix"] = row.Suffix
	}
	if row.DateOfBirth.Valid {
		resp["date_of_birth"] = row.DateOfBirth.Time.Format("2006-01-02")
	}
	if row.Email.Valid {
		resp["email"] = row.Email.String
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// GET /api/portal/appointments — Patient appointments
// ============================================================================

func portalAppointments(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	rows, err := model.Queries.ListPortalAppointments(c.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	// Ensure non-nil array in JSON response
	if rows == nil {
		rows = []dbgen.ListPortalAppointmentsRow{}
	}

	c.JSON(http.StatusOK, rows)
}

// ============================================================================
// GET /api/portal/medications — Active medications for the patient
// ============================================================================

func portalMedications(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	rows, err := model.Queries.ListMedications(c.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	if rows == nil {
		rows = []dbgen.Medication{}
	}

	c.JSON(http.StatusOK, rows)
}

// ============================================================================
// GET /api/portal/allergies — Active allergies for the patient
// ============================================================================

func portalAllergies(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	rows, err := model.Queries.ListAllergies(c.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	if rows == nil {
		rows = []dbgen.Allergy{}
	}

	c.JSON(http.StatusOK, rows)
}

// ============================================================================
// GET /api/portal/vitals — Vitals history for the patient
// ============================================================================

func portalVitals(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	rows, err := model.Queries.ListVitals(c.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	if rows == nil {
		rows = []dbgen.Vital{}
	}

	c.JSON(http.StatusOK, rows)
}

// ============================================================================
// GET /api/portal/problems — Combined current and chronic problems
// ============================================================================

func portalProblems(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	rows, err := model.Queries.ListPortalProblems(c.Request.Context(), dbgen.ListPortalProblemsParams{
		PatientID: patientID,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	if rows == nil {
		rows = []dbgen.ListPortalProblemsRow{}
	}

	c.JSON(http.StatusOK, rows)
}

// ============================================================================
// GET /api/portal/labs — Lab results for the patient
// ============================================================================

func portalLabs(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	rows, err := model.Queries.ListLabs(c.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	if rows == nil {
		rows = []dbgen.Lab{}
	}

	c.JSON(http.StatusOK, rows)
}

// ============================================================================
// GET /api/portal/documents — Scanned documents for the patient
// ============================================================================

func portalDocuments(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	rows, err := model.Queries.ListScannedDocs(c.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	if rows == nil {
		rows = []dbgen.ScannedDoc{}
	}

	c.JSON(http.StatusOK, rows)
}
