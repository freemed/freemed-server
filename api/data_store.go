package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["emr/data_store"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/get/:patient/:module/:id", dataStoreGet)
			//r.PUT("/put", dataStorePut)
		},
	}
}

func dataStoreGet(r *gin.Context) {
	var patient, module, id string
	patient = r.Param("patient")
	if patient == "" {
		log.Print("DataStoreGet(): No patient provided")
		r.JSON(http.StatusInternalServerError, false)
		return
	}
	module = r.Param("module")
	if module == "" {
		log.Print("DataStoreGet(): No module provided")
		r.JSON(http.StatusInternalServerError, false)
		return
	}
	id = r.Param("id")
	if id == "" {
		log.Print("DataStoreGet(): No id provided")
		r.JSON(http.StatusInternalServerError, false)
		return
	}

	var content []byte
	err := model.DbMap.SelectOne(&content, "SELECT contents FROM pds WHERE patient = ? AND module = LOWER(?) AND id = ?", patient, module, id)
	if err != nil {
		log.Print(err.Error())
		r.JSON(http.StatusInternalServerError, false)
		return
	}

	// TODO: FIXME: Need to properly determine mimetype
	r.Data(http.StatusOK, "application/x-binary", content)
	return
}
