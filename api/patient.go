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
	common.ApiMap["patient"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.POST("/search", PatientSearch)
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
	for paramName, paramValue := range params {
		log.Printf("PatientSearch(): paramName = %s, paramValue = %v [%s]", paramName, paramValue, reflect.TypeOf(paramValue))
		switch paramName {
		case "patient_id":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.ptid LIKE '%' + ? + '%'")
				v = append(v, paramValue)
			}
		case "age":
			if iv, found := paramValue.(float64); found && iv != 0 {
				k = append(k, fmt.Sprintf("FLOOR( ( TO_DAYS(NOW()) - TO_DAYS(p.ptdob) ) / 365 ) = %d", int64(paramValue.(float64))))
				// no value appended
			}
		default:
			break
		}
	}

	// Build query
	query := fmt.Sprintf("SELECT p.ptlname AS last_name, p.ptfname AS first_name, p.ptmname AS middle_name, p.ptid AS patient_id, FLOOR( ( TO_DAYS(NOW()) - TO_DAYS(p.ptdob) ) / 365 ) AS age, p.ptdob AS date_of_birth, p.id AS id FROM "+model.TABLE_PATIENT+" p LEFT OUTER JOIN "+model.TABLE_PATIENT_ADDRESS+" pa ON p.id = pa.patient WHERE "+strings.Join(k, " AND ")+" AND pa.active = 1 ORDER BY p.ptlname, p.ptfname, p.ptmname LIMIT %d", limit)

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
