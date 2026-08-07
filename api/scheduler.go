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
			r.POST("/", schedulerCreateAppointment)
			r.POST("/group", schedulerCreateGroupAppointment)
			r.GET("/group/:id", schedulerFindGroupAppointments)
			r.GET("/groups", listCalGroups)
			r.GET("/groups/:id", getCalGroup)
			r.GET("/blocks", schedulerListBlockedSlots)
			r.POST("/blocks", schedulerCreateBlockedSlot)
			r.DELETE("/blocks/:id", schedulerDeleteBlockedSlot)
			r.POST("/recurring", schedulerCreateRecurringAppointments)
			r.POST("/:id/copy", schedulerCopyAppointment)
			r.POST("/next-available", schedulerNextAvailable)
			r.DELETE("/:id", schedulerCancelAppointment)
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

func schedulerCancelAppointment(c *gin.Context) {
	id := common.ParseInt(c.Param("id"))
	if id < 1 {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("bad event id"))
		return
	}

	_, err := model.SqlDb.ExecContext(c.Request.Context(),
		"UPDATE scheduler SET calstatus = 'cancelled', calmodified = ? WHERE id = ?",
		time.Now(), id)
	if err != nil {
		log.Printf("schedulerCancelAppointment(%d): ERROR: %s", id, err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, true)
}

// schedulerCreateAppointment handles POST /api/scheduler — create a new appointment.
func schedulerCreateAppointment(c *gin.Context) {
	session, err := common.GetSession(c)
	if err != nil {
		log.Printf("schedulerCreateAppointment: failed to get session: %v", err)
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	var input struct {
		Date     string `json:"date"     binding:"required"`
		Hour     int64  `json:"hour"     binding:"required"`
		Minute   int64  `json:"minute"   binding:"required"`
		Duration int64  `json:"duration" binding:"required"`
		Type     string `json:"type"`
		Provider int64  `json:"provider"`
		Patient  int64  `json:"patient"`
		Note     string `json:"note"`
	}
	if err := c.ShouldBind(&input); err != nil {
		log.Printf("schedulerCreateAppointment: bind error: %v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if input.Type == "" {
		input.Type = "patient"
	}

	calDateOf, err := common.ParseDate(input.Date)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	result, err := model.Queries.CreateAppointment(c.Request.Context(), dbgen.CreateAppointmentParams{
		Caldateof:    calDateOf,
		Calhour:      input.Hour,
		Calminute:    input.Minute,
		Calduration:  input.Duration,
		Caltype:      input.Type,
		Calphysician: input.Provider,
		Calpatient:   input.Patient,
		Calprenote:   input.Note,
		User:         session.UserId,
	})
	if err != nil {
		log.Printf("schedulerCreateAppointment: ERROR: %s", err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": newID})
}

// schedulerCopyAppointment handles POST /api/scheduler/:id/copy — copy an appointment to a new date/time.
func schedulerCopyAppointment(c *gin.Context) {
	id := common.ParseInt(c.Param("id"))
	if id < 1 {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("bad event id"))
		return
	}

	var input struct {
		Date   string `json:"date"   binding:"required"`
		Hour   int64  `json:"hour"   binding:"required"`
		Minute int64  `json:"minute" binding:"required"`
	}
	if err := c.ShouldBind(&input); err != nil {
		log.Printf("schedulerCopyAppointment: bind error: %v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	calDateOf, err := common.ParseDate(input.Date)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	result, err := model.Queries.CopyAppointment(c.Request.Context(), dbgen.CopyAppointmentParams{
		NewCaldateof: calDateOf,
		NewCalhour:   input.Hour,
		NewCalminute: input.Minute,
		SourceID:     id,
	})
	if err != nil {
		log.Printf("schedulerCopyAppointment(%d): ERROR: %s", id, err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": newID})
}

// schedulerCreateGroupAppointment handles POST /api/scheduler/group — create a group appointment.
func schedulerCreateGroupAppointment(c *gin.Context) {
	session, err := common.GetSession(c)
	if err != nil {
		log.Printf("schedulerCreateGroupAppointment: failed to get session: %v", err)
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	var input struct {
		Date         string `json:"date"           binding:"required"`
		Hour         int64  `json:"hour"           binding:"required"`
		Minute       int64  `json:"minute"         binding:"required"`
		Duration     int64  `json:"duration"       binding:"required"`
		Provider     int64  `json:"provider"       binding:"required"`
		Note         string `json:"note"`
		GroupID      int64  `json:"group_id"       binding:"required"`
		GroupMembers string `json:"group_members"`
		Attendees    string `json:"attendees"`
	}
	if err := c.ShouldBind(&input); err != nil {
		log.Printf("schedulerCreateGroupAppointment: bind error: %v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	calDateOf, err := common.ParseDate(input.Date)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var attendees sql.NullString
	if input.Attendees != "" {
		attendees = sql.NullString{String: input.Attendees, Valid: true}
	}

	result, err := model.Queries.CreateGroupAppointment(c.Request.Context(), dbgen.CreateGroupAppointmentParams{
		Caldateof:       calDateOf,
		Calhour:         input.Hour,
		Calminute:       input.Minute,
		Calduration:     input.Duration,
		Calphysician:    input.Provider,
		Calprenote:      input.Note,
		Calgroupid:      input.GroupID,
		Calgroupmembers: sql.NullString{String: input.GroupMembers, Valid: input.GroupMembers != ""},
		Calattendees:    attendees,
		User:            session.UserId,
	})
	if err != nil {
		log.Printf("schedulerCreateGroupAppointment: ERROR: %s", err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": newID})
}

// schedulerFindGroupAppointments handles GET /api/scheduler/group/:id — list group appointments.
func schedulerFindGroupAppointments(c *gin.Context) {
	id := common.ParseInt(c.Param("id"))
	if id < 1 {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("bad group id"))
		return
	}

	rows, err := model.Queries.FindGroupAppointments(c.Request.Context(), id)
	if err != nil {
		log.Printf("schedulerFindGroupAppointments(%d): ERROR: %s", id, err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, rows)
}

// schedulerCreateRecurringAppointments handles POST /api/scheduler/recurring —
// clone an appointment to multiple future dates while keeping the same time.
func schedulerCreateRecurringAppointments(c *gin.Context) {
	session, err := common.GetSession(c)
	if err != nil {
		log.Printf("schedulerCreateRecurringAppointments: failed to get session: %v", err)
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	var input struct {
		ID    int64    `json:"id"    binding:"required"`
		Dates []string `json:"dates" binding:"required"`
	}
	if err := c.ShouldBind(&input); err != nil {
		log.Printf("schedulerCreateRecurringAppointments: bind error: %v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var createdIDs []int64
	for _, dateStr := range input.Dates {
		calDateOf, err := common.ParseDate(dateStr)
		if err != nil {
			log.Printf("schedulerCreateRecurringAppointments: invalid date %q: %v", dateStr, err)
			c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid date %q: %w", dateStr, err))
			return
		}

		err = model.Queries.CreateRecurringAppointment(c.Request.Context(), dbgen.CreateRecurringAppointmentParams{
			Caldateof: calDateOf,
			User:      session.UserId,
			SourceID:  input.ID,
		})
		if err != nil {
			log.Printf("schedulerCreateRecurringAppointments(id=%d): ERROR: %s", input.ID, err.Error())
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		createdIDs = append(createdIDs, input.ID) // sqlc :exec doesn't return LastInsertId
	}

	c.JSON(http.StatusCreated, gin.H{
		"count": len(createdIDs),
	})
}
