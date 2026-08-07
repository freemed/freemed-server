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
	common.ApiMap["payments"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.POST("/:id/attach-procedure", paymentsAttachToProcedure)
			r.GET("/unattached/copays", paymentsUnattachedCopays)
			r.GET("/unattached/deductibles", paymentsUnattachedDeductibles)
			r.GET("/unattached/payments", paymentsUnattachedPayments)
			r.POST("/:id/mistake", paymentsMarkAsMistake)
		},
	}
}

func init() {
	common.ApiMap["ledger"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/:patientId", standaloneLedger)
		},
	}
}

type attachProcedureInput struct {
	ProcedureID int64 `json:"procedure_id" binding:"required"`
}

// paymentsAttachToProcedure handles POST /api/payments/:id/attach-procedure
func paymentsAttachToProcedure(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var in attachProcedureInput
	if err := r.BindJSON(&in); err != nil {
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err := model.Queries.AttachPaymentToProcedure(r.Request.Context(), dbgen.AttachPaymentToProcedureParams{
		PaymentID:   common.ParseInt(id),
		ProcedureID: in.ProcedureID,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{"status": "attached"})
}

// paymentsUnattachedCopays handles GET /api/payments/unattached/copays
func paymentsUnattachedCopays(r *gin.Context) {
	rows, err := model.Queries.UnattachedCopays(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}

// paymentsUnattachedDeductibles handles GET /api/payments/unattached/deductibles
func paymentsUnattachedDeductibles(r *gin.Context) {
	rows, err := model.Queries.UnattachedDeductibles(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}

// paymentsUnattachedPayments handles GET /api/payments/unattached/payments
func paymentsUnattachedPayments(r *gin.Context) {
	rows, err := model.Queries.UnattachedPayments(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, rows)
}

// paymentsMarkAsMistake handles POST /api/payments/:id/mistake
func paymentsMarkAsMistake(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err := model.Queries.RemovePaymentAsMistake(r.Request.Context(), common.ParseInt(id))
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{"status": "inactive"})
}

// patientPayments handles GET /api/patient/:id/payments
func patientPayments(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	payments, err := model.Queries.PatientPayments(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, payments)
}

// patientLedger handles GET /api/patient/:id/ledger
func patientLedger(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	patientID := common.ParseInt(id)
	fromDate := parseOptionalDate(r.Query("from"))
	toDate := parseOptionalDate(r.Query("to"))
	offset := common.ParseInt(r.DefaultQuery("offset", "0"))
	limit := common.ParseInt(r.DefaultQuery("limit", "50"))

	ledger, err := model.Queries.PatientLedger(r.Request.Context(), dbgen.PatientLedgerParams{
		PatientID: patientID,
		FromDate:  fromDate,
		ToDate:    toDate,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	total, err := model.Queries.CountPatientLedger(r.Request.Context(), dbgen.CountPatientLedgerParams{
		PatientID: patientID,
		FromDate:  fromDate,
		ToDate:    toDate,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{
		"data":   ledger,
		"total":  total,
		"offset": offset,
		"limit":  limit,
	})
}

// standaloneLedger handles GET /api/ledger/:patientId
func standaloneLedger(r *gin.Context) {
	patientID := r.Param("patientId")
	if patientID == "" {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	pid := common.ParseInt(patientID)
	fromDate := parseOptionalDate(r.Query("from"))
	toDate := parseOptionalDate(r.Query("to"))
	offset := common.ParseInt(r.DefaultQuery("offset", "0"))
	limit := common.ParseInt(r.DefaultQuery("limit", "50"))

	ledger, err := model.Queries.PatientLedger(r.Request.Context(), dbgen.PatientLedgerParams{
		PatientID: pid,
		FromDate:  fromDate,
		ToDate:    toDate,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	total, err := model.Queries.CountPatientLedger(r.Request.Context(), dbgen.CountPatientLedgerParams{
		PatientID: pid,
		FromDate:  fromDate,
		ToDate:    toDate,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{
		"data":   ledger,
		"total":  total,
		"offset": offset,
		"limit":  limit,
	})
}

// patientCoverageInfo handles GET /api/patient/:id/coverage-info
func patientCoverageInfo(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	coverage, err := model.Queries.GetCoverageCopayInfo(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, coverage)
}
