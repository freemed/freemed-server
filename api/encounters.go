package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func patientEncounters(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	patientID := common.ParseInt(id)
	encounters, err := model.Queries.ListEncounters(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, encounters)
}

func patientEncounterDetail(r *gin.Context) {
	id := r.Param("id")
	encounterID := r.Param("encounterId")
	if id == "" || encounterID == "" {
		common.ErrorResponse(r, http.StatusBadRequest, "bad request")
		return
	}

	patientID := common.ParseInt(id)
	encounter, err := model.Queries.GetEncounter(r.Request.Context(), dbgen.GetEncounterParams{
		PatientID:   patientID,
		EncounterID: common.ParseInt(encounterID),
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, encounter)
}
