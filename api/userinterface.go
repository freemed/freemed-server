package api

import (
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func init() {
	common.ApiMap["userinterface"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			// GetCurrentUsername
			r.GET("/CurrentUsername", UserInterface_GetCurrentUsername)
			// GetCurrentProvider
			r.GET("/CurrentProvider", UserInterface_GetCurrentProvider)
			// CheckDuplicate
			r.GET("/CheckDuplicate/:username", UserInterface_CheckDuplicate)
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

func UserInterface_GetCurrentUsername(r *gin.Context) {
	session, err := common.GetSession(r)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	u, err := model.DbMap.Get(model.UserModel{}, session.UserId)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, u.(model.UserModel).Description)
	return
}

func UserInterface_GetCurrentProvider(r *gin.Context) {
	session, err := common.GetSession(r)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	providerId, err := model.DbMap.SelectInt("SELECT IFNULL(userrealphy,0) FROM user WHERE id = ?", session.UserId)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, providerId)
	return
}

func UserInterface_CheckDuplicate(r *gin.Context) {
	//session, err := common.GetSession(r)
	//if err != nil {
	//	log.Print(err.Error())
	//	r.AbortWithError(http.StatusInternalServerError, err)
	//	return
	//}

	username := r.Param("username")
	if username == "" {
		r.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c, err := model.DbMap.SelectInt("SELECT COUNT(*) FROM user WHERE username = ?", username)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, c > 0)
	return
}
