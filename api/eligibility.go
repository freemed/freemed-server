package api

import (
	"io"
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/freemed/freemed-server/pkg/billing"
	x12_270 "github.com/freemed/freemed-server/pkg/x12/270"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["eligibility"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.POST("/generate", eligibilityGenerate)
			r.POST("/generate/:id", eligibilityGenerateForPatient)
			r.POST("/parse", eligibilityParse271)
		},
	}
}

// eligibilityGenerate handles POST /api/eligibility/generate
// Accepts a JSON body with patient_id and generates an X12 270 inquiry.
func eligibilityGenerate(c *gin.Context) {
	var in struct {
		PatientID int64 `json:"patient_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "patient_id is required")
		return
	}

	_, err := common.GetSession(c)
	if err != nil {
		common.ErrorResponseFromError(c, http.StatusUnauthorized, err)
		return
	}

	inq, err := billing.AssembleEligibilityRequest(model.SqlDb, in.PatientID)
	if err != nil {
		log.Printf("eligibilityGenerate: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	raw270, err := x12_270.Encode270(inq)
	if err != nil {
		log.Printf("eligibilityGenerate: encode failed: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"patient_id": in.PatientID,
		"x12_270":   string(raw270),
	})
}

// eligibilityGenerateForPatient handles POST /api/eligibility/generate/:id
// Convenience endpoint that takes patient ID from URL.
func eligibilityGenerateForPatient(c *gin.Context) {
	patientID := common.ParseInt(c.Param("id"))
	if patientID == 0 {
		common.ErrorResponse(c, http.StatusBadRequest, "invalid patient ID")
		return
	}

	_, err := common.GetSession(c)
	if err != nil {
		common.ErrorResponseFromError(c, http.StatusUnauthorized, err)
		return
	}

	inq, err := billing.AssembleEligibilityRequest(model.SqlDb, patientID)
	if err != nil {
		log.Printf("eligibilityGenerateForPatient: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	raw270, err := x12_270.Encode270(inq)
	if err != nil {
		log.Printf("eligibilityGenerateForPatient: encode failed: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "ok",
		"patient_id": patientID,
		"x12_270":    string(raw270),
	})
}

// eligibilityParse271 handles POST /api/eligibility/parse
// Accepts raw X12 271 file upload and returns parsed eligibility data.
func eligibilityParse271(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		log.Printf("eligibilityParse271: failed to read file: %v", err)
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	parsed, err := x12_270.Parse271(data)
	if err != nil {
		log.Printf("eligibilityParse271: parse failed: %v", err)
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":          "ok",
		"payer_name":      parsed.PayerName,
		"payer_id":        parsed.PayerID,
		"trace_number":    parsed.TraceNumber,
		"response_date":   parsed.ResponseDate,
		"active_coverage": parsed.Subscriber.ActiveCoverage,
		"subscriber":      parsed.Subscriber,
		"dependent":       parsed.Dependent,
		"benefits":        parsed.Benefits,
	})
}
