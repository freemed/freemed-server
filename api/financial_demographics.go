package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

// financialDemographicsInput is the JSON payload for creating a financial demographics record.
type financialDemographicsInput struct {
	Income          int32   `json:"income"`
	IDType          string  `json:"id_type"`
	IDIssuer        string  `json:"id_issuer"`
	IDNumber        string  `json:"id_number"`
	IDExpire        string  `json:"id_expire"`
	HouseholdSize   int32   `json:"household_size"`
	Spouse          int32   `json:"spouse"`
	Children        int32   `json:"children"`
	OtherDependents int32   `json:"other_dependents"`
	FreeText        *string `json:"free_text"`
	EntryDesc       string  `json:"entry_desc"`
}

// patientFinancialDemographicsList handles GET /api/patient/:id/financial
func patientFinancialDemographicsList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	patientID := common.ParseInt(id)
	rows, err := model.Queries.ListFinancialDemographicsByPatient(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, rows)
}

// patientFinancialDemographicsCreate handles POST /api/patient/:id/financial
func patientFinancialDemographicsCreate(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	session, err := common.GetSession(r)
	if err != nil {
		log.Printf("patientFinancialDemographicsCreate: failed to get session: %v", err)
		common.ErrorResponseFromError(r, http.StatusUnauthorized, err)
		return
	}

	var in financialDemographicsInput
	if err := r.BindJSON(&in); err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	result, err := model.Queries.CreateFinancialDemographics(r.Request.Context(), dbgen.CreateFinancialDemographicsParams{
		PatientID:       patientID,
		Income:          in.Income,
		IDType:          in.IDType,
		IDIssuer:        in.IDIssuer,
		IDNumber:        in.IDNumber,
		IDExpire:        in.IDExpire,
		HouseholdSize:   in.HouseholdSize,
		Spouse:          in.Spouse,
		Children:        in.Children,
		OtherDependents: in.OtherDependents,
		FreeText:        toNullString(in.FreeText),
		EntryDesc:       in.EntryDesc,
		UserID:          session.UserId,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}
