package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

// ClinicalOrderInput is the JSON payload for creating a clinical order record.
type ClinicalOrderInput struct {
	OrderType        string `json:"order_type"`
	Description      string `json:"description"`
	Status           string `json:"status"`
	DateOrdered      string `json:"date_ordered"`
	OrderingProvider int64  `json:"ordering_provider"`
	Notes            string `json:"notes"`
}

func patientClinicalOrdersList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	orders, err := model.Queries.ListClinicalOrders(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, orders)
}

func patientClinicalOrdersCreate(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	session, err := common.GetSession(r)
	if err != nil {
		log.Printf("patientClinicalOrdersCreate: failed to get session: %v", err)
		r.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	var input ClinicalOrderInput
	if err := r.BindJSON(&input); err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	dateOrdered := parseOptionalDate(input.DateOrdered)

	params := dbgen.CreateClinicalOrderParams{
		PatientID:        common.ParseInt(id),
		OrderType:        input.OrderType,
		Description:      input.Description,
		Status:           input.Status,
		DateOrdered:      dateOrdered,
		OrderingProvider: input.OrderingProvider,
		Notes:            input.Notes,
		UserID:           session.UserId,
	}

	result, err := model.Queries.CreateClinicalOrder(r.Request.Context(), params)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}
