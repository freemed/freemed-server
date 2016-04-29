package api

import (
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/go-martini/martini"
	//"github.com/martini-contrib/binding"
	"github.com/martini-contrib/encoder"
	"github.com/martini-contrib/render"
	"log"
	"net/http"
)

func init() {
	common.ApiMap["userinterface"] = common.ApiMapping{
		Authenticated: true,
		JsonArmored:   true,
		RouterFunction: func(r martini.Router) {
			// GetCurrentUsername
			r.Get("/CurrentUsername", UserInterface_GetCurrentUsername)
			// GetCurrentProvider
			r.Get("/CurrentProvider", UserInterface_GetCurrentProvider)
			// CheckDuplicate
			r.Get("/CheckDuplicate/:username", UserInterface_CheckDuplicate)
			// GetUsers
			// GetEmrConfiguration
			// GetNewMessages
			// SetConfigValue
			// GetRecord
			// GetRecords
			// add
			// del
			// mod
			// GetReligions
			// GetUserTheme
			// GetUserType
		},
	}
}

func UserInterface_GetCurrentUsername(session common.SessionModel, enc encoder.Encoder, r render.Render) {
	u, err := model.DbMap.Get(model.UserModel{}, session.UserId)
	if err != nil {
		log.Print(err.Error())
		r.JSON(http.StatusInternalServerError, false)
		return
	}

	r.JSON(http.StatusOK, u.(model.UserModel).Description)
	return
}

func UserInterface_GetCurrentProvider(session common.SessionModel, enc encoder.Encoder, r render.Render) {
	providerId, err := model.DbMap.SelectInt("SELECT IFNULL(userrealphy,0) FROM user WHERE id = ?", session.UserId)
	if err != nil {
		log.Print(err.Error())
		r.JSON(http.StatusInternalServerError, false)
		return
	}

	r.JSON(http.StatusOK, providerId)
	return
}

func UserInterface_CheckDuplicate(session common.SessionModel, params martini.Params, enc encoder.Encoder, r render.Render) {
	var username string
	var ok bool
	if username, ok = params["username"]; !ok {
		r.JSON(http.StatusInternalServerError, false)
		return
	}

	c, err := model.DbMap.SelectInt("SELECT COUNT(*) FROM user WHERE username = ?", username)
	if err != nil {
		log.Print(err.Error())
		r.JSON(http.StatusInternalServerError, false)
		return
	}

	r.JSON(http.StatusOK, c > 0)
	return
}
