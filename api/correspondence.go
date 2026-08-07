package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

// CorrespondenceInput is the JSON payload for creating a correspondence record.
type CorrespondenceInput struct {
	CorrespondenceType string `json:"correspondence_type"`
	Direction          string `json:"direction"`
	ContactName        string `json:"contact_name"`
	ContactMethod      string `json:"contact_method"`
	Summary            string `json:"summary"`
	Date               string `json:"date"`
}

func patientCorrespondenceList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	items, err := model.Queries.ListCorrespondence(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, items)
}

func patientCorrespondenceCreate(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	session, err := common.GetSession(r)
	if err != nil {
		log.Printf("patientCorrespondenceCreate: failed to get session: %v", err)
		r.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	var input CorrespondenceInput
	if err := r.BindJSON(&input); err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	corrDate := parseOptionalDate(input.Date)

	params := dbgen.CreateCorrespondenceParams{
		PatientID:          common.ParseInt(id),
		CorrespondenceType: input.CorrespondenceType,
		Direction:          input.Direction,
		ContactName:        input.ContactName,
		ContactMethod:      input.ContactMethod,
		Summary: toNullString(input.Summary),
		Date:               corrDate,
		UserID:             session.UserId,
	}

	result, err := model.Queries.CreateCorrespondence(r.Request.Context(), params)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}
