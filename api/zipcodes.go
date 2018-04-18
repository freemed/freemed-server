package api

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["zipcodes"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/picklist/:param", cityStateZipPicklist)
		},
	}
}

type cszPicklistObj struct {
	Username string `json:"username" binding:"required"`
	ID       string `json:"id" binding:"required"`
}

func cityStateZipPicklist(r *gin.Context) {
	var o []model.ZipcodesModel
	var buf bytes.Buffer
	var err error

	param := r.Param("param")
	if param == "" {
		r.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	log.Print("CityStateZipPicklist(): param = '" + param + "'")

	buf.WriteString("SELECT * FROM " + model.TABLE_ZIPCODES + " WHERE ")

	intval, _ := strconv.Atoi(param)
	if len(param) >= 4 && param[2:3] == " " {
		// ST CITY
		buf.WriteString("state = UPPER(?) AND city LIKE CONCAT('%', ? ,'%')")
		buf.WriteString(" LIMIT 20")
		_, err = model.DbMap.Select(&o, buf.String(), param[0:2], param[3:])
	} else if len(param) > 4 && param[len(param)-4:len(param)-2] == ", " {
		// CITY, ST
		buf.WriteString("state = UPPER(?) AND city LIKE CONCAT('%', ? ,'%')")
		buf.WriteString(" LIMIT 20")
		_, err = model.DbMap.Select(&o, buf.String(), param[len(param)-2:len(param)], param[0:len(param)-4])
	} else if len(param) > 4 && param[len(param)-3:len(param)-2] == " " {
		// CITY ST
		buf.WriteString("state = UPPER(?) AND city LIKE CONCAT('%', ? ,'%')")
		buf.WriteString(" LIMIT 20")
		_, err = model.DbMap.Select(&o, buf.String(), param[len(param)-2:len(param)], param[0:len(param)-3])
	} else if len(param) >= 3 && !strings.ContainsAny(param, "0123456789") {
		// CITY
		buf.WriteString("city LIKE CONCAT('%', ? ,'%')")
		buf.WriteString(" LIMIT 20")
		_, err = model.DbMap.Select(&o, buf.String(), param)
	} else if intval > 0 {
		// ZIP
		buf.WriteString("zip LIKE CONCAT('%', ? ,'%')")
		buf.WriteString(" LIMIT 20")
		_, err = model.DbMap.Select(&o, buf.String(), param)
	} else {
		// Absolutely nothing
	}

	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	out := make(map[string]string, len(o))
	for _, v := range o {
		out[fmt.Sprintf("%d", v.Id)] = strings.TrimSpace(v.City + ", " + v.State + " " + v.Zip + " " + v.Country)
	}

	r.JSON(http.StatusOK, out)
	return
}
