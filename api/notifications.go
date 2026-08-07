package api

import (
	"log"
	"net/http"
	"time"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["notifications"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("", notificationsList)
			r.GET("/unread-count", notificationsUnreadCount)
			r.GET("/from", notificationsFromTimestamp)
			r.GET("/timestamp", notificationsLatestTimestamp)
			r.GET("/task-inbox/count", notificationsTaskInboxCount)
			r.GET("/patient/:id", notificationsForPatient)
			r.GET("/patient/:id/tasks", notificationsPatientTasks)
			r.GET("/patient/:id/tasks/count", notificationsPatientTasksCount)
		},
	}
}

// notificationsList handles GET /api/notifications
func notificationsList(c *gin.Context) {
	session, err := common.GetSession(c)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	notifications, err := model.Queries.UserNotifications(c.Request.Context(), session.UserId)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// notificationsUnreadCount handles GET /api/notifications/unread-count
func notificationsUnreadCount(c *gin.Context) {
	session, err := common.GetSession(c)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	count, err := model.Queries.UnreadCount(c.Request.Context(), session.UserId)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// notificationsFromTimestamp handles GET /api/notifications/from?timestamp=...
func notificationsFromTimestamp(c *gin.Context) {
	ts := c.Query("timestamp")
	if ts == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "timestamp query parameter required"})
		return
	}

	since, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		// Also try common datetime format
		since, err = time.Parse("2006-01-02 15:04:05", ts)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid timestamp format, use RFC3339 or YYYY-MM-DD HH:MM:SS"})
			return
		}
	}

	notifications, err := model.Queries.NotificationsFromTimestamp(c.Request.Context(), since)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// notificationsLatestTimestamp handles GET /api/notifications/timestamp
func notificationsLatestTimestamp(c *gin.Context) {
	ts, err := model.Queries.LatestTimestamp(c.Request.Context())
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"timestamp": ts})
}

// notificationsTaskInboxCount handles GET /api/notifications/task-inbox/count
func notificationsTaskInboxCount(c *gin.Context) {
	session, err := common.GetSession(c)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	count, err := model.Queries.SystemTaskInboxCount(c.Request.Context(), session.UserId)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// notificationsForPatient handles GET /api/notifications/patient/:id
func notificationsForPatient(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	notifications, err := model.Queries.PatientNotifications(c.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// notificationsPatientTasks handles GET /api/notifications/patient/:id/tasks
func notificationsPatientTasks(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	notifications, err := model.Queries.SystemTaskPatientInbox(c.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// notificationsPatientTasksCount handles GET /api/notifications/patient/:id/tasks/count
func notificationsPatientTasksCount(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	count, err := model.Queries.SystemTaskPatientInboxCount(c.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}
