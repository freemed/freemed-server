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
	common.ApiMap["patient"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/:id/info", patientInformation)
			r.GET("/:id/attachments", patientEmrAttachments)
			r.GET("/:id/attachments/:module", patientEmrAttachments)
		},
	}
}

type patientEmrAttachmentsResult struct {
	Patient     int64            `db:"patient" json:"patient_id"`
	Module      string           `db:"module" json:"module"`
	Oid         int64            `db:"oid" json:"module_id"`
	Annotation  model.NullString `db:"annotation" json:"annotation"`
	Summary     model.NullString `db:"summary" json:"summary"`
	Stamp       time.Time        `db:"stamp" json:"timestamp"`
	DateMdy     string           `db:"date_mdy" json:"date_mdy"`
	ModuleName  string           `db:"type" json:"module_name"`
	ModuleClass string           `db:"module_namespace" json:"module_namespace"`
	Locked      int              `db:"locked" json:"locked"`
	ID          int              `db:"id" json:"internal_id"`
}

func patientEmrAttachments(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var query string
	var err error
	var o []patientEmrAttachmentsResult

	module := r.Param("module")
	if module == "" {
		query = "SELECT p.patient AS patient, p.module AS module, p.oid AS oid, p.annotation AS annotation, p.summary AS summary, p.stamp AS stamp, DATE_FORMAT(p.stamp, '%m/%d/%Y') AS date_mdy, m.module_name AS type, m.module_class AS module_namespace, p.locked AS locked, p.id AS id FROM patient_emr p LEFT OUTER JOIN modules m ON m.module_table = p.module WHERE p.patient = ? AND m.module_hidden = 0"
		_, err = model.DbMap.Select(&o, query, id)
	} else {
		query = "SELECT p.patient AS patient, p.module AS module, p.oid AS oid, p.annotation AS annotation, p.summary AS summary, p.stamp AS stamp, DATE_FORMAT(p.stamp, '%m/%d/%Y') AS date_mdy, m.module_name AS type, m.module_class AS module_namespace, p.locked AS locked, p.id AS id FROM patient_emr p LEFT OUTER JOIN modules m ON m.module_table = p.module WHERE p.patient = ? AND p.module = ? AND m.module_hidden = 0"
		_, err = model.DbMap.Select(&o, query, id, module)
	}

	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, o)
	return
}

type patientInformationResult struct {
	Name           string `db:"patient_name" json:"patient_name"`
	ID             string `db:"patient_id" json:"patient_id"`
	DateOfBirth    string `db:"date_of_birth" json:"date_of_birth"`
	Language       string `db:"language" json:"language"`
	DateOfBirthMDY string `db:"date_of_birth_mdy" json:"date_of_birth_mdy"`
	Age            string `db:"age" json:"age"`
	Address1       string `db:"address_line_1" json:"address_line_1"`
	Address2       string `db:"address_line_2" json:"address_line_2"`
	HasAllergy     bool   `db:"hasallergy" json:"hasallergy"`
	City           string `db:"city" json:"city"`
	State          string `db:"state" json:"state"`
	Postal         string `db:"postal" json:"postal"`
	Csz            string `db:"csz" json:"csz"`
	//model.PatientModel
	Pcp      model.NullString `db:"pcp" json:"pcp"`
	Facility model.NullString `db:"facility" json:"facility"`
	Pharmacy model.NullString `db:"pharmacy" json:"pharmacy"`
}

func patientInformation(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}
	query := "SELECT " +
		"CONCAT( p.ptlname, ', ', p.ptfname, IF(NOT ISNULL(p.ptmname), CONCAT(' ', p.ptmname), '') ) AS patient_name" +
		", p.ptid AS patient_id" +
		", p.ptdob AS date_of_birth" +
		", p.ptprimarylanguage AS language" +
		", DATE_FORMAT(p.ptdob, '%m/%d/%Y') AS date_of_birth_mdy" +
		", CASE WHEN ( ( TO_DAYS(NOW()) - TO_DAYS(p.ptdob) ) / 365) >= 2 THEN CONCAT(FLOOR( ( TO_DAYS(NOW()) - TO_DAYS(p.ptdob) ) / 365),' years') ELSE CONCAT(FLOOR( ( TO_DAYS(NOW()) - TO_DAYS(p.ptdob) ) / 30),' months') END AS age" +
		", pa.line1 AS address_line_1" +
		", pa.line2 AS address_line_2" +
		", pa.city AS city" +
		", pa.stpr AS state" +
		", pa.postal AS postal" +
		", CONCAT( pa.city, ', ', pa.stpr, ' ', pa.postal ) AS csz" +
		", CASE WHEN p.id IN ( SELECT al.patient FROM allergies al WHERE al.patient=? AND active = 'active' ) THEN 'true' ELSE 'false' END AS hasallergy" +
		//", p.* " +
		", CONCAT( phy.phylname, ', ', phy.phyfname, ' ', phy.phymname ) AS pcp" +
		", CONCAT( fac.psrname, ' (', fac.psrcity, ', ', fac.psrstate,')' ) AS facility" +
		", CONCAT( ph.phname, ' (', ph.phcity, ', ', ph.phstate,')' ) AS pharmacy " +
		"FROM patient p " +
		"LEFT OUTER JOIN patient_address pa ON ( pa.patient = p.id AND pa.active = TRUE ) " +
		"LEFT OUTER JOIN physician phy ON ( phy.id = p.ptpcp) " +
		"LEFT OUTER JOIN facility fac ON ( fac.id = p.ptprimaryfacility) " +
		"LEFT OUTER JOIN pharmacy ph ON ( ph.id = p.ptpharmacy) " +
		"WHERE p.id = ? GROUP BY p.id"

	var o patientInformationResult
	err := model.DbMap.SelectOne(&o, query, id, id)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, o)
	return
}
