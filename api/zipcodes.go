package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
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
	var o []dbgen.Zipcode
	var err error

	param := r.Param("param")
	if param == "" {
		r.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	log.Print("CityStateZipPicklist(): param = '" + param + "'")

	intval, _ := strconv.Atoi(param)
	if len(param) >= 4 && param[2:3] == " " {
		// ST CITY
		o, err = model.Queries.CityStateZipByStateCity(r.Request.Context(), dbgen.CityStateZipByStateCityParams{
			State: param[0:2],
			City:  param[3:],
		})
	} else if len(param) > 4 && param[len(param)-4:len(param)-2] == ", " {
		// CITY, ST
		o, err = model.Queries.CityStateZipByStateCity(r.Request.Context(), dbgen.CityStateZipByStateCityParams{
			State: param[len(param)-2 : len(param)],
			City:  param[0 : len(param)-4],
		})
	} else if len(param) > 4 && param[len(param)-3:len(param)-2] == " " {
		// CITY ST
		o, err = model.Queries.CityStateZipByStateCity(r.Request.Context(), dbgen.CityStateZipByStateCityParams{
			State: param[len(param)-2 : len(param)],
			City:  param[0 : len(param)-3],
		})
	} else if len(param) >= 3 && !strings.ContainsAny(param, "0123456789") {
		// CITY
		o, err = model.Queries.CityStateZipByCity(r.Request.Context(), param)
	} else if intval > 0 {
		// ZIP
		o, err = model.Queries.CityStateZipByZip(r.Request.Context(), param)
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
		out[fmt.Sprintf("%d", v.ID)] = strings.TrimSpace(v.City + ", " + v.State + " " + v.Zip + " " + v.Country)
	}

	r.JSON(http.StatusOK, out)
}
