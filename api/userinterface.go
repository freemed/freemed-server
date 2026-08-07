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
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	u, err := model.Queries.GetUserById(r.Request.Context(), session.UserId)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, u.Userdescrip.String)
	return
}

func userInterfaceGetCurrentProvider(r *gin.Context) {
	session, err := common.GetSession(r)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	u, err := model.Queries.GetUserById(r.Request.Context(), session.UserId)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, u.Userrealphy)
	return
}

func userInterfaceCheckDuplicate(r *gin.Context) {
	//session, err := common.GetSession(r)
	//if err != nil {
	//	log.Print(err.Error())
	//	common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
	//	return
	//}

	username := r.Param("username")
	if username == "" {
		r.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	us, err := model.Queries.CheckDuplicateUsername(r.Request.Context(), username)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, len(us) > 0)
	return
}
