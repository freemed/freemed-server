package main

import (
	"bytes"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
	"github.com/martini-contrib/render"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type cszPicklistObj struct {
	Username string `json:"username" binding:"required"`
	Id       string `json:"id" binding:"required"`
}

func CityStateZipPicklist(enc encoder.Encoder, r render.Render, params martini.Params) {
	var o []ZipcodesModel
	var buf bytes.Buffer

	if _, ok := params["param"]; !ok {
		r.JSON(http.StatusInternalServerError, false)
		return
	}
	param, err := url.QueryUnescape(params["param"])
	if err != nil {
		log.Print(err.Error())
		r.JSON(http.StatusInternalServerError, false)
		return
	}
	log.Print("CityStateZipPicklist(): param = '" + param + "'")

	buf.WriteString("SELECT * FROM " + TABLE_ZIPCODES + " WHERE ")

	intval, _ := strconv.Atoi(param)
	if len(param) >= 4 && param[2:3] == " " {
		// ST CITY
		buf.WriteString("state = UPPER(?) AND city LIKE CONCAT('%', ? ,'%')")
		buf.WriteString(" LIMIT 20")
		_, err = dbmap.Select(&o, buf.String(), param[0:2], param[3:])
	} else if len(param) > 4 && param[len(param)-4:len(param)-2] == ", " {
		// CITY, ST
		buf.WriteString("state = UPPER(?) AND city LIKE CONCAT('%', ? ,'%')")
		buf.WriteString(" LIMIT 20")
		_, err = dbmap.Select(&o, buf.String(), param[len(param)-2:len(param)], param[0:len(param)-4])
	} else if len(param) > 4 && param[len(param)-3:len(param)-2] == " " {
		// CITY ST
		buf.WriteString("state = UPPER(?) AND city LIKE CONCAT('%', ? ,'%')")
		buf.WriteString(" LIMIT 20")
		_, err = dbmap.Select(&o, buf.String(), param[len(param)-2:len(param)], param[0:len(param)-3])
	} else if len(param) >= 3 && !strings.ContainsAny(param, "0123456789") {
		// CITY
		buf.WriteString("city LIKE CONCAT('%', ? ,'%')")
		buf.WriteString(" LIMIT 20")
		_, err = dbmap.Select(&o, buf.String(), param)
	} else if intval > 0 {
		// ZIP
		buf.WriteString("zip LIKE CONCAT('%', ? ,'%')")
		buf.WriteString(" LIMIT 20")
		_, err = dbmap.Select(&o, buf.String(), param)
	} else {
		// Absolutely nothing
	}

	if err != nil {
		log.Print(err.Error())
		r.JSON(http.StatusInternalServerError, false)
		return
	}

	out := make(map[string]string, len(o))
	for _, v := range o {
		out[fmt.Sprintf("%d", v.Id)] = strings.TrimSpace(v.City + ", " + v.State + " " + v.Zip + " " + v.Country)
	}

	r.JSON(http.StatusOK, out)
	return
}
