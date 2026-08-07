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

// patientPreviousOperationsList handles GET /api/patient/:id/surgical-history
func patientPreviousOperationsList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	patientID := common.ParseInt(id)
	rows, err := model.Queries.ListPreviousOperationsByPatient(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, rows)
}

type previousOperationInput struct {
	OperationDate string `json:"operation_date" binding:"required"`
	Operation     string `json:"operation" binding:"required"`
}

// patientPreviousOperationCreate handles POST /api/patient/:id/surgical-history
func patientPreviousOperationCreate(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	var in previousOperationInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	// Extract user ID from JWT session
	sess, err := common.GetSession(r)
	if err != nil {
		common.ErrorResponseFromError(r, http.StatusUnauthorized, err)
		return
	}

	opDate, err := common.ParseDate(in.OperationDate)
	if err != nil {
		common.ErrorResponse(r, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
		return
	}

	result, err := model.Queries.CreatePreviousOperation(r.Request.Context(), dbgen.CreatePreviousOperationParams{
		PatientID:     patientID,
		OperationDate: sql.NullTime{Time: opDate, Valid: true},
		Operation:     in.Operation,
		UserID:        sess.UserId,
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
