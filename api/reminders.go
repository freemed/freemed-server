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

func init() {
	common.ApiMap["reminders"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", listReminders)
			r.POST("/", createReminder)
			r.PUT("/:id/status", updateReminderStatus)
			r.DELETE("/:id", deleteReminder)
		},
	}
}

// listReminders handles GET /api/reminders
// Optional query parameter: ?patient=<id> to filter by patient
func listReminders(r *gin.Context) {
	session, err := common.GetSession(r)
	if err != nil {
		common.ErrorResponseFromError(r, http.StatusUnauthorized, err)
		return
	}

	// Optional patient filter
	if patientStr := r.Query("patient"); patientStr != "" {
		patientID := common.ParseInt(patientStr)
		rows, err := model.Queries.ListRemindersByPatient(r.Request.Context(), sql.NullInt64{Int64: patientID, Valid: true})
		if err != nil {
			log.Print(err.Error())
			common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
			return
		}
		r.JSON(http.StatusOK, rows)
		return
	}

	rows, err := model.Queries.ListRemindersByUser(r.Request.Context(), session.UserId)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}

type reminderInput struct {
	PatientID   *int64  `json:"patient"`
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description"`
	DueDate     *string `json:"due_date"` // YYYY-MM-DD HH:MM:SS
	Priority    *int32  `json:"priority"`
}

// createReminder handles POST /api/reminders
func createReminder(r *gin.Context) {
	session, err := common.GetSession(r)
	if err != nil {
		common.ErrorResponseFromError(r, http.StatusUnauthorized, err)
		return
	}

	var in reminderInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	params := dbgen.CreateReminderParams{
		UserID:   session.UserId,
		Title:    in.Title,
		Priority: 0,
	}

	if in.PatientID != nil {
		params.PatientID = sql.NullInt64{Int64: *in.PatientID, Valid: true}
	}
	if in.Description != nil {
		params.Description = sql.NullString{String: *in.Description, Valid: true}
	}
	if in.DueDate != nil {
		dueDate, err := common.ParseDate(*in.DueDate)
		if err == nil {
			params.DueDate = sql.NullTime{Time: dueDate, Valid: true}
		}
	}
	if in.Priority != nil {
		params.Priority = *in.Priority
	}

	result, err := model.Queries.CreateReminder(r.Request.Context(), params)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}

type reminderStatusInput struct {
	Status string `json:"status" binding:"required"`
}

// updateReminderStatus handles PUT /api/reminders/:id/status
func updateReminderStatus(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	reminderID := common.ParseInt(id)

	var in reminderStatusInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	if in.Status != "pending" && in.Status != "completed" && in.Status != "dismissed" {
		common.ErrorResponse(r, http.StatusBadRequest, "status must be pending, completed, or dismissed")
		return
	}

	err := model.Queries.UpdateReminderStatus(r.Request.Context(), dbgen.UpdateReminderStatusParams{
		ReminderID: reminderID,
		Status:     in.Status,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, true)
}

// deleteReminder handles DELETE /api/reminders/:id
func deleteReminder(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	reminderID := common.ParseInt(id)

	err := model.Queries.DeleteReminder(r.Request.Context(), reminderID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, true)
}
