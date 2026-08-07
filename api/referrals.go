package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

// ReferralInput is the JSON payload for creating a referral record.
type ReferralInput struct {
	ReferringProvider int64  `json:"referring_provider"`
	ReferralTo        int64  `json:"referral_to" binding:"required"`
	ReferralType      string `json:"referral_type"`
	Reason            string `json:"reason"`
	Status            string `json:"status"`
	DateReferred      string `json:"date_referred" binding:"required"`
	DateCompleted     string `json:"date_completed"`
	Notes             string `json:"notes"`
}

func patientReferralsList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	referrals, err := model.Queries.ListReferrals(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, referrals)
}

func patientReferralsCreate(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	session, err := common.GetSession(r)
	if err != nil {
		log.Printf("patientReferralsCreate: failed to get session: %v", err)
		r.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	var input ReferralInput
	if err := r.BindJSON(&input); err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	dateReferred, err := common.ParseDate(input.DateReferred)
	if err != nil {
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	params := dbgen.CreateReferralParams{
		PatientID:          common.ParseInt(id),
		ReferringProvider:  input.ReferringProvider,
		ReferralTo:         input.ReferralTo,
		ReferralType:       input.ReferralType,
		Reason:             input.Reason,
		Status:             input.Status,
		DateReferred:       dateReferred,
		DateCompleted:      parseOptionalDate(input.DateCompleted),
		Notes:              input.Notes,
		UserID:             session.UserId,
	}

	result, err := model.Queries.CreateReferral(r.Request.Context(), params)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}
