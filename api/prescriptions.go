package api

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["prescriptions"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/patient/:id/prescriptions", prescriptionsList)
			r.POST("/patient/:id/prescriptions", prescriptionCreate)
			r.DELETE("/prescriptions/:id", prescriptionDiscontinue)
		},
	}
}

type prescriptionInput struct {
	DrugName            string `json:"drug_name" binding:"required"`
	Dosage              string `json:"dosage"`
	Frequency           string `json:"frequency"`
	Quantity            string `json:"quantity"`
	Refills             int64  `json:"refills"`
	DateWritten         string `json:"date_written"`
	PrescribingProvider int64  `json:"prescribing_provider"`
	Pharmacy            string `json:"pharmacy"`
	Notes               string `json:"notes"`
}

func prescriptionsList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)

	prescriptions, err := model.Queries.ListPrescriptions(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, prescriptions)
}

func prescriptionCreate(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var in prescriptionInput
	if err := r.BindJSON(&in); err != nil {
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	dateWritten := parseOptionalDate(in.DateWritten)
	if !dateWritten.Valid {
		dateWritten = sql.NullTime{Time: time.Now(), Valid: true}
	}

	session, _ := common.GetSession(r)
	var userId int64
	if session.UserId > 0 {
		userId = session.UserId
	}

	result, err := model.Queries.CreatePrescription(r.Request.Context(), dbgen.CreatePrescriptionParams{
		Patient:             patientID,
		DrugName:            in.DrugName,
		Dosage:              in.Dosage,
		Frequency:           in.Frequency,
		Quantity:            in.Quantity,
		Refills:             in.Refills,
		DateWritten:         dateWritten.Time,
		PrescribingProvider: in.PrescribingProvider,
		Pharmacy:            in.Pharmacy,
		Status:              "active",
		Notes:               sql.NullString{String: in.Notes, Valid: in.Notes != ""},
		User:                userId,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}

func prescriptionDiscontinue(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err := model.Queries.DiscontinuePrescription(r.Request.Context(), id)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, true)
}
