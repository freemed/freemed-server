package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["users"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", common.RequireRole("admin"), usersList)
			r.POST("/", common.RequireRole("admin"), usersCreate)
			r.PUT("/:id/password", common.RequireRole("admin"), usersPasswordChange)
			r.PUT("/:id", common.RequireRole("admin"), usersUpdate)
			r.DELETE("/:id", common.RequireRole("admin"), usersDelete)
		},
	}
}

type userInput struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Description string `json:"description"`
	UserType    string `json:"user_type"`
}

// strToNullString converts an empty string to sql.NullString{Valid: false}.
func strToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func usersList(r *gin.Context) {
	rows, err := model.Queries.ListUsers(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}

func usersCreate(r *gin.Context) {
	var in userInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	// Hash password with bcrypt
	hashed, err := model.HashPassword(in.Password)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	result, err := model.Queries.CreateUser(r.Request.Context(), dbgen.CreateUserParams{
		Username:     in.Username,
		Userpassword: hashed,
		Userfname:    strToNullString(in.FirstName),
		Userlname:    strToNullString(in.LastName),
		Userdescrip:  strToNullString(in.Description),
		Usertype:     strToNullString(in.UserType),
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}

func usersUpdate(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var in userInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	err := model.Queries.UpdateUser(r.Request.Context(), dbgen.UpdateUserParams{
		ID:          id,
		Username:    in.Username,
		Userfname:   strToNullString(in.FirstName),
		Userlname:   strToNullString(in.LastName),
		Userdescrip: strToNullString(in.Description),
		Usertype:    strToNullString(in.UserType),
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func usersPasswordChange(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var in struct {
		Password string `json:"password" binding:"required"`
	}
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	// Hash password with bcrypt
	hashed, err := model.HashPassword(in.Password)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	err = model.Queries.UpdateUserPassword(r.Request.Context(), dbgen.UpdateUserPasswordParams{
		ID:           id,
		Userpassword: hashed,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{"status": "password_changed"})
}

func usersDelete(r *gin.Context) {
	id := common.ParseInt(r.Param("id"))
	if id == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err := model.Queries.DeleteUser(r.Request.Context(), id)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
