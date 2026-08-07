package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

// phoneInput is the JSON body for POST /api/patient/:id/phones
type phoneInput struct {
	PhoneType string `json:"phone_type" binding:"required"`
	Number    string `json:"number" binding:"required"`
}

// patientPhonesList handles GET /api/patient/:id/phones
func patientPhonesList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	phones, err := model.Queries.ListPhones(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, phones)
}

// patientPhonesCreate handles POST /api/patient/:id/phones
func patientPhonesCreate(r *gin.Context) {
	patientID := common.ParseInt(r.Param("id"))
	if patientID == 0 {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var in phoneInput
	if err := r.BindJSON(&in); err != nil {
		common.ErrorResponseFromError(r, http.StatusBadRequest, err)
		return
	}

	// Extract user ID from JWT session
	sess, err := common.GetSession(r)
	if err != nil {
		common.ErrorResponseFromError(r, http.StatusUnauthorized, err)
		return
	}

	result, err := model.Queries.CreatePhone(r.Request.Context(), dbgen.CreatePhoneParams{
		PatientID: patientID,
		PhoneType: in.PhoneType,
		Number:    in.Number,
		UserID:    sess.UserId,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(r, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}
