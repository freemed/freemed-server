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
	tx := model.Db.Find(&o)
	if tx.Error != nil {
		log.Print(tx.Error.Error())
		r.AbortWithError(http.StatusInternalServerError, tx.Error)
		return
	}
	r.JSON(http.StatusOK, o)
}
