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

type photoIDInput struct {
	Photo       string `json:"photo"`
	PhotoMime   string `json:"photo_mime"`
	PageCount   int64  `json:"page_count"`
	Description string `json:"description"`
}

// patientPhotoIDList handles GET /api/patient/:id/photo-id
func patientPhotoIDList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	patientID := common.ParseInt(id)
	rows, err := model.Queries.GetPhotoID(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, rows)
}

// patientPhotoIDCreate handles POST /api/patient/:id/photo-id
func patientPhotoIDCreate(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	sess, err := common.GetSession(r)
	if err != nil {
		log.Printf("patientPhotoIDCreate: failed to get session: %v", err)
		common.ErrorResponseFromError(r, http.StatusUnauthorized, err)
		return
	}

	var in photoIDInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	var photoData sql.NullString
	if in.Photo != "" {
		photoData = sql.NullString{String: in.Photo, Valid: true}
	}

	pageCount := in.PageCount
	if pageCount == 0 {
		pageCount = 1
	}

	result, err := model.Queries.CreatePhotoID(r.Request.Context(), dbgen.CreatePhotoIDParams{
		Patient:     patientID,
		Photo:       photoData,
		PhotoMime:   in.PhotoMime,
		PageCount:   pageCount,
		Description: in.Description,
		User:        sess.UserId,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}
