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

// familyHistoryInput is the JSON payload for creating/updating a family history record.
type familyHistoryInput struct {
	Relationship  string `json:"relationship" binding:"required"`
	ConditionName string `json:"condition_name" binding:"required"`
	Icd10Code     string `json:"icd10_code"`
	OnsetAge      int32  `json:"onset_age"`
	Deceased      bool   `json:"deceased"`
	Notes         string `json:"notes"`
}

// familyHistoryNullString converts a string to sql.NullString, treating empty as NULL.
func familyHistoryNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// patientFamilyHistoryList handles GET /api/patient/:id/family-history
func patientFamilyHistoryList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	patientID := common.ParseInt(id)
	rows, err := model.Queries.ListFamilyHistory(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, rows)
}

// patientFamilyHistoryCreate handles POST /api/patient/:id/family-history
func patientFamilyHistoryCreate(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	session, err := common.GetSession(r)
	if err != nil {
		log.Printf("patientFamilyHistoryCreate: failed to get session: %v", err)
		common.ErrorResponseFromError(r, http.StatusUnauthorized, err)
		return
	}

	var in familyHistoryInput
	if err := r.BindJSON(&in); err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	result, err := model.Queries.CreateFamilyHistory(r.Request.Context(), dbgen.CreateFamilyHistoryParams{
		Patient:       patientID,
		Relationship:  in.Relationship,
		ConditionName: in.ConditionName,
		Icd10Code:     in.Icd10Code,
		OnsetAge:      in.OnsetAge,
		Deceased:      in.Deceased,
		Notes:         familyHistoryNullString(in.Notes),
		User:          session.UserId,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}

// patientFamilyHistoryUpdate handles PUT /api/patient/:id/family-history/:itemId
func patientFamilyHistoryUpdate(r *gin.Context) {
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

	var in familyHistoryInput
	if err := r.BindJSON(&in); err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	if err := model.Queries.UpdateFamilyHistory(r.Request.Context(), dbgen.UpdateFamilyHistoryParams{
		Relationship:  in.Relationship,
		ConditionName: in.ConditionName,
		Icd10Code:     in.Icd10Code,
		OnsetAge:      in.OnsetAge,
		Deceased:      in.Deceased,
		Notes:         familyHistoryNullString(in.Notes),
		ID:            itemID,
		Patient:       patientID,
	}); err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// patientFamilyHistoryRemove handles DELETE /api/patient/:id/family-history/:itemId
func patientFamilyHistoryRemove(r *gin.Context) {
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

	if err := model.Queries.RemoveFamilyHistory(r.Request.Context(), dbgen.RemoveFamilyHistoryParams{
		ID:      itemID,
		Patient: patientID,
	}); err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{"status": "deactivated"})
}
