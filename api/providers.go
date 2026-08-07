package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["providers"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", providersList)
			r.GET("/lookup-npi", providersLookupNPI)
		},
	}
}

func providersList(r *gin.Context) {
	rows, err := model.Queries.ListProviders(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}

func providersLookupNPI(r *gin.Context) {
	npi := r.Query("npi")
	if npi == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}
	rows, err := model.Queries.LookupNPI(r.Request.Context(), npi)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}
