package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["dashboard"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", dashboardGet)
		},
	}
}

func dashboardGet(c *gin.Context) {
	session, err := common.GetSession(c)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	ctx := c.Request.Context()

	// Run all four queries concurrently using goroutines + channels
	type countResult struct {
		name  string
		value int64
		err   error
	}

	ch := make(chan countResult, 4)

	go func() {
		v, err := model.Queries.DashboardPatientCount(ctx)
		ch <- countResult{"patientCount", v, err}
	}()
	go func() {
		v, err := model.Queries.DashboardTodayAppointmentsCount(ctx)
		ch <- countResult{"todayAppointments", v, err}
	}()
	go func() {
		v, err := model.Queries.DashboardUnreadMessagesCount(ctx, session.UserId)
		ch <- countResult{"unreadMessages", v, err}
	}()
	go func() {
		v, err := model.Queries.DashboardActiveAuthorizationsCount(ctx)
		ch <- countResult{"activeAuthorizations", v, err}
	}()

	result := gin.H{}
	for i := 0; i < 4; i++ {
		r := <-ch
		if r.err != nil {
			log.Printf("dashboard: %s query failed: %v", r.name, r.err)
			c.AbortWithError(http.StatusInternalServerError, r.err)
			return
		}
		result[r.name] = r.value
	}

	// Get the current user's display name
	u, err := model.Queries.GetUserById(ctx, session.UserId)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	result["username"] = u.Userdescrip.String

	c.JSON(http.StatusOK, result)
}
