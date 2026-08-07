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

// certificationInput is the JSON payload for creating a certification record.
type certificationInput struct {
	CertType    int64   `json:"cert_type" binding:"required"`
	CertFormNum *int64  `json:"cert_form_num"`
	CertDesc    string  `json:"cert_desc" binding:"required"`
	CertFormData *string `json:"cert_form_data"`
}

// patientCertificationsList handles GET /api/patient/:id/certifications
func patientCertificationsList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	patientID := common.ParseInt(id)
	certs, err := model.Queries.ListCertificationsByPatient(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, certs)
}

// patientCertificationsCreate handles POST /api/patient/:id/certifications
func patientCertificationsCreate(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	session, err := common.GetSession(r)
	if err != nil {
		log.Printf("patientCertificationsCreate: failed to get session: %v", err)
		common.ErrorResponseFromError(r, http.StatusUnauthorized, err)
		return
	}

	var in certificationInput
	if err := r.BindJSON(&in); err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	result, err := model.Queries.CreateCertification(r.Request.Context(), dbgen.CreateCertificationParams{
		PatientID:    patientID,
		CertType:     in.CertType,
		CertFormNum:  toNullInt64(in.CertFormNum),
		CertDesc:     in.CertDesc,
		CertFormData: toNullString(in.CertFormData),
		UserID:       session.UserId,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}

// toNullInt64 converts *int64 to sql.NullInt64.
func toNullInt64(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}
