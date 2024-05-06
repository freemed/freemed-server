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

	var u model.UserModel
	tx := model.Db.First(&u, session.UserId)
	if tx.Error != nil {
		log.Print(tx.Error.Error())
		r.AbortWithError(http.StatusInternalServerError, tx.Error)
		return
	}

	r.JSON(http.StatusOK, u.Description)
	return
}

func userInterfaceGetCurrentProvider(r *gin.Context) {
	session, err := common.GetSession(r)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var u model.UserModel
	tx := model.Db.First(&u, "id = ?", session.UserId)
	//providerID, err := model.DbMap.SelectInt("SELECT IFNULL(userrealphy,0) FROM user WHERE id = ?", session.UserId)
	if tx.Error != nil {
		log.Print(tx.Error.Error())
		r.AbortWithError(http.StatusInternalServerError, tx.Error)
		return
	}

	r.JSON(http.StatusOK, u.ProviderId)
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

	var us model.UserModel
	tx := model.Db.Find(&us, "username = ?", username)
	if tx.Error != nil {
		log.Print(tx.Error.Error())
		r.AbortWithError(http.StatusInternalServerError, tx.Error)
		return
	}

	r.JSON(http.StatusOK, tx.RowsAffected > 0)
	return
}
