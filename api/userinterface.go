package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["userinterface"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			// GetCurrentUsername
			r.GET("/CurrentUsername", userInterfaceGetCurrentUsername)
			// GetCurrentProvider
			r.GET("/CurrentProvider", userInterfaceGetCurrentProvider)
			// CheckDuplicate
			r.GET("/CheckDuplicate/:username", userInterfaceCheckDuplicate)
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

func userInterfaceGetCurrentUsername(r *gin.Context) {
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

func userInterfaceGetCurrentProvider(r *gin.Context) {
	session, err := common.GetSession(r)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	providerID, err := model.DbMap.SelectInt("SELECT IFNULL(userrealphy,0) FROM user WHERE id = ?", session.UserId)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, providerID)
	return
}

func userInterfaceCheckDuplicate(r *gin.Context) {
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
