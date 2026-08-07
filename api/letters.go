package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

// LetterInput is the JSON payload for creating a letter record.
type LetterInput struct {
	LetterType string `json:"letter_type"`
	Recipient  string `json:"recipient"`
	Subject    string `json:"subject"`
	Body       string `json:"body"`
	DateSent   string `json:"date_sent"`
}

func patientLettersList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	letters, err := model.Queries.ListLetters(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, letters)
}

func patientLettersCreate(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	session, err := common.GetSession(r)
	if err != nil {
		log.Printf("patientLettersCreate: failed to get session: %v", err)
		r.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	var input LetterInput
	if err := r.BindJSON(&input); err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	dateSent := parseOptionalDate(input.DateSent)

	params := dbgen.CreateLetterParams{
		PatientID:  common.ParseInt(id),
		LetterType: input.LetterType,
		Recipient:  input.Recipient,
		Subject:    input.Subject,
		Body: toNullString(&input.Body),
		DateSent:   dateSent,
		UserID:     session.UserId,
	}

	result, err := model.Queries.CreateLetter(r.Request.Context(), params)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}
