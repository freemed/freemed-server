package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	// NOTE: this is the ONLY registration for the "claims" ApiMap key. It used to
	// be assigned twice — here and in api/claim_export.go — and because the map
	// key is the route prefix, the second assignment silently replaced the first:
	// POST /api/claims/generate and GET /api/claims/:id/x12 were dead routes that
	// fell through to the SPA fallback (HTTP 200 + HTML). The two registrations
	// are merged into one RouterFunction below so nothing shadows anything.
	//
	// Both admin routes use the :id wildcard name (not :voucher) because gin
	// panics at startup when two patterns place different wildcard names at the
	// same path position ("':voucher' in new path ... conflicts with existing
	// wildcard ':id'").
	common.ApiMap["claims"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/recent", getRecentClaims)
			r.GET("/pending", getPendingClaims)
			// Admin-only: changing a claim's status is a billing write. It is the
			// same class as /generate and /:id/x12 below, which have always been
			// guarded; leaving this one open let any authenticated user (including
			// a non-billing account) mutate claim state. Unguarded at base revision
			// 873e1a as well - found by audit, not a regression.
			r.PUT("/:id/status", common.RequireRole("admin"), updateClaimStatus)
			// Moved here from api/claim_export.go (see NOTE above).
			r.POST("/generate", common.RequireRole("admin"), claimGenerate)
			r.GET("/:id/x12", common.RequireRole("admin"), claimX12ByVoucher)
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

// getPendingClaims returns pending claim log entries
func getPendingClaims(c *gin.Context) {
	claims, err := model.Queries.PendingClaims(c.Request.Context())
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, claims)
}

// updateClaimStatus updates a claim's status
func updateClaimStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var input struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err := model.Queries.UpdateClaimStatus(c.Request.Context(), dbgen.UpdateClaimStatusParams{
		ID:     common.ParseInt(id),
		Status: input.Status,
	})
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
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
