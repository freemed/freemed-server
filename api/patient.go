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
			r.GET("/:id/attachments", patientEmrAttachments)
			r.GET("/:id/attachments/:module", patientEmrAttachments)
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
