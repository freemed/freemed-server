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
	common.ApiMap["patient"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/:id/info", patientInformation)
			r.GET("/:id/progress-notes", patientProgressNotes)
			r.GET("/:id/diagnoses", patientDiagnoses)
			r.GET("/:id/attachments", patientEmrAttachments)
			r.GET("/:id/attachments/:module", patientEmrAttachments)
			r.GET("/:id/vitals", patientVitalsList)
			r.GET("/:id/vitals/latest", patientVitalsLatest)
			r.POST("/:id/vitals", patientVitalsCreate)
			r.GET("/:id/encounters", patientEncounters)
			r.GET("/:id/encounters/:encounterId", patientEncounterDetail)
			r.GET("/:id/payments", patientPayments)
			r.GET("/:id/ledger", patientLedger)
			r.GET("/:id/procedures", patientProcedures)
			r.GET("/:id/authorizations", patientAuthorizations)
			r.GET("/:id/claims", patientClaims)
			r.GET("/:id/referrals", patientReferralsList)
			r.POST("/:id/referrals", patientReferralsCreate)
			r.GET("/:id/immunizations", patientImmunizationsList)
			r.POST("/:id/immunizations", patientImmunizationsCreate)
			r.GET("/:id/allergies", patientAllergiesList)
			r.POST("/:id/allergies", patientAllergiesCreate)
			r.GET("/:id/addresses", patientAddressesList)
			r.PUT("/:id/addresses/:addressId", patientAddressUpdate)
			r.DELETE("/:id/addresses/:addressId", patientAddressDelete)
			r.GET("/:id/tags", patientTagsList)
			r.POST("/:id/tags", patientTagsCreate)
			r.DELETE("/:id/tags/:tagId", patientTagsExpire)
			r.GET("/:id/drug-samples", patientDrugSamplesList)
			r.POST("/:id/drug-samples", patientDrugSampleCreate)
			r.GET("/:id/history", patientHistory)
			},
	}
}

func patientEmrAttachments(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	module := r.Param("module")

	if module == "" {
		o, err := model.Queries.PatientEmrAttachments(r.Request.Context(), patientID)
		if err != nil {
			log.Print(err.Error())
			r.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		r.JSON(http.StatusOK, o)
		return
	}

	o, err := model.Queries.PatientEmrAttachmentsByModule(r.Request.Context(), dbgen.PatientEmrAttachmentsByModuleParams{
		PatientID: patientID,
		Module:    module,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, o)
}

func patientDiagnoses(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	diagnoses, err := model.Queries.PatientDiagnoses(r.Request.Context(), dbgen.PatientDiagnosesParams{
		PatientID: patientID,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, diagnoses)
}

func patientInformation(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	o, err := model.Queries.PatientInformation(r.Request.Context(), dbgen.PatientInformationParams{
		PatientID: patientID,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, o)
}

func patientHistory(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	rows, err := model.Queries.PatientHistory(r.Request.Context(), dbgen.PatientHistoryParams{
		PatientID: patientID,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}
