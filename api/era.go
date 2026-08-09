package api

import (
	"io"
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/freemed/freemed-server/pkg/era"
	x12_835 "github.com/freemed/freemed-server/pkg/x12/835"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["era"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.POST("/upload", common.RequireRole("admin"), eraUpload)
		},
	}
}

// eraUpload handles POST /api/era/upload — multipart file upload of X12 835 ERA file.
func eraUpload(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		log.Printf("eraUpload: failed to read file: %v", err)
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	// Parse the 835 file
	parsedERA, err := x12_835.Parse835(data)
	if err != nil {
		log.Printf("eraUpload: failed to parse 835: %v", err)
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	// Get user ID from session
	sess, err := common.GetSession(c)
	if err != nil {
		common.ErrorResponseFromError(c, http.StatusUnauthorized, err)
		return
	}

	// Process and post to ledger
	result, err := era.Process835ERA(model.SqlDb, model.Queries, parsedERA, sess.UserId)
	if err != nil {
		log.Printf("eraUpload: failed to process ERA: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"claims_processed":  result.ClaimsProcessed,
		"payments_posted":   result.PaymentsPosted,
		"adjustments":       result.Adjustments,
		"errors":            result.Errors,
		"payer":             parsedERA.PayerName,
		"payment_amount":    parsedERA.PaymentAmount,
		"trace_number":      parsedERA.TraceNumber,
	})
}
