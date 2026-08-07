package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["user-groups"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", common.RequireRole("admin"), userGroupsList)
			r.POST("/", common.RequireRole("admin"), userGroupsAdd)
			r.DELETE("/", common.RequireRole("admin"), userGroupsRemove)
		},
	}
}

func userGroupsList(r *gin.Context) {
	userID := common.ParseInt(r.Query("user"))
	if userID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "missing or invalid user id")
		return
	}
	rows, err := model.Queries.ListUserGroups(r.Request.Context(), userID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}

type userGroupsInput struct {
	UserID  int64 `json:"user_id" binding:"required"`
	GroupID int64 `json:"group_id" binding:"required"`
}

func userGroupsAdd(r *gin.Context) {
	var in userGroupsInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}
	err := model.Queries.AddUserToGroup(r.Request.Context(), dbgen.AddUserToGroupParams{
		UserID:  in.UserID,
		GroupID: in.GroupID,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, gin.H{"status": "added"})
}

func userGroupsRemove(r *gin.Context) {
	userID := common.ParseInt(r.Query("user"))
	groupID := common.ParseInt(r.Query("group"))
	if userID == 0 || groupID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "missing or invalid user id or group id")
		return
	}
	err := model.Queries.RemoveUserFromGroup(r.Request.Context(), dbgen.RemoveUserFromGroupParams{
		UserID:  userID,
		GroupID: groupID,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, gin.H{"status": "removed"})
}
