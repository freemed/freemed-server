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
			r.GET("/patients-to-bill", patientsToBillList)
			r.GET("/patient/:id/procedures-to-bill", proceduresToBillList)
			r.GET("/claim-info", claimInfoList)
			r.POST("/mark-billed", markBilled)
			r.GET("/rebill-list", rebillList)
			r.POST("/process-claims", processClaims)
			r.POST("/process-statement", processStatement)
		},
	}
}

// remittStatus returns the remitt configuration status
func remittStatus(c *gin.Context) {
	url, err := model.ConfigGetByKey("remitt_url")
	if err != nil {
		log.Printf("remittStatus: %s", err.Error())
		common.ErrorResponse(c, http.StatusInternalServerError, "remitt not configured")
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
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
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

// patientsToBillList returns distinct patients with unbilled procedures
func patientsToBillList(c *gin.Context) {
	rows, err := model.Queries.PatientsToBill(c.Request.Context())
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, rows)
}

// proceduresToBillList returns unbilled procedures for a specific patient
func proceduresToBillList(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.ErrorResponse(c, http.StatusBadRequest, "bad request")
		return
	}
	rows, err := model.Queries.ProceduresToBill(c.Request.Context(), common.ParseInt(id))
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, rows)
}

// claimInfoList returns all billkey entries (claim info)
func claimInfoList(c *gin.Context) {
	rows, err := model.Queries.GetClaimInfo(c.Request.Context())
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, rows)
}

type markBilledInput struct {
	IDs []int64 `json:"ids" binding:"required"`
}

// markBilled marks multiple billkey entries as billed
func markBilled(c *gin.Context) {
	var in markBilledInput
	if err := c.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}
	for _, id := range in.IDs {
		if err := model.Queries.MarkAsBilled(c.Request.Context(), id); err != nil {
			log.Printf("markBilled: id=%d: %s", id, err.Error())
			common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "count": len(in.IDs)})
}

// rebillList returns all billkey entries (rebill candidates)
func rebillList(c *gin.Context) {
	rows, err := model.Queries.GetRebillList(c.Request.Context())
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, rows)
}

// processClaims is a placeholder for external claim transport
func processClaims(c *gin.Context) {
	// TODO: Implement claim processing via external transport
	c.JSON(http.StatusOK, gin.H{
		"status":  "not_implemented",
		"message": "Claim processing transport not yet implemented",
	})
}

// processStatement is a placeholder for external statement generation
func processStatement(c *gin.Context) {
	// TODO: Implement statement generation via external transport
	c.JSON(http.StatusOK, gin.H{
		"status":  "not_implemented",
		"message": "Statement generation transport not yet implemented",
	})
}
