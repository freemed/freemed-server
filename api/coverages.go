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
	common.ApiMap["coverages"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/patient/:id/coverages", coveragesList)
			r.POST("/patient/:id/coverages", coverageCreate)
			r.DELETE("/patient/:id/coverages/:coverageId", coverageRemove)
		},
	}
}

type coverageInput struct {
	InsuranceCompany int64  `json:"insurance_company" binding:"required"`
	CoverageType     int64  `json:"coverage_type" binding:"required"`
	PolicyNumber     string `json:"policy_number"`
	GroupNumber      string `json:"group_number"`
	EffectiveDate    string `json:"effective_date"`
	TerminationDate  string `json:"termination_date"`
	PrimaryCoverage  bool   `json:"primary_coverage"`
}

func coveragesList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	coverages, err := model.Queries.ListCoverages(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, coverages)
}

func coverageCreate(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var in coverageInput
	if err := r.BindJSON(&in); err != nil {
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	result, err := model.Queries.CreateCoverage(r.Request.Context(), dbgen.CreateCoverageParams{
		Patient:          patientID,
		InsuranceCompany: in.InsuranceCompany,
		CoverageType:     in.CoverageType,
		PolicyNumber:     in.PolicyNumber,
		GroupNumber:      in.GroupNumber,
		EffectiveDate:    parseOptionalDate(in.EffectiveDate),
		TerminationDate:  parseOptionalDate(in.TerminationDate),
		PrimaryCoverage:  in.PrimaryCoverage,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}

func coverageRemove(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	coverageID := common.ParseInt(r.Param("coverageId"))
	if coverageID == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err := model.Queries.RemoveCoverage(r.Request.Context(), dbgen.RemoveCoverageParams{
		CoverageID: coverageID,
		PatientID:  patientID,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{"status": "deactivated"})
}
