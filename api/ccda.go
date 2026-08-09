package api

import (
	"io"
	"log"
	"net/http"

	"github.com/freemed/freemed-server/ccda"
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["ccda"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/patient/:id/ccd", ccdaExportCCD)
			r.POST("/import", ccdaImportCCD)
		},
	}
}

// ccdaExportCCD handles GET /api/ccda/patient/:id/ccd
// Exports a Continuity of Care Document (CCD) in C-CDA XML format.
func ccdaExportCCD(c *gin.Context) {
	idStr := c.Param("id")
	patientID := common.ParseInt(idStr)
	if patientID == 0 {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid patient ID")
		return
	}

	xmlData, err := ccda.GenerateCCD(patientID)
	if err != nil {
		log.Printf("ccdaExportCCD: error generating CCD for patient %d: %v", patientID, err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=ccd_patient_"+idStr+".xml")
	c.Data(http.StatusOK, "application/xml; charset=utf-8", xmlData)
}

// ccdaImportCCD handles POST /api/ccda/import
// Accepts a multipart file upload of a C-CDA XML document and imports
// extracted clinical data into the target patient's record.
func ccdaImportCCD(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "file is required (multipart form field 'file')")
		return
	}
	defer file.Close()

	// Parse patient_id from form
	patientIDStr := c.PostForm("patient_id")
	patientID := common.ParseInt(patientIDStr)
	if patientID == 0 {
		common.ErrorResponse(c, http.StatusBadRequest, "patient_id is required")
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		log.Printf("ccdaImportCCD: failed to read file: %v", err)
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	// Get user from session
	sess, err := common.GetSession(c)
	if err != nil {
		common.ErrorResponseFromError(c, http.StatusUnauthorized, err)
		return
	}

	// Parse the CCD XML
	parsed, err := ccda.ParseCCD(data)
	if err != nil {
		log.Printf("ccdaImportCCD: parse failed: %v", err)
		common.ErrorResponse(c, http.StatusBadRequest, "Failed to parse C-CDA document: "+err.Error())
		return
	}

	// Import clinical data
	result, err := ccda.ImportCCD(model.SqlDb, parsed, patientID, sess.UserId)
	if err != nil {
		log.Printf("ccdaImportCCD: import failed: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "ok",
		"patient_id":    patientID,
		"imported":      result.Total,
		"allergies":     result.Allergies,
		"medications":   result.Medications,
		"problems":      result.Problems,
		"vitals":        result.Vitals,
		"immunizations": result.Immunizations,
		"procedures":    result.Procedures,
		"results":       result.Results,
		"social_history": result.SocialHistory,
		"encounters":    result.Encounters,
	})
}
