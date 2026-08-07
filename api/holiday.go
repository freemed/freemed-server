package api

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["holidays"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", listHolidays)
			r.GET("/check", checkHoliday)
		},
	}
}

func listHolidays(c *gin.Context) {
	rows, err := model.Queries.ListHolidays(c.Request.Context())
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, rows)
}

func checkHoliday(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		common.ErrorResponse(c, http.StatusBadRequest, "date parameter is required")
		return
	}
	parsedDate, err := common.ParseDate(dateStr)
	if err != nil {
		log.Printf("checkHoliday: bad date: %v", err)
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	row, err := model.Queries.CheckHoliday(c.Request.Context(), parsedDate)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{"is_holiday": false, "holiday": nil})
			return
		}
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"is_holiday": true,
		"holiday":    row.HolidayName,
		"date":       row.HolidayDate.Format(time.RFC3339),
	})
}
