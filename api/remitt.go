package api

import (
	"github.com/freemed/freemed-server/common"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["remitt"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			/*
				r.GET("/submit/:billkey/:render/:transport/:option", schedulerDailyApptRange)
				r.GET("/dailyapptscheduler/:date", schedulerDailyApptScheduler)
				r.GET("/dateappt/:date", schedulerFindDateAppt)
				r.GET("/event/:id", schedulerGetEvent)
				r.POST("/reschedule/:id", schedulerReschedule)
			*/
		},
	}
}
