package api

import (
	"log"
	"net/http"
	"time"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

type annotationInput struct {
	Module     string `json:"module" binding:"required"`
	TableName  string `json:"table_name" binding:"required"`
	RecordID   int64  `json:"record_id" binding:"required"`
	Annotation string `json:"annotation" binding:"required"`
}

func patientAnnotationsList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	rows, err := model.Queries.ListAnnotations(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, rows)
}

func patientAnnotationCreate(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	session, err := common.GetSession(r)
	if err != nil {
		log.Printf("patientAnnotationCreate: failed to get session: %v", err)
		common.ErrorResponseFromError(r, http.StatusUnauthorized, err)
		return
	}

	var in annotationInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	result, err := model.Queries.CreateAnnotation(r.Request.Context(), dbgen.CreateAnnotationParams{
		Atimestamp: time.Now(),
		PatientID:  patientID,
		Amodule:    in.Module,
		Atable:     in.TableName,
		Aid:        in.RecordID,
		UserID:     session.UserId,
		Annotation: in.Annotation,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}
