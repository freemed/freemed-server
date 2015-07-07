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
	common.ApiMap["config"] = common.ApiMapping{
		Authenticated: true,
		JsonArmored:   true,
		RouterFunction: func(r martini.Router) {
			r.Get("/all", ConfigGetAll)
			//r.Get("/view", MessagesView)
		},
	}
}

func ConfigGetAll(enc encoder.Encoder, r render.Render) {
	var o []model.ConfigModel
	_, err := model.DbMap.Select(&o, "SELECT * FROM "+model.TABLE_CONFIG)
	if err != nil {
		log.Print(err.Error())
		r.JSON(http.StatusInternalServerError, false)
		return
	}
	r.JSON(http.StatusOK, o)
	return
}
