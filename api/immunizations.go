package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

// patientImmunizationsList handles GET /api/patient/:id/immunizations
func patientImmunizationsList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	immunizations, err := model.Queries.ListImmunizations(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, immunizations)
}

type immunizationInput struct {
	Dateof        string `json:"dateof" binding:"required"`
	Provider      int64  `json:"provider"`
	AdminProvider int64  `json:"admin_provider"`
	Immunization  int64  `json:"immunization" binding:"required"`
	Route         int64  `json:"route"`
	BodySite      int64  `json:"body_site"`
	Manufacturer  string `json:"manufacturer"`
	LotNumber     string `json:"lot_number"`
	PreviousDoses int64  `json:"previous_doses"`
	Recovered     bool   `json:"recovered"`
	Notes         string `json:"notes"`
}

// patientImmunizationsCreate handles POST /api/patient/:id/immunizations
func patientImmunizationsCreate(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var in immunizationInput
	if err := r.BindJSON(&in); err != nil {
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	dateof, err := common.ParseDate(in.Dateof)
	if err != nil {
		r.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	result, err := model.Queries.CreateImmunization(r.Request.Context(), dbgen.CreateImmunizationParams{
		Dateof:        dateof,
		PatientID:     patientID,
		Provider:      in.Provider,
		AdminProvider: in.AdminProvider,
		Immunization:  in.Immunization,
		Route:         in.Route,
		BodySite:      in.BodySite,
		Manufacturer:  toNullString(&in.Manufacturer),
		LotNumber:     toNullString(&in.LotNumber),
		PreviousDoses: in.PreviousDoses,
		Recovered:     in.Recovered,
		Notes:         toNullString(&in.Notes),
		UserID:        0, // TODO: populate from session
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newID, err := result.LastInsertId()
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusCreated, gin.H{"id": newID})
}
