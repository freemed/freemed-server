package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["claims"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/recent", getRecentClaims)
		},
	}
}

// getRecentClaims returns the 50 most recent claim log entries system-wide
func getRecentClaims(c *gin.Context) {
	claims, err := model.Queries.RecentClaims(c.Request.Context())
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, claims)
}

// patientClaims returns claim log entries for a specific patient
func patientClaims(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	claims, err := model.Queries.PatientClaims(c.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, claims)
}
