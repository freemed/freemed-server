package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["callin"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("", callInList)
			r.GET("/:id", callInDetail)
		},
	}
}

// callInList handles GET /api/callin
func callInList(c *gin.Context) {
	callins, err := model.Queries.ListCallIns(c.Request.Context())
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, callins)
}

// callInDetail handles GET /api/callin/:id
func callInDetail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	callinID := common.ParseInt(id)
	callin, err := model.Queries.GetCallIn(c.Request.Context(), callinID)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, callin)
}
