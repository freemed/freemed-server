package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["signatures"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/:id", signatureGet)
		},
	}
}

// signatureGet handles GET /api/signatures/:id
// Returns the raw signature blob with the appropriate Content-Type.
func signatureGet(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	sigID := common.ParseInt(id)
	sig, err := model.Queries.GetSignature(r.Request.Context(), sigID)
	if err != nil {
		if err == sql.ErrNoRows {
			r.AbortWithStatus(http.StatusNotFound)
			return
		}
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	contentType := "application/octet-stream"
	switch sig.Format {
	case "JPG", "JPEG":
		contentType = "image/jpeg"
	case "PNG":
		contentType = "image/png"
	case "TOPAZ":
		contentType = "application/octet-stream"
	}

	r.Data(http.StatusOK, contentType, []byte(sig.SignatureData.String))
}

// patientSignaturesList handles GET /api/patient/:id/signatures
func patientSignaturesList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	rows, err := model.Queries.ListSignaturesByPatient(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, rows)
}

type signatureInput struct {
	Module             string  `json:"module" binding:"required"`
	ModuleField        string  `json:"module_field" binding:"required"`
	Oid                int64   `json:"oid" binding:"required"`
	SignatureData      []byte  `json:"signature_data"`
	Format             string  `json:"format"`
	CollectorLocation  *string `json:"collector_location"`
	CollectorModel     *string `json:"collector_model"`
	CollectorJobid     *string `json:"collector_jobid"`
	CollectorFinished  bool    `json:"collector_finished"`
}

// patientSignaturesCreate handles POST /api/patient/:id/signatures
func patientSignaturesCreate(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	session, err := common.GetSession(r)
	if err != nil {
		log.Printf("patientSignaturesCreate: failed to get session: %v", err)
		common.ErrorResponseFromError(r, http.StatusUnauthorized, err)
		return
	}

	var in signatureInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	format := in.Format
	if format == "" {
		format = "UNKNOWN"
	}

	sigData := sql.NullString{}
	if in.SignatureData != nil {
		sigData = sql.NullString{String: string(in.SignatureData), Valid: true}
	}

	collectorLocation := toNullString(in.CollectorLocation)
	collectorModel := toNullString(in.CollectorModel)
	collectorJobid := toNullString(in.CollectorJobid)

	result, err := model.Queries.CreateSignature(r.Request.Context(), dbgen.CreateSignatureParams{
		Patient:            patientID,
		Module:             in.Module,
		ModuleField:        in.ModuleField,
		Oid:                in.Oid,
		SignatureData:      sigData,
		Format:             format,
		CollectorLocation:  collectorLocation,
		CollectorModel:     collectorModel,
		CollectorJobid:     collectorJobid,
		CollectorFinished:  in.CollectorFinished,
		User:               session.UserId,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}
