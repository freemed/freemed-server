package api

import (
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func init() {
	common.ApiMap["config"] = common.ApiMapping{
		Authenticated: true,
		JsonArmored:   true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/all", ConfigGetAll)
			//r.Get("/view", MessagesView)
		},
	}
}

func ConfigGetAll(r *gin.Context) {
	var o []model.ConfigModel
	_, err := model.DbMap.Select(&o, "SELECT * FROM "+model.TABLE_CONFIG)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, o)
}
