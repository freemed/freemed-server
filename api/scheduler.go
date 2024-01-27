package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/config"
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
			// canBookAppointment
			// SetAppointment
			// SetGroupAppointment
			// set_recurring_appointment
			//
		},
	}
}

type schedulerItem struct {
	DateOf                time.Time        `json:"date_of" db:"date_of"`
	DateOfMDY             string           `json:"date_of_mdy" db:"date_of_mdy"`
	Hour                  int              `json:"hour" db:"hour"`
	Minute                int              `json:"minute" db:"minute"`
	AppointmentTime       string           `json:"appointment_time" db:"appointment_time"`
	Duration              int              `json:"duration" db:"duration"`
	ProviderName          model.NullString `json:"provider" db:"provider"`
	ProviderID            model.NullInt64  `json:"provider_id" db:"provider_id"`
	ResourceType          string           `json:"resource_type" db:"resource_type"`
	PatientName           string           `json:"patient" db:"patient"`
	PatientID             int64            `json:"patient_id" db:"patient_id"`
	Note                  string           `json:"note" db:"note"`
	String                model.NullString `json:"status" db:"status"`
	StatusColor           model.NullString `json:"status_color" db:"status_color"`
	SchedulerID           int64            `json:"scheduler_id" db:"scheduler_id"`
	AppointmentTemplateID model.NullInt64  `json:"appointment_template_id" db:"appointment_template_id"`
	TemplateColor         model.NullString `json:"template_color" db:"template_color"`
}

func schedulerDailyApptRange(c *gin.Context) {
	pFrom, err := common.ParseDate(c.Param("from"))
	if err != nil {
		c.Error(err)
	}
	pTo, err := common.ParseDate(c.Param("to"))
	if err != nil {
		c.Error(err)
	}
	provider := common.ParseInt(c.Query("provider"))
	vars := []interface{}{}

	query := "SELECT s.caldateof AS date_of" +
		", DATE_FORMAT(s.caldateof, '%m/%d/%Y') AS date_of_mdy" +
		", s.calhour AS hour" +
		", s.calminute AS minute" +
		", CONCAT(LPAD(s.calhour, 2, '0'),':',LPAD(s.calminute, 2, '0')) AS appointment_time" +
		", s.calduration AS duration" +
		", CONCAT(ph.phylname, ', ', ph.phyfname) AS provider" +
		", ph.id AS provider_id" +
		", s.caltype AS resource_type" +
		", CASE s.caltype WHEN 'block' THEN '-' WHEN 'temp' THEN CONCAT( '[!] ', ci.cilname, ', ', ci.cifname, ' (', ci.cicomplaint, ')' ) WHEN 'group' THEN CONCAT( cg.groupname, ' (', cg.grouplength, ' members)') ELSE CONCAT(pa.ptlname, ', ', pa.ptfname, IF(LENGTH(pa.ptmname)>0,CONCAT(' ',pa.ptmname),''), IF(LENGTH(pa.ptsuffix)>0,CONCAT(' ',pa.ptsuffix),''),IF(LENGTH(pa.ptid)>0,CONCAT(' (',pa.ptid,')'),'')) END AS patient" +
		", s.calpatient AS patient_id" +
		", s.calprenote AS note" +
		", SUBSTRING_INDEX(GROUP_CONCAT(st.sname), ',', -1) AS status" +
		", SUBSTRING_INDEX(GROUP_CONCAT(st.scolor), ',', -1) AS status_color" +
		", s.id AS scheduler_id" +
		", s.calappttemplate as appointment_template_id" +
		", aptm.atcolor as template_color" +
		" FROM scheduler s" +
		" LEFT OUTER JOIN appttemplate aptm ON s.calappttemplate=aptm.id" +
		" LEFT OUTER JOIN scheduler_status ss ON s.id=ss.csappt" +
		" LEFT OUTER JOIN schedulerstatustype st ON st.id=ss.csstatus" +
		" LEFT OUTER JOIN physician ph ON s.calphysician=ph.id " +
		" LEFT OUTER JOIN patient pa ON s.calpatient=pa.id " +
		" LEFT OUTER JOIN callin ci ON s.calpatient=ci.id " +
		" LEFT OUTER JOIN calgroup cg ON s.calpatient=cg.id " +
		" WHERE ("
	query += " s.caldateof >= ? AND s.caldateof <= ? "
	vars = append(vars, pFrom, pTo)
	query +=
		" ) " +
			" AND s.calstatus NOT IN ( 'noshow', 'cancelled' ) "
	if provider > 0 {
		query += " AND s.calphysician=? "
		vars = append(vars, provider)
	}
	query += " GROUP BY s.id, ss.csappt " +
		" ORDER BY s.caldateof, s.calhour, s.calminute, s.calphysician DESC"
	var out []schedulerItem
	_, err = model.DbMap.Select(&out, query, vars...)
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

	query := "CALL schedulerGenerateDailySchedule ( ?, ?, ?, ?, ? );"
	type schedulerItem struct {
		AppointmentTime string           `json:"appointment_time" db:"appointment_time"`
		Hour            int              `json:"hour" db:"hour"`
		Minute          int              `json:"minute" db:"minute"`
		ResourceType    string           `json:"resource_type" db:"resource_type"`
		Cont            int              `json:"cont" db:"cont"`
		SchedulerID     int              `json:"scheduler_id" db:"scheduler_id"`
		PatientID       int              `json:"patient_id" db:"patient_id"`
		Patient         string           `json:"patient" db:"patient"`
		ProviderID      int              `json:"provider_id" db:"provider_id"`
		Provider        model.NullString `json:"provider" db:"provider"`
		Note            string           `json:"note" db:"note"`
		Duration        int              `json:"duration" db:"duration"`
		Status          model.NullString `json:"status" db:"status"`
		StatusColor     model.NullString `json:"status_color" db:"status_color"`
	}
	var out []schedulerItem
	_, err = model.DbMap.Select(&out, query, mysqlDateFormat(dt), calshr, calehr, calinterval, provider)
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
	vars := []interface{}{}

	query := "SELECT * FROM scheduler WHERE " +
		" (caldateof = ? " +
		" AND calstatus != 'cancelled' "
	vars = append(vars, dt)
	if provider > 0 {
		query += "AND calphysician = ?"
		vars = append(vars, provider)
	}

	var out []model.SchedulerModel
	_, err = model.DbMap.Select(&out, query, vars...)
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

	query := "SELECT s.caldateof AS date_of" +
		", DATE_FORMAT(s.caldateof, '%m/%d/%Y') AS date_of_mdy" +
		", s.calhour AS hour" +
		", s.calminute AS minute" +
		", CONCAT(LPAD(s.calhour, 2, '0'),':',LPAD(s.calminute, 2, '0')) AS appointment_time" +
		", s.calduration AS duration" +
		", CONCAT(ph.phylname, ', ', ph.phyfname) AS provider" +
		", ph.id AS provider_id" +
		", s.caltype AS resource_type" +
		", CASE s.caltype WHEN 'block' THEN '-' WHEN 'temp' THEN CONCAT( '[!] ', ci.cilname, ', ', ci.cifname, ' (', ci.cicomplaint, ')' ) WHEN 'group' THEN CONCAT( cg.groupname, ' (', cg.grouplength, ' members)') ELSE CONCAT(pa.ptlname, ', ', pa.ptfname, IF(LENGTH(pa.ptmname)>0,CONCAT(' ',pa.ptmname),''), IF(LENGTH(pa.ptsuffix)>0,CONCAT(' ',pa.ptsuffix),''),IF(LENGTH(pa.ptid)>0,CONCAT(' (',pa.ptid,')'),'')) END AS patient" +
		", s.calpatient AS patient_id" +
		", s.calprenote AS note" +
		", SUBSTRING_INDEX(GROUP_CONCAT(st.sname), ',', -1) AS status" +
		", SUBSTRING_INDEX(GROUP_CONCAT(st.scolor), ',', -1) AS status_color" +
		", s.id AS scheduler_id" +
		", s.calappttemplate as appointment_template_id" +
		", aptm.atcolor as template_color" +
		" FROM scheduler s" +
		" LEFT OUTER JOIN appttemplate aptm ON s.calappttemplate=aptm.id" +
		" LEFT OUTER JOIN scheduler_status ss ON s.id=ss.csappt" +
		" LEFT OUTER JOIN schedulerstatustype st ON st.id=ss.csstatus" +
		" LEFT OUTER JOIN physician ph ON s.calphysician=ph.id " +
		" LEFT OUTER JOIN patient pa ON s.calpatient=pa.id " +
		" LEFT OUTER JOIN callin ci ON s.calpatient=ci.id " +
		" LEFT OUTER JOIN calgroup cg ON s.calpatient=cg.id " +
		" WHERE ( s.id = ? ) "
	var out schedulerItem
	err := model.DbMap.SelectOne(&out, query, id)
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
	}

	// rescheduleInfo carries all of the possible fields to update, with null as default
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

	var eventObj model.SchedulerModel
	err = model.DbMap.SelectOne(&eventObj, "SELECT * FROM "+model.TABLE_SCHEDULER+" WHERE id = ?", id)
	if err != nil {
		log.Printf("schedulerReschedule(%d): ERROR: %s", id, err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Adjust if we're supposed to adjust...
	if info.Date.Valid {
		dt, err := common.ParseDate(info.Date.String)
		if err != nil {
			eventObj.Date = dt
		}
	}
	if info.Hour.Valid {
		eventObj.Hour = int(info.Hour.Int64)
	}
	if info.Minute.Valid {
		eventObj.Minute = int(info.Minute.Int64)
	}
	if info.Duration.Valid {
		eventObj.Duration = int(info.Duration.Int64)
	}

	// Adjust modification stamp
	eventObj.Modified.Time = time.Now()

	// ... and we store it
	log.Printf("schedulerReschedule(%d): %#v", id, eventObj)
	_, err = model.DbMap.Update(&eventObj)
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
