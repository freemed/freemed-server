package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["allergies"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.DELETE("/:allergyId", allergiesDeactivate)
		},
	}
}

// patientAllergiesList handles GET /api/patient/:id/allergies
func patientAllergiesList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	allergies, err := model.Queries.ListAllergies(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, allergies)
}

// patientAllergiesCreate handles POST /api/patient/:id/allergies
func patientAllergiesCreate(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	result, err := model.Queries.CreateAllergy(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	allergyID, err := result.LastInsertId()
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusCreated, gin.H{"id": allergyID})
}

// allergiesDeactivate handles DELETE /api/allergies/:allergyId
func allergiesDeactivate(r *gin.Context) {
	allergyIDStr := r.Param("allergyId")
	if allergyIDStr == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	allergyID := common.ParseInt(allergyIDStr)
	if err := model.Queries.DeactivateAllergy(r.Request.Context(), allergyID); err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{"status": "deactivated"})
}
