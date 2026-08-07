package api

import (
	"log"
	"net/http"
	"time"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

type drugSampleInput struct {
	DrugCode             string `json:"drug_code"`
	NDC                  string `json:"ndc"`
	DrugClass            string `json:"drug_class"`
	PackageCount         int64  `json:"package_count"`
	Location             string `json:"location"`
	DrugCompany          string `json:"drug_company"`
	DrugRepresentative   string `json:"drug_representative"`
	Invoice              string `json:"invoice"`
	SampleCount          int64  `json:"sample_count"`
	SampleCountRemaining int64  `json:"sample_count_remaining"`
	Lot                  string `json:"lot"`
	Expiration           string `json:"expiration"`
	Received             string `json:"received"`
	AssignedTo           string `json:"assigned_to"`
}

func patientDrugSamplesList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	rows, err := model.Queries.ListDrugSamples(r.Request.Context())
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, rows)
}

func patientDrugSampleCreate(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	session, err := common.GetSession(r)
	if err != nil {
		log.Printf("patientDrugSampleCreate: failed to get session: %v", err)
		r.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	var in drugSampleInput
	if err := r.BindJSON(&in); err != nil {
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Default sample_count_remaining to sample_count if not provided
	sampleCountRemaining := in.SampleCountRemaining
	if sampleCountRemaining == 0 && in.SampleCount > 0 {
		sampleCountRemaining = in.SampleCount
	}

	result, err := model.Queries.CreateDrugSample(r.Request.Context(), dbgen.CreateDrugSampleParams{
		Drugcode:          in.DrugCode,
		Drugndc:           in.NDC,
		Drugclass:         in.DrugClass,
		Packagecount:      in.PackageCount,
		Location:          in.Location,
		Drugco:            in.DrugCompany,
		Drugrep:           in.DrugRepresentative,
		Invoice:           in.Invoice,
		Samplecount:       in.SampleCount,
		Samplecountremain: sampleCountRemaining,
		Lot:               in.Lot,
		Expiration:        parseOptionalDate(in.Expiration),
		Received:          parseOptionalDate(in.Received),
		Assignedto:        in.AssignedTo,
		Loguser:           session.UserId,
		Logdate:           time.Now(),
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}
