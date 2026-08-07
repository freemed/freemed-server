package api

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["billkey"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", common.RequireRole("admin"), billkeyList)
			r.POST("/", common.RequireRole("admin"), billkeyCreate)
		},
	}
}

type billkeyInput struct {
	Date       string `json:"date" binding:"required"`
	Key        string `json:"key"`
	Procedures string `json:"procedures" binding:"required"`
}

// billkeyList handles GET /api/billkey
func billkeyList(r *gin.Context) {
	rows, err := model.Queries.GetClaimInfo(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}

// billkeyCreate handles POST /api/billkey
func billkeyCreate(r *gin.Context) {
	var in billkeyInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	date, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		common.ErrorResponse(r, http.StatusBadRequest, "invalid date format, expected YYYY-MM-DD")
		return
	}

	var key sql.NullString
	if in.Key != "" {
		key = sql.NullString{String: in.Key, Valid: true}
	}

	result, err := model.Queries.InsertBillkey(r.Request.Context(), dbgen.InsertBillkeyParams{
		Billkeydate: date,
		Billkey:     key,
		Bkprocs:     in.Procedures,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}
