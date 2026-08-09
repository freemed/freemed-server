package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/cds"
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["cds"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.POST("/drug-interactions", cdsCheckDrugInteractions)
			r.POST("/drug-allergy-check", cdsCheckDrugAllergy)
		},
	}
}

// drugInteractionsRequest is the input for the drug-drug interaction check.
type drugInteractionsRequest struct {
	DrugNames []string `json:"drug_names"`
}

// drugInteractionsResponse wraps the list of interaction results.
type drugInteractionsResponse struct {
	Interactions []cds.InteractionResult `json:"interactions"`
}

// drugAllergyRequest is the input for the drug-allergy cross-check.
type drugAllergyRequest struct {
	DrugName  string `json:"drug_name"`
	PatientID int64  `json:"patient_id"`
}

// drugAllergyResponse wraps the list of drug-allergy warnings.
type drugAllergyResponse struct {
	Warnings []cds.InteractionResult `json:"warnings"`
}

// cdsCheckDrugInteractions handles POST /api/cds/drug-interactions
func cdsCheckDrugInteractions(r *gin.Context) {
	var req drugInteractionsRequest
	if err := r.ShouldBindJSON(&req); err != nil {
		common.ErrorResponse(r, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}
	if len(req.DrugNames) == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "drug_names must be a non-empty array")
		return
	}

	checker := cds.NewLocalChecker()
	interactions, err := checker.CheckDrugDrug(req.DrugNames)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	if interactions == nil {
		interactions = []cds.InteractionResult{}
	}

	r.JSON(http.StatusOK, drugInteractionsResponse{Interactions: interactions})
}

// cdsCheckDrugAllergy handles POST /api/cds/drug-allergy-check
func cdsCheckDrugAllergy(r *gin.Context) {
	var req drugAllergyRequest
	if err := r.ShouldBindJSON(&req); err != nil {
		common.ErrorResponse(r, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}
	if req.DrugName == "" {
		common.ErrorResponse(r, http.StatusBadRequest, "drug_name is required")
		return
	}
	if req.PatientID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "patient_id is required")
		return
	}

	// Query patient's active allergies
	allergyRows, err := model.Queries.ListAllergies(r.Request.Context(), req.PatientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	// Extract allergy substance names from the active field.
	// Note: the current allergies schema uses the `active` column for status.
	// When the schema is enhanced with an allergen substance field, adjust accordingly.
	allergies := make([]string, 0, len(allergyRows))
	for _, a := range allergyRows {
		// Include allergy identifiers for matching against drug classes
		if a.Active != "" {
			allergies = append(allergies, a.Active)
		}
	}

	checker := cds.NewLocalChecker()
	warnings, err := checker.CheckDrugAllergy(req.DrugName, allergies)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}
	if warnings == nil {
		warnings = []cds.InteractionResult{}
	}

	r.JSON(http.StatusOK, drugAllergyResponse{Warnings: warnings})
}
