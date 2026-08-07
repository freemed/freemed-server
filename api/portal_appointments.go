package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/config"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

// portalAppointmentRequestInput is the JSON body for POST /portal/appointments/request.
type portalAppointmentRequestInput struct {
	Date       string `json:"date"        binding:"required"`
	Hour       int64  `json:"hour"        binding:"required"`
	Minute     int64  `json:"minute"      binding:"required"`
	ProviderID int64  `json:"provider_id" binding:"required"`
	Reason     string `json:"reason"`
}

// extractPortalPatientID extracts the patient_id from portal JWT claims.
// Returns the patient ID and true if the token is a valid portal patient token.
func extractPortalPatientID(c *gin.Context) (int64, bool) {
	claims := jwt.ExtractClaims(c)
	if len(claims) == 0 {
		return 0, false
	}

	// Check that this is a portal token (role=patient, portal=true)
	role, _ := claims["role"]
	portal, _ := claims["portal"]
	if role != "patient" || portal != true {
		return 0, false
	}

	patientID, ok := claims["patient_id"]
	if !ok {
		return 0, false
	}

	var pid int64
	switch v := patientID.(type) {
	case float64:
		pid = int64(v)
	case int64:
		pid = v
	default:
		return 0, false
	}
	return pid, true
}

// PortalAppointmentRequest handles POST /portal/appointments/request.
// Validates scheduling hours, interval alignment, future date, and conflict detection.
func PortalAppointmentRequest(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "invalid portal session")
		return
	}

	var input portalAppointmentRequestInput
	if err := c.ShouldBind(&input); err != nil {
		log.Printf("portalAppointmentRequest: bind error: %v", err)
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	calDate, err := common.ParseDate(input.Date)
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "invalid date format; expected YYYY-MM-DD")
		return
	}

	// --- Scheduling hours validation ---
	startHour := int64(config.Config.Scheduler.Start)
	endHour := int64(config.Config.Scheduler.End)
	interval := int64(config.Config.Scheduler.Interval)

	// Validate hour range
	if input.Hour < startHour || input.Hour >= endHour {
		msg := fmt.Sprintf(
			"Appointments can only be requested between %d:00 AM and %d:00 PM",
			startHour,
			endHour,
		)
		common.ErrorResponse(c, http.StatusBadRequest, msg)
		return
	}

	// Validate interval alignment (minute must be a multiple of Interval)
	if input.Minute%interval != 0 {
		msg := fmt.Sprintf(
			"Appointment minutes must be in %d-minute intervals (e.g., 0, %d, %d, ...)",
			interval,
			interval,
			interval*2,
		)
		common.ErrorResponse(c, http.StatusBadRequest, msg)
		return
	}

	// --- Future date requirement ---
	today := time.Now().Truncate(24 * time.Hour)
	requestDate := calDate.Truncate(24 * time.Hour)
	if !requestDate.After(today) {
		common.ErrorResponse(c, http.StatusBadRequest, "Appointments must be requested for tomorrow or later")
		return
	}

	// --- Conflict detection ---
	// Fetch existing appointments for the same date and provider
	existingAppts, err := model.Queries.SchedulerFindDateApptByProvider(
		c.Request.Context(),
		dbgen.SchedulerFindDateApptByProviderParams{
			ReqDate:    calDate,
			ProviderID: input.ProviderID,
		},
	)
	if err != nil {
		log.Printf("portalAppointmentRequest: conflict check error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	// Check if requested time slot overlaps with any existing appointment.
	// Default duration is 15 minutes (matches the interval).
	requestDuration := int64(15)
	requestStartMin := input.Hour*60 + input.Minute
	requestEndMin := requestStartMin + requestDuration

	for _, appt := range existingAppts {
		existingStartMin := appt.Calhour*60 + appt.Calminute
		existingEndMin := existingStartMin + appt.Calduration

		// Overlap check: two intervals [A,B) and [C,D) overlap if A < D AND C < B
		if requestStartMin < existingEndMin && existingStartMin < requestEndMin {
			common.ErrorResponse(
				c,
				http.StatusConflict,
				"This time slot is already booked. Please choose a different time.",
			)
			return
		}
	}

	// --- Create the appointment with status 'requested' ---
	result, err := model.Queries.CreatePortalAppointmentRequest(c.Request.Context(), dbgen.CreatePortalAppointmentRequestParams{
		DateOf:     calDate,
		Hour:       input.Hour,
		Minute:     input.Minute,
		ProviderID: input.ProviderID,
		PatientID:  patientID,
		Note:       input.Reason,
		UserID:     0, // Portal appointments are self-service; no authenticated staff user
	})
	if err != nil {
		log.Printf("portalAppointmentRequest: create error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{
		"id":      newID,
		"status":  "requested",
		"message": "Appointment requested successfully",
	})
}
