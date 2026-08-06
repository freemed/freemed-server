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
	common.ApiMap["medications"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/patient/:id/medications", medicationsList)
			r.POST("/patient/:id/medications", medicationCreate)
			r.GET("/medications/:id", medicationGet)
			r.PUT("/medications/:id", medicationUpdate)
			r.DELETE("/medications/:id", medicationDiscontinue)
		},
	}
}

type medicationInput struct {
	DrugName            string `json:"drug_name" binding:"required"`
	Dosage              string `json:"dosage"`
	Frequency           string `json:"frequency"`
	StartDate           string `json:"start_date"`
	EndDate             string `json:"end_date"`
	PrescribingProvider int64  `json:"prescribing_provider"`
}

func parseOptionalDate(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{Valid: false}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func medicationsList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)

	// Support ?all=1 to include inactive meds
	allMeds := r.Query("all") == "1"

	var meds []dbgen.Medication
	var err error
	if allMeds {
		meds, err = model.Queries.ListAllMedications(r.Request.Context(), patientID)
	} else {
		meds, err = model.Queries.ListMedications(r.Request.Context(), patientID)
	}
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, meds)
}

func medicationGet(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	med, err := model.Queries.GetMedicationById(r.Request.Context(), id)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, med)
}

func medicationCreate(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var in medicationInput
	if err := r.BindJSON(&in); err != nil {
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	result, err := model.Queries.CreateMedication(r.Request.Context(), dbgen.CreateMedicationParams{
		Patient:             patientID,
		DrugName:            in.DrugName,
		Dosage:              in.Dosage,
		Frequency:           in.Frequency,
		StartDate:           parseOptionalDate(in.StartDate),
		EndDate:             parseOptionalDate(in.EndDate),
		PrescribingProvider: in.PrescribingProvider,
		Active:              "active",
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}

func medicationUpdate(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var in medicationInput
	if err := r.BindJSON(&in); err != nil {
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err := model.Queries.UpdateMedication(r.Request.Context(), dbgen.UpdateMedicationParams{
		ID:                  id,
		DrugName:            in.DrugName,
		Dosage:              in.Dosage,
		Frequency:           in.Frequency,
		StartDate:           parseOptionalDate(in.StartDate),
		EndDate:             parseOptionalDate(in.EndDate),
		PrescribingProvider: in.PrescribingProvider,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, true)
}

func medicationDiscontinue(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err := model.Queries.DiscontinueMedication(r.Request.Context(), dbgen.DiscontinueMedicationParams{
		ID:      id,
		EndDate: sql.NullTime{Time: time.Now(), Valid: true},
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, true)
}
