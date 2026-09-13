package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/config"
	"github.com/freemed/freemed-server/model"
	"github.com/freemed/freemed-server/pkg/ncpdp"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["erx"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.POST("/new", erxNewPrescription)
			r.GET("/status/:id", erxPrescriptionStatus)
		},
	}
}

// erxNewPrescription handles POST /api/erx/new
// Generates a NCPDP SCRIPT NewRx message for a prescription and returns the XML.
//
// Body: { "prescription_id": 123 }
// Returns: NCPDP SCRIPT XML for transmission to pharmacy/Surescripts gateway.
func erxNewPrescription(c *gin.Context) {
	var in struct {
		PrescriptionID int64 `json:"prescription_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "prescription_id is required")
		return
	}

	_, err := common.GetSession(c)
	if err != nil {
		common.ErrorResponseFromError(c, http.StatusUnauthorized, err)
		return
	}

	// The NCPDP sender ID is assigned per-organization by NCPDP; without it the
	// resulting NewRx would carry a fabricated sender identity that any pharmacy
	// would reject. Refuse rather than emit a misleading message.
	senderNCPDPID := config.Config.Ncpdp.SenderID
	if senderNCPDPID == "" {
		common.ErrorResponse(c, http.StatusServiceUnavailable,
			"e-prescribing is not configured: set ncpdp.sender-id in config.yml or FREEMED_NCPDP_SENDER_ID")
		return
	}

	// Assemble prescription data from DB
	input, err := ncpdp.AssemblePrescriptionInput(model.SqlDb, in.PrescriptionID, senderNCPDPID)
	if err != nil {
		log.Printf("erxNewPrescription: assemble failed: %v", err)
		common.ErrorResponseFromError(c, http.StatusNotFound, err)
		return
	}

	// Build NewRx struct
	rx := ncpdp.AssembleNewRx(input)

	// Encode to NCPDP SCRIPT XML
	xmlData, err := ncpdp.EncodeNewRx(rx, input.SenderNCPDPID, input.PharmacyNCPDPID)
	if err != nil {
		log.Printf("erxNewPrescription: encode failed: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	// Update prescription status to indicate eRx generated
	_, err = model.SqlDb.ExecContext(c.Request.Context(),
		"UPDATE prescriptions SET status = 'erx_generated', updated_at = NOW() WHERE id = ?",
		in.PrescriptionID)
	if err != nil {
		log.Printf("erxNewPrescription: failed to update status: %v", err)
		// Non-fatal — still return the XML
	}

	c.JSON(http.StatusOK, gin.H{
		"status":          "ok",
		"prescription_id": in.PrescriptionID,
		"drug_name":       input.DrugName,
		"patient_name":    input.PatientFName + " " + input.PatientLName,
		"pharmacy_name":   input.PharmacyName,
		"pharmacy_ncpdp_id": input.PharmacyNCPDPID,
		"prescriber_npi":  input.PrescriberNPI,
		"ncpdp_xml":       string(xmlData),
		"warning":         "NCPDP SCRIPT message generated. Pharmacy transmission requires Surescripts certification or gateway integration.",
	})
}

// erxPrescriptionStatus handles GET /api/erx/status/:id
// Returns the e-prescribing status for a prescription.
func erxPrescriptionStatus(c *gin.Context) {
	prescriptionID := common.ParseInt(c.Param("id"))
	if prescriptionID == 0 {
		common.ErrorResponse(c, http.StatusBadRequest, "invalid prescription ID")
		return
	}

	var status, drugName string
	err := model.SqlDb.QueryRowContext(c.Request.Context(),
		"SELECT status, drug_name FROM prescriptions WHERE id = ? AND deleted_at IS NULL",
		prescriptionID).Scan(&status, &drugName)
	if err != nil {
		common.ErrorResponseFromError(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"prescription_id": prescriptionID,
		"drug_name":       drugName,
		"status":          status,
	})
}
