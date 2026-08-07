package api

import (
	"log"
	"net/http"
	"time"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["superbills"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", listSuperbills)
			r.POST("/", createSuperbill)
		},
	}
}

type superbillInput struct {
	Patient      int64   `json:"patient" binding:"required"`
	DateFrom     string  `json:"date_from" binding:"required"`
	DateTo       string  `json:"date_to" binding:"required"`
	Provider     int64   `json:"provider" binding:"required"`
	Status       string  `json:"status"`
	TotalCharges float64 `json:"total_charges"`
}

func listSuperbills(c *gin.Context) {
	rows, err := model.Queries.ListSuperbills(c.Request.Context())
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, rows)
}

func createSuperbill(c *gin.Context) {
	var in superbillInput
	if err := c.BindJSON(&in); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	dateFrom, err := time.Parse("2006-01-02", in.DateFrom)
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "invalid date_from format, use YYYY-MM-DD")
		return
	}
	dateTo, err := time.Parse("2006-01-02", in.DateTo)
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "invalid date_to format, use YYYY-MM-DD")
		return
	}

	sess, err := common.GetSession(c)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	result, err := model.Queries.CreateSuperbill(c.Request.Context(), dbgen.CreateSuperbillParams{
		Patient:      in.Patient,
		DateFrom:     dateFrom,
		DateTo:       dateTo,
		Provider:     in.Provider,
		Status:       in.Status,
		TotalCharges: in.TotalCharges,
		DateCreated:  time.Now(),
		User:         sess.UserId,
	})
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": newID})
}
