package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["superbill-templates"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", listSuperbillTemplates)
			r.GET("/:id", getSuperbillTemplate)
		},
	}
}

func listSuperbillTemplates(c *gin.Context) {
	rows, err := model.Queries.ListSuperbillTemplates(c.Request.Context())
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, rows)
}

func getSuperbillTemplate(c *gin.Context) {
	id := common.ParseInt(c.Param("id"))
	if id == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	row, err := model.Queries.GetSuperbillTemplate(c.Request.Context(), id)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, row)
}
