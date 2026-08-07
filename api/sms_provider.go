package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["sms-providers"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", listSmsProviders)
			r.POST("/send", sendSms)
		},
	}
}

func listSmsProviders(c *gin.Context) {
	rows, err := model.Queries.ListSmsProviders(c.Request.Context())
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, rows)
}

type smsSendInput struct {
	ProviderID int64  `json:"provider_id" binding:"required"`
	To         string `json:"to" binding:"required"`
	Message    string `json:"message" binding:"required"`
}

func sendSms(c *gin.Context) {
	var in smsSendInput
	if err := c.BindJSON(&in); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Verify the provider exists
	_, err := model.Queries.GetSmsProvider(c.Request.Context(), in.ProviderID)
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	// TODO: Integrate with actual SMS sending service
	// For now, return accepted
	c.JSON(http.StatusAccepted, gin.H{
		"status":  "queued",
		"to":      in.To,
		"message": in.Message,
	})
}
