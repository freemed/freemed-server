package api

import (
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
	"github.com/martini-contrib/render"
	"log"
	"net/http"
)

func init() {
	common.ApiMap["emr/data_store"] = common.ApiMapping{
		Authenticated: true,
		JsonArmored:   false,
		RouterFunction: func(r martini.Router) {
			r.Get("/get/:patient/:module/:id", DataStoreGet)
			//r.Put("/put", DataStorePut)
		},
	}
}

func DataStoreGet(enc encoder.Encoder, r render.Render, params martini.Params, res http.ResponseWriter) {
	var ok bool
	var patient, module, id string
	if patient, ok = params["patient"]; !ok {
		log.Print("DataStoreGet(): No patient provided")
		r.JSON(http.StatusInternalServerError, false)
		return
	}
	if module, ok = params["module"]; !ok {
		log.Print("DataStoreGet(): No module provided")
		r.JSON(http.StatusInternalServerError, false)
		return
	}
	if id, ok = params["id"]; !ok {
		log.Print("DataStoreGet(): No id provided")
		r.JSON(http.StatusInternalServerError, false)
		return
	}

	// FIXME: THIS PROBABLY NEEDS TO DIRECTLY RETURN []BYTE
	var content string
	_, err := model.DbMap.SelectStr(content, "SELECT contents FROM pds WHERE patient = ? AND module = LOWER(?) AND id = ?", patient, module, id)
	if err != nil {
		log.Print(err.Error())
		r.JSON(http.StatusInternalServerError, false)
		return
	}

	res.WriteHeader(200)
	res.Write([]byte(content))
	return
}
