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

type eocInput struct {
	StartDate   string  `json:"start_date" binding:"required"`
	EndDate     string  `json:"end_date"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Provider    int64   `json:"provider_id"`
	Notes       *string `json:"notes"`
}

// patientEOCsList handles GET /api/patient/:id/episodes-of-care
func patientEOCsList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	patientID := common.ParseInt(id)
	eocs, err := model.Queries.ListEOCs(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, eocs)
}

// patientEOCCreate handles POST /api/patient/:id/episodes-of-care
func patientEOCCreate(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	patientID := common.ParseInt(id)

	var in eocInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	sess, err := common.GetSession(r)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	var endDate sql.NullTime
	if in.EndDate != "" {
		endDate = parseOptionalDate(in.EndDate)
	}

	result, err := model.Queries.CreateEOC(r.Request.Context(), dbgen.CreateEOCParams{
		PatientID:   patientID,
		StartDate:   parseOptionalDate(in.StartDate).Time,
		EndDate:     endDate,
		Description: in.Description,
		Status:      in.Status,
		ProviderID:  in.Provider,
		Notes:       derefString(in.Notes),
		UserID:      sess.UserId,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	newID, err := result.LastInsertId()
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusCreated, gin.H{"id": newID})
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
