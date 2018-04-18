package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["config"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/all", configGetAll)
			//r.Get("/view", MessagesView)
		},
	}
}

func configGetAll(r *gin.Context) {
	var o []model.ConfigModel
	_, err := model.DbMap.Select(&o, "SELECT * FROM "+model.TABLE_CONFIG)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, o)
}
