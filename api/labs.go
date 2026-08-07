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
	common.ApiMap["labs"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/", listAllLabs)
		},
	}
}

type labInput struct {
	LabName        string `json:"lab_name" binding:"required"`
	LabDate        string `json:"lab_date" binding:"required"`
	Result         string `json:"result"`
	Unit           string `json:"unit"`
	ReferenceRange string `json:"reference_range"`
	Status         string `json:"status"`
	Notes          string `json:"notes"`
}

// patientLabsList handles GET /api/patient/:id/labs
func patientLabsList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	labs, err := model.Queries.ListLabs(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, labs)
}

// patientLabsCreate handles POST /api/patient/:id/labs
func patientLabsCreate(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var in labInput
	if err := r.BindJSON(&in); err != nil {
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	labDate, err := common.ParseDate(in.LabDate)
	if err != nil {
		r.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	sess, err := common.GetSession(r)
	if err != nil {
		r.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	result, err := model.Queries.CreateLab(r.Request.Context(), dbgen.CreateLabParams{
		PatientID:      patientID,
		LabName:        in.LabName,
		LabDate:        labDate,
		Result:         in.Result,
		Unit:           in.Unit,
		ReferenceRange: in.ReferenceRange,
		Status:         in.Status,
		Notes:          toNullString(&in.Notes),
		UserID:         sess.UserId,
	})
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newID, err := result.LastInsertId()
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusCreated, gin.H{"id": newID})
}

// listAllLabs handles GET /api/labs/
func listAllLabs(r *gin.Context) {
	r.JSON(http.StatusOK, []interface{}{})
}
