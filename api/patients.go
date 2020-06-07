package api

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["patients"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.POST("/searchDuplicates", patientSearchForDuplicates)
			r.POST("/search", patientSearch)
			r.GET("/picklist/:param", patientPicklist)
			r.GET("/total", patientTotalInSystem)
		},
	}
}

type patientSearchResult struct {
	LastName    string `db:"last_name" json:"last_name"`
	FirstName   string `db:"first_name" json:"first_name"`
	MiddleName  string `db:"middle_name" json:"middle_name"`
	PatientID   string `db:"patient_id" json:"patient_id"`
	Age         int64  `db:"age" json:"age"`
	DateOfBirth string `db:"date_of_birth" json:"date_of_birth"`
	ID          int64  `db:"id" json:"id"`
}

type picklistItem struct {
	Value string `db:"value" json:"value"`
	ID    int64  `db:"id" json:"id"`
}

func patientPicklist(r *gin.Context) {
	param := r.Param("param")
	if param == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	clauses := make([]string, 0)
	params := make([]interface{}, 0)

	limit := 20 // only return 20 results

	var first, last, either string
	if strings.Index(param, ",") > -1 {
		// If there's a comma ...
		parts := strings.SplitN(param, ",", 2)
		last = strings.TrimSpace(parts[0])
		first = strings.TrimSpace(parts[1])
	} else {
		if strings.Index(param, " ") > -1 {
			// Space but no comma
			parts := strings.SplitN(param, " ", 2)
			first = strings.TrimSpace(parts[0])
			last = strings.TrimSpace(parts[1])
		} else {
			either = strings.TrimSpace(param)
		}
	}

	if either != "" {
		clauses = append(clauses, "ptfname LIKE CONCAT(?, '%')")
		params = append(params, either)
		clauses = append(clauses, "ptlname LIKE CONCAT(?, '%')")
		params = append(params, either)
	}

	if first != "" && last != "" {
		clauses = append(clauses, "( ptlname LIKE CONCAT(?, '%') AND ptfname LIKE CONCAT(?, '%') )")
		params = append(params, last)
		params = append(params, first)
	} else if first != "" {
		clauses = append(clauses, "ptfname LIKE CONCAT(?, '%')")
		params = append(params, first)
		clauses = append(clauses, "ptid LIKE CONCAT(?, '%')")
		params = append(params, first)
	} else if last != "" {
		clauses = append(clauses, "ptlname LIKE CONCAT(?, '%')")
		params = append(params, last)
		clauses = append(clauses, "ptid LIKE CONCAT(?, '%')")
		params = append(params, last)
	} else {
		clauses = append(clauses, "ptid LIKE CONCAT(?, '%')")
		params = append(params, either)
	}

	params = append(params, limit)

	query := "SELECT CONCAT(ptlname, ', ', ptfname, ' (', ptid, ')') AS value, id FROM patient WHERE ( " + strings.Join(clauses, " OR ") + " ) AND ( ISNULL(ptarchive) OR ptarchive=0 ) LIMIT ?"
	var o []picklistItem
	_, err := model.DbMap.Select(&o, query, params...)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, o)
	return
}

func patientSearch(r *gin.Context) {
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
				k = append(k, "pa.city LIKE CONCAT('%%', ?, '%%')")
				v = append(v, paramValue)
			}
		case "dmv":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.dmv LIKE CONCAT('%%', ?, '%%')")
				v = append(v, paramValue)
			}
		case "email":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.pemail LIKE CONCAT('%%', ?, '%%')")
				v = append(v, paramValue)
			}
		case "first_name":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.ptfname LIKE CONCAT('%%', ?, '%%')")
				v = append(v, paramValue)
			}
		case "last_name":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.ptlname LIKE CONCAT('%%', ?, '%%')")
				v = append(v, paramValue)
			}
		case "patient_id":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.ptid LIKE CONCAT('%%', ?, '%%')")
				v = append(v, paramValue)
			}
		case "ssn":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.ssn LIKE CONCAT('%%', ?, '%%')")
				v = append(v, paramValue)
			}
		case "zip":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "pa.zip LIKE CONCAT('%%', ?, '%%')")
				v = append(v, paramValue)
			}
		default:
			break
		}
	}

	if len(v) < 1 {
		r.AbortWithError(http.StatusBadRequest, fmt.Errorf("no valid parameters presented"))
		return
	}

	// Build query
	query := fmt.Sprintf("SELECT p.ptlname AS last_name, p.ptfname AS first_name, p.ptmname AS middle_name, p.ptid AS patient_id, FLOOR( ( TO_DAYS(NOW()) - TO_DAYS(p.ptdob) ) / 365 ) AS age, p.ptdob AS date_of_birth, p.id AS id FROM "+model.TABLE_PATIENT+" p LEFT OUTER JOIN "+model.TABLE_PATIENT_ADDRESS+" pa ON p.id = pa.patient WHERE "+strings.Join(k, " AND ")+" AND pa.active = 1 "+archive+" ORDER BY p.ptlname, p.ptfname, p.ptmname LIMIT %d", limit)

	log.Printf("patientSearch(): query: %s", query)
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

func patientTotalInSystem(r *gin.Context) {
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

func patientSearchForDuplicates(r *gin.Context) {
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
