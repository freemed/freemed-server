package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

// VitalsInput is the JSON payload for creating a vitals record.
type VitalsInput struct {
	DateTaken        string  `json:"date_taken"`
	Systolic         *int32  `json:"systolic"`
	Diastolic        *int32  `json:"diastolic"`
	HeartRate        *int32  `json:"heart_rate"`
	RespiratoryRate  *int32  `json:"respiratory_rate"`
	Temperature      *string `json:"temperature"`
	OxygenSaturation *int32  `json:"oxygen_saturation"`
	HeightCm         *string `json:"height_cm"`
	WeightKg         *string `json:"weight_kg"`
	Bmi              *string `json:"bmi"`
	Notes            *string `json:"notes"`
}

func patientVitalsList(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	vitals, err := model.Queries.ListVitals(r.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, vitals)
}

func patientVitalsLatest(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	patientID := common.ParseInt(id)
	v, err := model.Queries.GetLatestVitals(r.Request.Context(), patientID)
	if err != nil {
		if err == sql.ErrNoRows {
			r.JSON(http.StatusOK, nil)
			return
		}
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	r.JSON(http.StatusOK, v)
}

func patientVitalsCreate(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	session, err := common.GetSession(r)
	if err != nil {
		log.Printf("patientVitalsCreate: failed to get session: %v", err)
		r.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	var input VitalsInput
	if err := r.BindJSON(&input); err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	dateTaken, err := common.ParseDate(input.DateTaken)
	if err != nil {
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Compute BMI automatically if height and weight are provided but BMI is not
	bmi := toNullString(input.Bmi)
	if (!bmi.Valid || bmi.String == "") && input.HeightCm != nil && input.WeightKg != nil {
		heightStr := *input.HeightCm
		weightStr := *input.WeightKg
		if h, errH := parseFloat(heightStr); errH == nil && h > 0 {
			if w, errW := parseFloat(weightStr); errW == nil && w > 0 {
				calculated := w / (h * h) * 10000 // height in cm, convert to meters
				bmi = sql.NullString{String: formatFloat(calculated), Valid: true}
			}
		}
	}

	params := dbgen.CreateVitalsParams{
		PatientID:        common.ParseInt(id),
		DateTaken:        dateTaken,
		Systolic:         toNullInt32(input.Systolic),
		Diastolic:        toNullInt32(input.Diastolic),
		HeartRate:        toNullInt32(input.HeartRate),
		RespiratoryRate:  toNullInt32(input.RespiratoryRate),
		Temperature:      toNullString(input.Temperature),
		OxygenSaturation: toNullInt32(input.OxygenSaturation),
		HeightCm:         toNullString(input.HeightCm),
		WeightKg:         toNullString(input.WeightKg),
		Bmi:              bmi,
		Notes:            toNullString(input.Notes),
		UserID:           session.UserId,
	}

	result, err := model.Queries.CreateVitals(r.Request.Context(), params)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	r.JSON(http.StatusCreated, gin.H{"id": newID})
}

// Helper: convert *int32 to sql.NullInt32
func toNullInt32(v *int32) sql.NullInt32 {
	if v == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *v, Valid: true}
}

// Helper: convert *string to sql.NullString
func toNullString(v *string) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *v, Valid: true}
}

// Helper: parse a decimal string to float64
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// Helper: format a float64 to one-decimal string
func formatFloat(f float64) string {
	return fmt.Sprintf("%.1f", f)
}
