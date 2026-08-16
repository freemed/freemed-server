package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/config"
	"github.com/freemed/freemed-server/model"
	"github.com/freemed/freemed-server/pkg/sms"
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
	if _, err := model.Queries.GetSmsProvider(c.Request.Context(), in.ProviderID); err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	// Dispatch through the configured SMS provider (noop unless a real
	// provider is configured via config.yml sms.provider).
	sender := sms.New(sms.Config{
		Provider: config.Config.Sms.Provider,
		Settings: config.Config.Sms.Settings,
	})
	if err := sender.Send(c.Request.Context(), in.To, in.Message); err != nil {
		log.Printf("sendSms: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status":  "queued",
		"to":      in.To,
		"message": in.Message,
	})
}
