package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/freemed/freemed-server/common"
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
		},
	}
}

func schedulerDailyApptRange(c *gin.Context) {
	pFrom, err := dateParse(c.Param("from"))
	if err != nil {
		c.Error(err)
	}
	pTo, err := dateParse(c.Param("to"))
	if err != nil {
		c.Error(err)
	}
	provider, _ := strconv.ParseInt(c.Query("provider"), 10, 64)
	vars := []interface{}{}

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
	calshr := 9       // FIXME: TODO: IMPLEMENT: XXX
	calehr := 16      // FIXME: TODO: IMPLEMENT: XXX
	calinterval := 15 // FIXME: TODO: IMPLEMENT: XXX

	dt, err := dateParse(c.Param("date"))
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
	dt, err := dateParse(c.Param("date"))
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

// dateParse accepts a date parameter and attempts to parse it properly
func dateParse(s string) (t time.Time, e error) {
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		// TODO: FIXME: IMPLEMENT: More commmon formats
	}
	if s == "" {
		return time.Now(), fmt.Errorf("Unable to parse null date")
	}
	for _, f := range formats {
		t, e = time.Parse(f, s)
		if e == nil {
			return
		}
	}
	return time.Now(), fmt.Errorf("Unable to parse date")
}

func mysqlDateFormat(t time.Time) string {
	return t.Format("2006-01-02")
}
