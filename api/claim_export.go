package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/freemed/freemed-server/pkg/billing"
	"github.com/gin-gonic/gin"
)

// NOTE: claimGenerate and claimX12ByVoucher are registered under the "claims"
// ApiMap key in api/claims.go. They deliberately have no init() here: assigning
// ApiMap["claims"] a second time would replace that registration (the map key IS
// the route prefix) and silently drop the routes declared there.

type claimGenerateInput struct {
	ProcedureIDs []int64 `json:"procedure_ids" binding:"required,min=1"`
	FacilityID   int64   `json:"facility_id" binding:"required"`
}

// claimGenerate handles POST /api/claims/generate
// Accepts procedure IDs and a facility ID, assembles and encodes an X12 837P claim.
func claimGenerate(c *gin.Context) {
	var in claimGenerateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	x12String, err := billing.AssembleAndEncodeProfessionalClaim(
		model.SqlDb,
		in.ProcedureIDs,
		in.FacilityID,
	)
	if err != nil {
		log.Printf("claimGenerate: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"x12_837": x12String,
	})
}

// claimX12ByVoucher handles GET /api/claims/:id/x12
// Generates an X12 837P for all procedures sharing a claim voucher,
// and returns the X12 text as a downloadable file.
// The path parameter is ":id" (registered in api/claims.go) because gin cannot
// have two different wildcard names at the same position as /:id/status.
func claimX12ByVoucher(c *gin.Context) {
	voucher := c.Param("id")
	if voucher == "" {
		common.ErrorResponse(c, http.StatusBadRequest, "voucher parameter is required")
		return
	}

	// Query procrec for all procedures with this voucher
	rows, err := model.SqlDb.Query(
		`SELECT id FROM procrec WHERE procvoucher = ?`, voucher)
	if err != nil {
		log.Printf("claimX12ByVoucher query: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	var procedureIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			log.Printf("claimX12ByVoucher scan: %v", err)
			common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
			return
		}
		procedureIDs = append(procedureIDs, id)
	}
	if err := rows.Err(); err != nil {
		log.Printf("claimX12ByVoucher rows: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	if len(procedureIDs) == 0 {
		common.ErrorResponse(c, http.StatusNotFound, "no procedures found for voucher")
		return
	}

	// Use the facility from the first procedure (or default to facility 1)
	facilityID := int64(1) // default
	// Try to get facility from patient's primary facility
	err = model.SqlDb.QueryRow(
		`SELECT ptprimaryfacility FROM patient p
		 JOIN procrec pr ON pr.procpatient = p.id
		 WHERE pr.procvoucher = ? LIMIT 1`, voucher).Scan(&facilityID)
	if err != nil {
		// Use default facility if lookup fails
		log.Printf("claimX12ByVoucher facility lookup: %v (using default)", err)
	}

	x12String, err := billing.AssembleAndEncodeProfessionalClaim(
		model.SqlDb,
		procedureIDs,
		facilityID,
	)
	if err != nil {
		log.Printf("claimX12ByVoucher assemble: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=\""+voucher+".x12\"")
	c.String(http.StatusOK, x12String)
}
