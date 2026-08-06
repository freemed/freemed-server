package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func patientProgressNotes(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	notes, err := model.Queries.PatientProgressNotes(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, notes)
}
