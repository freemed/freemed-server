package api

import (
	"log"
	"net/http"

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
			r.GET("/patient/:id", notificationsForPatient)
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
