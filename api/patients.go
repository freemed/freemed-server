package api

import (
	"fmt"
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"reflect"
	"strings"
)

func init() {
	common.ApiMap["patients"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.POST("/searchDuplicates", PatientSearchForDuplicates)
			r.POST("/search", PatientSearch)
			r.GET("/total", PatientTotalInSystem)
		},
	}
}

type patientSearchResult struct {
	LastName    string `db:"last_name" json:"last_name"`
	FirstName   string `db:"first_name" json:"first_name"`
	MiddleName  string `db:"middle_name" json:"middle_name"`
	PatientId   string `db:"patient_id" json:"patient_id"`
	Age         int64  `db:"age" json:"age"`
	DateOfBirth string `db:"date_of_birth" json:"date_of_birth"`
	Id          int64  `db:"id" json:"id"`
}

func PatientSearch(r *gin.Context) {
	var params gin.H
	if err := r.BindJSON(&params); err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}
	log.Printf("PatientSearch(): raw params = %v", params)

	if len(params) < 1 {
		log.Print("PatientSearch(): no usable search parameters found")
		r.AbortWithStatus(http.StatusBadRequest)
	}

	limit := 20

	// Break passed parameters into something usable
	k := make([]string, 0)
	v := make([]interface{}, 0)
	archive := " AND p.ptarchive = 0 "
	for paramName, paramValue := range params {
		log.Printf("PatientSearch(): paramName = %s, paramValue = %v [%s]", paramName, paramValue, reflect.TypeOf(paramValue))
		switch paramName {
		case "age":
			if iv, found := paramValue.(float64); found && iv != 0 {
				k = append(k, fmt.Sprintf("FLOOR( ( TO_DAYS(NOW()) - TO_DAYS(p.ptdob) ) / 365 ) = %d", int64(paramValue.(float64))))
				// no value appended
			}
		case "archive":
			if bv, found := paramValue.(bool); found && bv {
				// If archived patients are included...
				archive = ""
			}
		case "city":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "pa.city LIKE '%' + ? + '%'")
				v = append(v, paramValue)
			}
		case "dmv":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.dmv LIKE '%' + ? + '%'")
				v = append(v, paramValue)
			}
		case "email":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.pemail LIKE '%' + ? + '%'")
				v = append(v, paramValue)
			}
		case "first_name":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.ptfname LIKE '%' + ? + '%'")
				v = append(v, paramValue)
			}
		case "last_name":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.ptlname LIKE '%' + ? + '%'")
				v = append(v, paramValue)
			}
		case "patient_id":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.ptid LIKE '%' + ? + '%'")
				v = append(v, paramValue)
			}
		case "ssn":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.ssn LIKE '%' + ? + '%'")
				v = append(v, paramValue)
			}
		case "zip":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "pa.zip LIKE '%' + ? + '%'")
				v = append(v, paramValue)
			}
		default:
			break
		}
	}

	// Build query
	query := fmt.Sprintf("SELECT p.ptlname AS last_name, p.ptfname AS first_name, p.ptmname AS middle_name, p.ptid AS patient_id, FLOOR( ( TO_DAYS(NOW()) - TO_DAYS(p.ptdob) ) / 365 ) AS age, p.ptdob AS date_of_birth, p.id AS id FROM "+model.TABLE_PATIENT+" p LEFT OUTER JOIN "+model.TABLE_PATIENT_ADDRESS+" pa ON p.id = pa.patient WHERE "+strings.Join(k, " AND ")+" AND pa.active = 1 "+archive+" ORDER BY p.ptlname, p.ptfname, p.ptmname LIMIT %d", limit)

	var o []patientSearchResult
	_, err := model.DbMap.Select(&o, query, v...)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, o)
	return
}

func PatientTotalInSystem(r *gin.Context) {
	var o int64
	err := model.DbMap.SelectOne(&o, "SELECT COUNT(*) FROM patient WHERE ptarchive=0")
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, o)
	return
}

func PatientSearchForDuplicates(r *gin.Context) {
	var params gin.H
	if err := r.BindJSON(&params); err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}
	log.Printf("PatientSearchForDuplicates(): raw params = %v", params)

	var o model.NullString
	fill := make([]interface{}, 0)

	query := "SELECT ptid FROM patient p WHERE " +
		"ptlname = ? AND " +
		"ptfname = ? AND "
	fill = append(fill, params["ptlname"])
	fill = append(fill, params["ptfname"])

	if _, ok := params["ptmname"]; ok {
		query += " ptmname = ? AND "
		fill = append(fill, params["ptmname"])
	}
	if _, ok := params["ptsuffix"]; ok {
		query += " ptsuffix = ? AND "
		fill = append(fill, params["ptsuffix"])
	}
	if _, ok := params["ptdob"]; ok {
		query += " ptdob = ? AND "
		fill = append(fill, params["ptdob"])
	}
	query += " ptarchive = 0"

	err := model.DbMap.SelectOne(&o, query, fill...)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, o)
	return
}
