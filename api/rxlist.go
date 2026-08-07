package api

import (
	"log"
	"net/http"
	"time"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["rxlist"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/:patientId", rxlistPrescriptions)
		},
	}
}

// rxlistPrescriptionResponse is the JSON response for a single prescription
type rxlistPrescriptionResponse struct {
	ID                  int64     `json:"id"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	Patient             int64     `json:"patient"`
	DrugName            string    `json:"drug_name"`
	Dosage              string    `json:"dosage"`
	Frequency           string    `json:"frequency"`
	Quantity            string    `json:"quantity"`
	Refills             int64     `json:"refills"`
	RefillsRemaining    int64     `json:"refills_remaining"`
	RefillsUsed         int64     `json:"refills_used"`
	DateWritten         time.Time `json:"date_written"`
	PrescribingProvider int64     `json:"prescribing_provider"`
	Status              string    `json:"status"`
	Notes               *string   `json:"notes"`
	User                int64     `json:"user"`
	Pharmacy            *struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		City  string `json:"city,omitempty"`
		State string `json:"state,omitempty"`
	} `json:"pharmacy"`
	LastFillDate *time.Time `json:"last_fill_date,omitempty"`
}

// rxlistPrescriptions handles GET /api/rxlist/:patientId
func rxlistPrescriptions(c *gin.Context) {
	patientID := common.ParseInt(c.Param("patientId"))
	if patientID == 0 {
		common.ErrorResponse(c, http.StatusBadRequest, "invalid patient ID")
		return
	}

	rows, err := model.Queries.ListPrescriptionsWithPharmacy(c.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	response := make([]rxlistPrescriptionResponse, 0, len(rows))
	for _, row := range rows {
		item := rxlistPrescriptionResponse{
			ID:                  row.ID,
			CreatedAt:           row.CreatedAt,
			UpdatedAt:           row.UpdatedAt,
			Patient:             row.Patient,
			DrugName:            row.DrugName,
			Dosage:              row.Dosage,
			Frequency:           row.Frequency,
			Quantity:            row.Quantity,
			Refills:             row.Refills,
			RefillsUsed:         row.RefillsUsed,
			DateWritten:         row.DateWritten,
			PrescribingProvider: row.PrescribingProvider,
			Status:              row.Status,
			User:                row.User,
		}

		// Compute refills remaining
		item.RefillsRemaining = row.Refills - row.RefillsUsed
		if item.RefillsRemaining < 0 {
			item.RefillsRemaining = 0
		}

		// Notes
		if row.Notes.Valid {
			item.Notes = &row.Notes.String
		}

		// Pharmacy details
		if row.PharmacyName.Valid {
			item.Pharmacy = &struct {
				ID    int64  `json:"id"`
				Name  string `json:"name"`
				City  string `json:"city,omitempty"`
				State string `json:"state,omitempty"`
			}{
				Name: row.PharmacyName.String,
			}
			if row.PharmacyID.Valid {
				item.Pharmacy.ID = row.PharmacyID.Int64
			}
			if row.PharmacyCity.Valid {
				item.Pharmacy.City = row.PharmacyCity.String
			}
			if row.PharmacyState.Valid {
				item.Pharmacy.State = row.PharmacyState.String
			}
		}

		// Last fill date — sqlc returns interface{} for nullable subquery datetime
		if row.LastFillDate != nil {
			switch v := row.LastFillDate.(type) {
			case time.Time:
				item.LastFillDate = &v
			}
		}

		response = append(response, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"total": len(response),
	})
}
