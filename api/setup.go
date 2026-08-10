package api

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["setup"] = common.ApiMapping{
		Authenticated: false,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/status", setupStatus)
			r.POST("/initialize", setupInitialize)
		},
	}
}

// setupStatus checks whether the database has been populated.
// Returns { needs_setup: true } when the user table is empty.
func setupStatus(c *gin.Context) {
	count, err := model.Queries.GetUserCount(c.Request.Context())
	if err != nil {
		log.Printf("setupStatus: GetUserCount error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"needs_setup": count == 0,
	})
}

// setupInput is the payload for the initial system setup.
type setupInput struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	Title       string `json:"title"`
}

// setupInitialize creates the first admin user in the system.
// Refuses to run if users already exist (idempotent guard).
func setupInitialize(c *gin.Context) {
	// Guard: refuse if already initialized
	count, err := model.Queries.GetUserCount(c.Request.Context())
	if err != nil {
		log.Printf("setupInitialize: GetUserCount error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}
	if count > 0 {
		common.ErrorResponse(c, http.StatusConflict, "System has already been initialized")
		return
	}

	var in setupInput
	if err := c.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	// Hash password with bcrypt
	hashedPassword, err := model.HashPassword(in.Password)
	if err != nil {
		log.Printf("setupInitialize: HashPassword error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	// Create admin user with both password fields set
	result, err := model.Queries.CreateUser(c.Request.Context(), dbgen.CreateUserParams{
		Username:     in.Username,
		Userpassword: hashedPassword,
		Userfname:    strToNullString(in.FirstName),
		Userlname:    strToNullString(in.LastName),
		Userdescrip:  strToNullString("Initial administrator account"),
		Usertype:     sql.NullString{String: "admin", Valid: true},
	})
	if err != nil {
		log.Printf("setupInitialize: CreateUser error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}
	newID, _ := result.LastInsertId()

	// Set userpassword_bcrypt column for bcrypt auth
	if model.SqlDb != nil {
		_, err = model.SqlDb.ExecContext(
			context.Background(),
			"UPDATE user SET userpassword_bcrypt = ? WHERE id = ?",
			hashedPassword, newID,
		)
		if err != nil {
			log.Printf("setupInitialize: set userpassword_bcrypt error: %v", err)
		}
	}

	// Set admin email if provided
	if in.Email != "" {
		if model.SqlDb != nil {
			_, err = model.SqlDb.ExecContext(
				context.Background(),
				"UPDATE user SET useremail = ? WHERE id = ?",
				in.Email, newID,
			)
			if err != nil {
				log.Printf("setupInitialize: set useremail error: %v", err)
			}
		}
	}

	// Set admin title if provided
	if in.Title != "" {
		if model.SqlDb != nil {
			_, err = model.SqlDb.ExecContext(
				context.Background(),
				"UPDATE user SET usertitle = ? WHERE id = ?",
				in.Title, newID,
			)
			if err != nil {
				log.Printf("setupInitialize: set usertitle error: %v", err)
			}
		}
	}

	log.Printf("setupInitialize: created admin user %s (id=%d)", in.Username, newID)
	c.JSON(http.StatusCreated, gin.H{
		"id":       newID,
		"username": in.Username,
		"status":   "initialized",
	})
}
