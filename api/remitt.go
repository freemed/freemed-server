package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["remitt"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/status", remittStatus)
			r.GET("/months", remittMonths)
		},
	}
}

// remittStatus returns the remitt configuration status
func remittStatus(c *gin.Context) {
	url, err := model.ConfigGetByKey("remitt_url")
	if err != nil {
		log.Printf("remittStatus: %s", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "remitt not configured"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"configured": url.Value.Valid,
		"url":        url.Value.String,
	})
}

// remittMonths returns available billing months from billkey table
func remittMonths(c *gin.Context) {
	months, err := model.Queries.ListBillkeyMonths(c.Request.Context())
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Transform into month labels
	type monthEntry struct {
		ID    int64  `json:"id"`
		Month string `json:"month"`
		Count int64  `json:"count"`
	}

	result := make([]monthEntry, 0, len(months))
	for _, m := range months {
		result = append(result, monthEntry{
			ID:    m.ID,
			Month: m.Billkeydate.Format("2006-01"),
			Count: int64(len(m.Bkprocs)),
		})
	}

	c.JSON(http.StatusOK, result)
}
