package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/gin-gonic/gin"

	"github.com/freemed/freemed-server/model"
)

func init() {
	common.ApiMap["authorizations"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/:id", getAuthorization)
		},
	}
}

func getAuthorization(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	authorization, err := model.Queries.GetAuthorization(c.Request.Context(), common.ParseInt(id))
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, authorization)
}

func patientAuthorizations(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	authorizations, err := model.Queries.ListAuthorizations(c.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, authorizations)
}
