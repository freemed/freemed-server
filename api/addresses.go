package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

// patientAddressesList handles GET /api/patient/:id/addresses
func patientAddressesList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	addresses, err := model.Queries.ListAddresses(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, addresses)
}

// patientAddressUpdate handles PUT /api/patient/:id/addresses/:addressId
func patientAddressUpdate(r *gin.Context) {
	id := r.Param("id")
	addressID := r.Param("addressId")
	if id == "" || addressID == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var input struct {
		Line1  string `json:"line1"`
		Line2  string `json:"line2"`
		City   string `json:"city"`
		State  string `json:"stpr"`
		Postal string `json:"postal"`
	}
	if err := r.BindJSON(&input); err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	patientID := common.ParseInt(id)
	addrID := common.ParseInt(addressID)

	if err := model.Queries.UpdateAddress(r.Request.Context(), dbgen.UpdateAddressParams{
		Line1:     addrNullString(input.Line1),
		Line2:     addrNullString(input.Line2),
		City:      addrNullString(input.City),
		Stpr:      addrNullString(input.State),
		Postal:    addrNullString(input.Postal),
		AddressID: addrID,
		PatientID: patientID,
	}); err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// patientAddressDelete handles DELETE /api/patient/:id/addresses/:addressId
func patientAddressDelete(r *gin.Context) {
	id := r.Param("id")
	addressID := r.Param("addressId")
	if id == "" || addressID == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	addrID := common.ParseInt(addressID)

	if err := model.Queries.DeleteAddress(r.Request.Context(), dbgen.DeleteAddressParams{
		AddressID: addrID,
		PatientID: patientID,
	}); err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, gin.H{"status": "deactivated"})
}

// addrNullString converts a string to sql.NullString, treating empty as NULL.
func addrNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
