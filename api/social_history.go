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

type socialHistoryInput struct {
	SmokingStatus        string `json:"smoking_status"`
	SmokingDetail        string `json:"smoking_detail"`
	AlcoholUse           string `json:"alcohol_use"`
	AlcoholDetail        string `json:"alcohol_detail"`
	DrugUse              string `json:"drug_use"`
	DrugDetail           string `json:"drug_detail"`
	ExerciseFrequency    string `json:"exercise_frequency"`
	Occupation           string `json:"occupation"`
	LivingSituation      string `json:"living_situation"`
	FoodInsecurity       bool   `json:"food_insecurity"`
	TransportationAccess bool   `json:"transportation_access"`
	Notes                string `json:"notes"`
	RecordedDate         string `json:"recorded_date" binding:"required"`
}

func patientSocialHistoryList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	patientID := common.ParseInt(id)
	rows, err := model.Queries.ListSocialHistory(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}

func patientSocialHistoryLatest(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	patientID := common.ParseInt(id)
	row, err := model.Queries.GetLatestSocialHistory(r.Request.Context(), patientID)
	if err != nil {
		if err == sql.ErrNoRows {
			r.JSON(http.StatusOK, nil)
			return
		}
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, row)
}

func patientSocialHistoryCreate(r *gin.Context) {
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

	var in socialHistoryInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	recordedDate, err := time.Parse("2006-01-02", in.RecordedDate)
	if err != nil {
		common.ErrorResponse(r, http.StatusBadRequest, "invalid recorded_date format, expected YYYY-MM-DD")
		return
	}

	result, err := model.Queries.CreateSocialHistory(r.Request.Context(), dbgen.CreateSocialHistoryParams{
		Patient:              patientID,
		SmokingStatus:        in.SmokingStatus,
		SmokingDetail:        in.SmokingDetail,
		AlcoholUse:           in.AlcoholUse,
		AlcoholDetail:        in.AlcoholDetail,
		DrugUse:              in.DrugUse,
		DrugDetail:           in.DrugDetail,
		ExerciseFrequency:    in.ExerciseFrequency,
		Occupation:           in.Occupation,
		LivingSituation:      in.LivingSituation,
		FoodInsecurity:       in.FoodInsecurity,
		TransportationAccess: in.TransportationAccess,
		Notes:                strToNullString(in.Notes),
		RecordedDate:         recordedDate,
		User:                 sess.UserId,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}

func patientSocialHistoryRemove(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}
	itemID := common.ParseInt(r.Param("itemId"))
	if itemID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	err := model.Queries.RemoveSocialHistory(r.Request.Context(), dbgen.RemoveSocialHistoryParams{
		ID:      itemID,
		Patient: patientID,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, gin.H{"status": "deactivated"})
}
