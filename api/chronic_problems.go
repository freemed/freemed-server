package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

type chronicProblemInput struct {
	Problem string `json:"problem" binding:"required"`
	Date    string `json:"date" binding:"required"`
}

// patientChronicProblemsList handles GET /api/patient/:id/chronic-problems
func patientChronicProblemsList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	patientID := common.ParseInt(id)
	rows, err := model.Queries.ListChronicProblemsByPatient(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, rows)
}

// patientChronicProblemCreate handles POST /api/patient/:id/chronic-problems
func patientChronicProblemCreate(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	sess, err := common.GetSession(r)
	if err != nil {
		common.ErrorResponseFromError(r, http.StatusUnauthorized, err)
		return
	}

	var in chronicProblemInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	chronicDate, err := common.ParseDate(in.Date)
	if err != nil {
		common.ErrorResponse(r, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
		return
	}

	result, err := model.Queries.CreateChronicProblem(r.Request.Context(), dbgen.CreateChronicProblemParams{
		PatientID: patientID,
		Date:      chronicDate,
		Problem:   in.Problem,
		UserID:    sess.UserId,
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
