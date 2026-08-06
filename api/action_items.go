package api

import (
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["action-items"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", billingActionItems)
		},
	}
}

func billingActionItems(c *gin.Context) {
	items, err := model.Queries.BillingActionItems(c.Request.Context())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if items == nil {
		items = []dbgen.BillingActionItemsRow{}
	}
	c.JSON(http.StatusOK, items)
}
