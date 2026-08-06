package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/config"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["scheduler"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/dailyapptrange/:from/:to", schedulerDailyApptRange)
			r.GET("/dailyapptscheduler/:date", schedulerDailyApptScheduler)
			r.GET("/dateappt/:date", schedulerFindDateAppt)
			r.GET("/event/:id", schedulerGetEvent)
			r.POST("/reschedule/:id", schedulerReschedule)
		},
	}
}

func schedulerDailyApptRange(c *gin.Context) {
	pFrom, err := common.ParseDate(c.Param("from"))
	if err != nil {
		c.Error(err)
		return
	}
	pTo, err := common.ParseDate(c.Param("to"))
	if err != nil {
		c.Error(err)
		return
	}
	provider := common.ParseInt(c.Query("provider"))

	if provider > 0 {
		out, err := model.Queries.SchedulerDailyApptRangeByProvider(c.Request.Context(), dbgen.SchedulerDailyApptRangeByProviderParams{
			FromDate:   pFrom,
			ToDate:     pTo,
			ProviderID: provider,
		})
		if err != nil {
			log.Print(err.Error())
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, out)
		return
	}

	out, err := model.Queries.SchedulerDailyApptRange(c.Request.Context(), dbgen.SchedulerDailyApptRangeParams{
		FromDate: pFrom,
		ToDate:   pTo,
	})
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func schedulerDailyApptScheduler(c *gin.Context) {
	calshr := config.Config.Scheduler.Start
	calehr := config.Config.Scheduler.End
	calinterval := config.Config.Scheduler.Interval

	dt, err := common.ParseDate(c.Param("date"))
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	provider, _ := strconv.ParseInt(c.Query("provider"), 10, 64)

	out, err := model.Queries.SchedulerDailyApptScheduler(c.Request.Context(), dbgen.SchedulerDailyApptSchedulerParams{
		ReqDate:          dt,
		StartHour:        int64(calshr),
		EndHour:          int64(calehr),
		IntervalMinutes:  int64(calinterval),
		ProviderID:       provider,
	})
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func schedulerFindDateAppt(c *gin.Context) {
	dt, err := common.ParseDate(c.Param("date"))
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	provider, _ := strconv.ParseInt(c.Query("provider"), 10, 64)

	if provider > 0 {
		out, err := model.Queries.SchedulerFindDateApptByProvider(c.Request.Context(), dbgen.SchedulerFindDateApptByProviderParams{
			ReqDate:    dt,
			ProviderID: provider,
		})
		if err != nil {
			log.Print(err.Error())
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, out)
		return
	}

	out, err := model.Queries.SchedulerFindDateAppt(c.Request.Context(), dt)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func schedulerGetEvent(c *gin.Context) {
	id := common.ParseInt(c.Param("id"))
	if id == 0 {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid id presented"))
		return
	}

	out, err := model.Queries.SchedulerGetEvent(c.Request.Context(), id)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	log.Printf("%#v", out)
	c.JSON(http.StatusOK, out)
}

func schedulerReschedule(c *gin.Context) {
	id := common.ParseInt(c.Param("id"))
	if id < 1 {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("bad event id"))
		return
	}

	type rescheduleInfo struct {
		Date     model.NullString `json:"date"`
		Hour     model.NullInt64  `json:"hour"`
		Minute   model.NullInt64  `json:"minute"`
		Duration model.NullInt64  `json:"duration"`
	}
	var info rescheduleInfo
	err := c.ShouldBind(&info)
	if err != nil {
		log.Printf("schedulerReschedule(%d): ERROR: %s", id, err.Error())
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	eventObj, err := model.Queries.GetSchedulerById(c.Request.Context(), id)
	if err != nil {
		log.Printf("schedulerReschedule(%d): ERROR: %s", id, err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if info.Date.Valid {
		dt, err := common.ParseDate(info.Date.String)
		if err == nil {
			eventObj.Caldateof = dt
		}
	}
	if info.Hour.Valid {
		eventObj.Calhour = info.Hour.Int64
	}
	if info.Minute.Valid {
		eventObj.Calminute = info.Minute.Int64
	}
	if info.Duration.Valid {
		eventObj.Calduration = info.Duration.Int64
	}

	now := time.Now()

	log.Printf("schedulerReschedule(%d): %#v", id, eventObj)
	err = model.Queries.UpdateScheduler(c.Request.Context(), dbgen.UpdateSchedulerParams{
		ID:       id,
		DateOf:   eventObj.Caldateof,
		Hour:     eventObj.Calhour,
		Minute:   eventObj.Calminute,
		Duration: eventObj.Calduration,
		Modified: sql.NullTime{Time: now, Valid: true},
	})
	if err != nil {
		log.Printf("schedulerReschedule(%d): ERROR: %s", id, err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, true)
}

func mysqlDateFormat(t time.Time) string {
	return t.Format("2006-01-02")
}
