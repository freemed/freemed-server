package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

// currentProblemInput is the JSON payload for creating a current problem record.
type currentProblemInput struct {
	Date    string `json:"date" binding:"required"`
	Problem string `json:"problem" binding:"required"`
}

// patientCurrentProblemsList handles GET /api/patient/:id/current-problems
func patientCurrentProblemsList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	patientID := common.ParseInt(id)
	problems, err := model.Queries.ListCurrentProblemsByPatient(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, problems)
}

// patientCurrentProblemsCreate handles POST /api/patient/:id/current-problems
func patientCurrentProblemsCreate(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	session, err := common.GetSession(r)
	if err != nil {
		log.Printf("patientCurrentProblemsCreate: failed to get session: %v", err)
		common.ErrorResponseFromError(r, http.StatusUnauthorized, err)
		return
	}

	var in currentProblemInput
	if err := r.BindJSON(&in); err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	dateVal, err := common.ParseDate(in.Date)
	if err != nil {
		common.ErrorResponse(r, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
		return
	}

	result, err := model.Queries.CreateCurrentProblem(r.Request.Context(), dbgen.CreateCurrentProblemParams{
		PatientID: patientID,
		Date:      dateVal,
		Problem:   in.Problem,
		UserID:    session.UserId,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}
