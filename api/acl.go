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
	common.ApiMap["acl"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			// Groups
			r.GET("/groups", aclGroupsList)
			r.POST("/groups", aclGroupsCreate)
			r.PUT("/groups/:id", aclGroupsUpdate)
			r.DELETE("/groups/:id", aclGroupsDelete)

			// Permissions
			r.GET("/permissions", aclPermissionsList)
			r.POST("/permissions", aclPermissionsCreate)
			r.DELETE("/permissions/:id", aclPermissionsDelete)

			// Group permissions
			r.GET("/groups/:id/permissions", aclGroupPermissionsList)
			r.POST("/groups/:id/permissions", aclGroupPermissionAdd)
			r.DELETE("/groups/:id/permissions/:permId", aclGroupPermissionRemove)

			// User group assignments
			r.GET("/users/:id/groups", aclUserGroupsList)
			r.POST("/users/:id/groups", aclUserGroupAdd)
			r.DELETE("/users/:id/groups/:groupId", aclUserGroupRemove)
		},
	}
}

// ---- Groups ----

func aclGroupsList(r *gin.Context) {
	rows, err := model.Queries.ListGroups(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}

type aclGroupInput struct {
	Groupname    string `json:"groupname" binding:"required"`
	Groupdescrip string `json:"groupdescrip"`
}

func aclGroupsCreate(r *gin.Context) {
	var in aclGroupInput
	if err := r.BindJSON(&in); err != nil {
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}
	result, err := model.Queries.CreateGroup(r.Request.Context(), dbgen.CreateGroupParams{
		Groupname:    in.Groupname,
		Groupdescrip: in.Groupdescrip,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}

func aclGroupsUpdate(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}
	var in aclGroupInput
	if err := r.BindJSON(&in); err != nil {
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}
	err := model.Queries.UpdateGroup(r.Request.Context(), dbgen.UpdateGroupParams{
		ID:           id,
		Groupname:    in.Groupname,
		Groupdescrip: in.Groupdescrip,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func aclGroupsDelete(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}
	err := model.Queries.DeleteGroup(r.Request.Context(), id)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// ---- Permissions ----

func aclPermissionsList(r *gin.Context) {
	rows, err := model.Queries.ListPermissions(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}

type aclPermissionInput struct {
	PermissionName string `json:"permission_name" binding:"required"`
	PermissionDesc string `json:"permission_desc"`
}

func aclPermissionsCreate(r *gin.Context) {
	var in aclPermissionInput
	if err := r.BindJSON(&in); err != nil {
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}
	result, err := model.Queries.CreatePermission(r.Request.Context(), dbgen.CreatePermissionParams{
		PermissionName: in.PermissionName,
		PermissionDesc: in.PermissionDesc,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}

func aclPermissionsDelete(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}
	err := model.Queries.DeletePermission(r.Request.Context(), id)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// ---- Group Permissions ----

func aclGroupPermissionsList(r *gin.Context) {
	groupID := common.ParseInt(r.Param("id"))
	if groupID == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}
	rows, err := model.Queries.GetGroupPermissions(r.Request.Context(), groupID)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}

type aclGroupPermissionInput struct {
	PermissionID int64 `json:"permission_id" binding:"required"`
}

func aclGroupPermissionAdd(r *gin.Context) {
	groupID := common.ParseInt(r.Param("id"))
	if groupID == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}
	var in aclGroupPermissionInput
	if err := r.BindJSON(&in); err != nil {
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}
	err := model.Queries.AddGroupPermission(r.Request.Context(), dbgen.AddGroupPermissionParams{
		GroupID:      groupID,
		PermissionID: in.PermissionID,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, gin.H{"status": "added"})
}

func aclGroupPermissionRemove(r *gin.Context) {
	groupID := common.ParseInt(r.Param("id"))
	if groupID == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}
	permID := common.ParseInt(r.Param("permId"))
	if permID == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}
	err := model.Queries.RemoveGroupPermission(r.Request.Context(), dbgen.RemoveGroupPermissionParams{
		GroupID:      groupID,
		PermissionID: permID,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, gin.H{"status": "removed"})
}

// ---- User Group Assignments ----

func aclUserGroupsList(r *gin.Context) {
	userID := common.ParseInt(r.Param("id"))
	if userID == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}
	rows, err := model.Queries.ListUserGroups(r.Request.Context(), userID)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}

type aclUserGroupInput struct {
	GroupID int64 `json:"group_id" binding:"required"`
}

func aclUserGroupAdd(r *gin.Context) {
	userID := common.ParseInt(r.Param("id"))
	if userID == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}
	var in aclUserGroupInput
	if err := r.BindJSON(&in); err != nil {
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}
	err := model.Queries.AddUserToGroup(r.Request.Context(), dbgen.AddUserToGroupParams{
		UserID:  userID,
		GroupID: in.GroupID,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, gin.H{"status": "added"})
}

func aclUserGroupRemove(r *gin.Context) {
	userID := common.ParseInt(r.Param("id"))
	if userID == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}
	groupID := common.ParseInt(r.Param("groupId"))
	if groupID == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}
	err := model.Queries.RemoveUserFromGroup(r.Request.Context(), dbgen.RemoveUserFromGroupParams{
		UserID:  userID,
		GroupID: groupID,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, gin.H{"status": "removed"})
}
