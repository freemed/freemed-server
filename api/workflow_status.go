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

type workflowStatusInput struct {
	StatusType      int64  `json:"status_type" binding:"required"`
	StatusCompleted bool   `json:"status_completed"`
	Stamp           string `json:"stamp"`
}

// patientWorkflowStatusList handles GET /api/patient/:id/workflow-status
func patientWorkflowStatusList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)

	// Check for date filter query parameter
	dateStr := r.Query("date")
	if dateStr != "" {
		statusDate, err := common.ParseDate(dateStr)
		if err != nil {
			r.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
			return
		}
		rows, err := model.Queries.ListWorkflowStatusForDate(r.Request.Context(), dbgen.ListWorkflowStatusForDateParams{
			PatientID:  patientID,
			StatusDate: statusDate,
		})
		if err != nil {
			log.Print(err.Error())
			r.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		r.JSON(http.StatusOK, rows)
		return
	}

	rows, err := model.Queries.ListWorkflowStatus(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, rows)
}

// patientWorkflowStatusCreate handles POST /api/patient/:id/workflow-status
func patientWorkflowStatusCreate(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var in workflowStatusInput
	if err := r.BindJSON(&in); err != nil {
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	sess, err := common.GetSession(r)
	if err != nil {
		r.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	var stamp time.Time
	if in.Stamp != "" {
		stamp, err = common.ParseDate(in.Stamp)
		if err != nil {
			r.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
			return
		}
	} else {
		stamp = time.Now()
	}

	result, err := model.Queries.SetWorkflowStatus(r.Request.Context(), dbgen.SetWorkflowStatusParams{
		PatientID:       patientID,
		UserID:          sess.UserId,
		StatusType:      in.StatusType,
		StatusCompleted: in.StatusCompleted,
		Stamp:           stamp,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newID, err := result.LastInsertId()
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusCreated, gin.H{"id": newID})
}
