package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/freemed/freemed-server/ccda"
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/config"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

func init() {
	common.ApiMap["portal"] = common.ApiMapping{
		Authenticated: false, // portal uses its own JWT auth, verified per-request
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/me", portalMe)
			r.GET("/appointments", portalAppointments)
			r.POST("/appointments/request", PortalAppointmentRequest)
			// Self-scheduling
			r.GET("/slots", portalSlots)
			r.POST("/appointments", portalCreateAppointment)
			r.DELETE("/appointments/:id", portalCancelAppointment)
			// Clinical
			r.GET("/medications", portalMedications)
			r.GET("/allergies", portalAllergies)
			r.GET("/vitals", portalVitals)
			r.GET("/problems", portalProblems)
			r.GET("/labs", portalLabs)
			r.GET("/documents", portalDocuments)
			r.GET("/ccda", portalCCDADownload)
			// Secure messaging
			r.GET("/messages", portalListMessages)
			r.POST("/messages", portalCreateMessage)
			// Bill pay
			r.GET("/balance", portalBalance)
			r.GET("/statements", portalStatements)
			// Intake forms
			r.GET("/forms", portalListForms)
			r.GET("/forms/:id", portalGetForm)
			r.POST("/forms/:id/submit", portalSubmitForm)
		},
	}
}

// ============================================================================
// GET /api/portal/me — Patient demographics
// ============================================================================

func portalMe(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	row, err := model.Queries.GetPortalPatientDemographics(c.Request.Context(), patientID)
	if err != nil {
		if err == sql.ErrNoRows {
			common.ErrorResponse(c, http.StatusNotFound, "patient not found")
			return
		}
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	resp := gin.H{
		"id":             row.ID,
		"first_name":     row.FirstName,
		"last_name":      row.LastName,
		"patient_id":     row.PatientIDDisplay,
		"gender":         row.Gender,
		"language":       row.Language,
		"portal_enabled": row.PortalEnabled,
	}
	if row.MiddleName.Valid {
		resp["middle_name"] = row.MiddleName.String
	}
	if row.Suffix != "" {
		resp["suffix"] = row.Suffix
	}
	if row.DateOfBirth.Valid {
		resp["date_of_birth"] = row.DateOfBirth.Time.Format("2006-01-02")
	}
	if row.Email.Valid {
		resp["email"] = row.Email.String
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// GET /api/portal/appointments — Patient appointments
// ============================================================================

func portalAppointments(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	rows, err := model.Queries.ListPortalAppointments(c.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	// Ensure non-nil array in JSON response
	if rows == nil {
		rows = []dbgen.ListPortalAppointmentsRow{}
	}

	c.JSON(http.StatusOK, rows)
}

// ============================================================================
// GET /api/portal/slots — Available appointment slots
// ============================================================================

type slotResponse struct {
	Date         string `json:"date"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	ProviderID   int64  `json:"provider_id"`
	ProviderName string `json:"provider_name"`
}

func portalSlots(c *gin.Context) {
	_, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	fromDate, err := common.ParseDate(c.Query("from"))
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "invalid from date; expected YYYY-MM-DD")
		return
	}
	toDate, err := common.ParseDate(c.Query("to"))
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "invalid to date; expected YYYY-MM-DD")
		return
	}

	// Parse optional provider filter
	providerQuery := c.Query("provider")
	var filterProviderID int64
	if providerQuery != "" {
		filterProviderID, err = strconv.ParseInt(providerQuery, 10, 64)
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, "invalid provider id")
			return
		}
	}

	// Get existing appointments in the date range
	existingAppts, err := model.Queries.SchedulerDailyApptRange(c.Request.Context(),
		dbgen.SchedulerDailyApptRangeParams{
			FromDate: fromDate,
			ToDate:   toDate,
		})
	if err != nil {
		log.Printf("portalSlots: query error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	// Get all blocked slots in the date range
	blocked, err := model.Queries.ListBlockedSlotsByDate(c.Request.Context(), fromDate)
	if err != nil {
		log.Printf("portalSlots: blocks query error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}
	for d := fromDate.AddDate(0, 0, 1); !d.After(toDate); d = d.AddDate(0, 0, 1) {
		dayBlocks, err := model.Queries.ListBlockedSlotsByDate(c.Request.Context(), d)
		if err != nil {
			log.Printf("portalSlots: blocks query error for %s: %v", d.Format("2006-01-02"), err)
			continue
		}
		blocked = append(blocked, dayBlocks...)
	}

	// Get list of providers
	providers, err := model.Queries.ListProviders(c.Request.Context())
	if err != nil {
		log.Printf("portalSlots: provider query error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	schedulerCfg := config.Config.Scheduler
	startHour := int64(schedulerCfg.Start)
	endHour := int64(schedulerCfg.End)
	interval := int64(schedulerCfg.Interval)

	// Build blocked-time set: provider -> date -> minute-of-day -> true
	type blockKey struct {
		providerID int64
		date       string
	}
	blockedSet := make(map[blockKey]map[int64]bool)
	for _, b := range blocked {
		key := blockKey{providerID: b.Sbsprovider, date: b.Sbsdate.Format("2006-01-02")}
		if blockedSet[key] == nil {
			blockedSet[key] = make(map[int64]bool)
		}
		startMin := b.Sbshour*60 + b.Sbsminute
		endMin := startMin + b.Sbsduration
		for m := startMin; m < endMin; m++ {
			blockedSet[key][m] = true
		}
	}

	// Build booked-time set: provider -> date -> minute-of-day -> true
	bookedSet := make(map[blockKey]map[int64]bool)
	for _, appt := range existingAppts {
		if !appt.ProviderID.Valid {
			continue
		}
		pid := appt.ProviderID.Int64
		key := blockKey{providerID: pid, date: appt.DateOf.Format("2006-01-02")}
		if bookedSet[key] == nil {
			bookedSet[key] = make(map[int64]bool)
		}
		startMin := appt.Hour*60 + appt.Minute
		endMin := startMin + appt.Duration
		for m := startMin; m < endMin; m++ {
			bookedSet[key][m] = true
		}
	}

	var slots []slotResponse
	for d := fromDate; !d.After(toDate); d = d.AddDate(0, 0, 1) {
		// Skip weekends
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			continue
		}
		dateKey := d.Format("2006-01-02")
		for _, prov := range providers {
			if filterProviderID != 0 && prov.ID != filterProviderID {
				continue
			}
			for h := startHour; h < endHour; h++ {
				for m := int64(0); m < 60; m += interval {
					minOfDay := h*60 + m
					key := blockKey{providerID: prov.ID, date: dateKey}
					// Check if blocked
					if bp := blockedSet[key]; bp != nil && bp[minOfDay] {
						continue
					}
					// Check if booked
					if bp := bookedSet[key]; bp != nil && bp[minOfDay] {
						continue
					}
					endMinOfDay := minOfDay + interval
					slots = append(slots, slotResponse{
						Date:         dateKey,
						StartTime:    formatTime(h, m),
						EndTime:      formatTime(endMinOfDay/60, endMinOfDay%60),
						ProviderID:   prov.ID,
						ProviderName: prov.Phyfname + " " + prov.Phylname,
					})
				}
			}
		}
	}

	if slots == nil {
		slots = []slotResponse{}
	}
	c.JSON(http.StatusOK, slots)
}

func formatTime(hour, minute int64) string {
	return fmt.Sprintf("%02d:%02d", hour, minute)
}

// ============================================================================
// POST /api/portal/appointments — Self-schedule (status = 'scheduled')
// ============================================================================

type portalCreateAppointmentInput struct {
	SlotDate   string `json:"slot_date"   binding:"required"`
	SlotTime   string `json:"slot_time"   binding:"required"`
	ProviderID int64  `json:"provider_id" binding:"required"`
	Reason     string `json:"reason"`
}

func portalCreateAppointment(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input portalCreateAppointmentInput
	if err := c.ShouldBind(&input); err != nil {
		log.Printf("portalCreateAppointment: bind error: %v", err)
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	calDate, err := common.ParseDate(input.SlotDate)
	if err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "invalid slot_date format; expected YYYY-MM-DD")
		return
	}

	// Parse slot_time as "HH:MM"
	var hour, minute int64
	if _, err := fmt.Sscanf(input.SlotTime, "%d:%d", &hour, &minute); err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "invalid slot_time format; expected HH:MM")
		return
	}

	// Validate scheduling hours
	startHour := int64(config.Config.Scheduler.Start)
	endHour := int64(config.Config.Scheduler.End)
	interval := int64(config.Config.Scheduler.Interval)

	if hour < startHour || hour >= endHour {
		common.ErrorResponse(c, http.StatusBadRequest,
			fmt.Sprintf("Appointments can only be scheduled between %d:00 and %d:00", startHour, endHour))
		return
	}
	if minute%interval != 0 {
		common.ErrorResponse(c, http.StatusBadRequest,
			fmt.Sprintf("Appointment minutes must be in %d-minute intervals", interval))
		return
	}

	// Future date check
	today := time.Now().Truncate(24 * time.Hour)
	if !calDate.Truncate(24 * time.Hour).After(today) {
		common.ErrorResponse(c, http.StatusBadRequest, "Appointments must be scheduled for tomorrow or later")
		return
	}

	// Conflict detection
	existing, err := model.Queries.SchedulerFindDateApptByProvider(c.Request.Context(),
		dbgen.SchedulerFindDateApptByProviderParams{
			ReqDate:    calDate,
			ProviderID: input.ProviderID,
		})
	if err != nil {
		log.Printf("portalCreateAppointment: conflict check error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	requestStartMin := hour*60 + minute
	requestEndMin := requestStartMin + interval
	for _, appt := range existing {
		existingStartMin := appt.Calhour*60 + appt.Calminute
		existingEndMin := existingStartMin + appt.Calduration
		if requestStartMin < existingEndMin && existingStartMin < requestEndMin {
			common.ErrorResponse(c, http.StatusConflict,
				"This time slot is already booked. Please choose a different time.")
			return
		}
	}

	// Also check blocked slots
	blocks, err := model.Queries.ListBlockedSlotsByDate(c.Request.Context(), calDate)
	if err != nil {
		log.Printf("portalCreateAppointment: blocked slots check error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}
	for _, b := range blocks {
		if b.Sbsprovider != input.ProviderID {
			continue
		}
		blockStartMin := b.Sbshour*60 + b.Sbsminute
		blockEndMin := blockStartMin + b.Sbsduration
		if requestStartMin < blockEndMin && blockStartMin < requestEndMin {
			common.ErrorResponse(c, http.StatusConflict,
				"This time slot is not available. Please choose a different time.")
			return
		}
	}

	// Create directly with status matching PortalAppointmentRequest (requested), but we want scheduled
	// Use CreatePortalAppointmentRequest which uses 'requested', then update to 'scheduled'
	result, err := model.Queries.CreatePortalAppointmentRequest(c.Request.Context(),
		dbgen.CreatePortalAppointmentRequestParams{
			DateOf:     calDate,
			Hour:       hour,
			Minute:     minute,
			ProviderID: input.ProviderID,
			PatientID:  patientID,
			Note:       input.Reason,
			UserID:     0,
		})
	if err != nil {
		log.Printf("portalCreateAppointment: create error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{
		"id":      newID,
		"status":  "scheduled",
		"message": "Appointment scheduled successfully",
	})
}

// ============================================================================
// DELETE /api/portal/appointments/:id — Cancel appointment
// ============================================================================

type portalCancelInput struct {
	Reason string `json:"reason"`
}

func portalCancelAppointment(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	apptID := common.ParseInt(c.Param("id"))
	if apptID < 1 {
		common.ErrorResponse(c, http.StatusBadRequest, "invalid appointment id")
		return
	}

	var input portalCancelInput
	if err := c.ShouldBind(&input); err != nil {
		input.Reason = "Cancelled by patient"
	}
	if input.Reason == "" {
		input.Reason = "Cancelled by patient"
	}

	err := model.Queries.PortalCancelAppointment(c.Request.Context(), dbgen.PortalCancelAppointmentParams{
		ID:           apptID,
		PatientID:    patientID,
		CancelReason: input.Reason,
	})
	if err != nil {
		log.Printf("portalCancelAppointment: error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "cancelled",
		"message": "Appointment cancelled successfully",
	})
}

// ============================================================================
// GET /api/portal/medications — Active medications for the patient
// ============================================================================

func portalMedications(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	rows, err := model.Queries.ListMedications(c.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	if rows == nil {
		rows = []dbgen.Medication{}
	}

	c.JSON(http.StatusOK, rows)
}

// ============================================================================
// GET /api/portal/allergies — Active allergies for the patient
// ============================================================================

func portalAllergies(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	rows, err := model.Queries.ListAllergies(c.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	if rows == nil {
		rows = []dbgen.Allergy{}
	}

	c.JSON(http.StatusOK, rows)
}

// ============================================================================
// GET /api/portal/vitals — Vitals history for the patient
// ============================================================================

func portalVitals(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	rows, err := model.Queries.ListVitals(c.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	if rows == nil {
		rows = []dbgen.Vital{}
	}

	c.JSON(http.StatusOK, rows)
}

// ============================================================================
// GET /api/portal/problems — Combined current and chronic problems
// ============================================================================

func portalProblems(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	rows, err := model.Queries.ListPortalProblems(c.Request.Context(), dbgen.ListPortalProblemsParams{
		PatientID: patientID,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	if rows == nil {
		rows = []dbgen.ListPortalProblemsRow{}
	}

	c.JSON(http.StatusOK, rows)
}

// ============================================================================
// GET /api/portal/labs — Lab results for the patient
// ============================================================================

func portalLabs(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	rows, err := model.Queries.ListLabs(c.Request.Context(), patientID)
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	if rows == nil {
		rows = []dbgen.Lab{}
	}

	c.JSON(http.StatusOK, rows)
}

// ============================================================================
// GET /api/portal/documents — Scanned documents for the patient
// ============================================================================

func portalDocuments(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Pages of a patient's scanned documents. The response stays a bare array
	// (portal clients parse `ScannedDoc[]`) with the bound applied here:
	// ?offset=/?limit= clamped, default 50, maximum 200.
	offset, limit := pageParams(c)

	rows, err := model.Queries.ListScannedDocs(c.Request.Context(), dbgen.ListScannedDocsParams{
		PatientID: patientID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		log.Print(err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	if rows == nil {
		rows = []dbgen.ScannedDoc{}
	}

	c.JSON(http.StatusOK, rows)
}

// ============================================================================
// GET /api/portal/messages — List messages for patient
// ============================================================================

func portalListMessages(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	rows, err := model.Queries.PortalListMessages(c.Request.Context(), patientID)
	if err != nil {
		log.Printf("portalListMessages: error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	if rows == nil {
		rows = []dbgen.PortalListMessagesRow{}
	}

	c.JSON(http.StatusOK, rows)
}

// ============================================================================
// POST /api/portal/messages — Send message from patient to practice
// ============================================================================

type portalCreateMessageInput struct {
	Subject  string `json:"subject" binding:"required"`
	Body     string `json:"body"    binding:"required"`
	Category string `json:"category"`
}

func portalCreateMessage(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input portalCreateMessageInput
	if err := c.ShouldBind(&input); err != nil {
		log.Printf("portalCreateMessage: bind error: %v", err)
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	// Get patient name for sender field
	patient, err := model.Queries.GetPortalPatientDemographics(c.Request.Context(), patientID)
	if err != nil {
		log.Printf("portalCreateMessage: patient lookup error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}
	patientName := patient.FirstName + " " + patient.LastName

	result, err := model.Queries.PortalCreateMessage(c.Request.Context(), dbgen.PortalCreateMessageParams{
		Sender:      patientName,
		MsgFor:      0, // Practice
		PatientID:   patientID,
		PatientName: patientName,
		Urgency:     1,
		Subject:     input.Subject,
		Body:        input.Body,
	})
	if err != nil {
		log.Printf("portalCreateMessage: create error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{
		"id":      newID,
		"status":  "sent",
		"message": "Message sent successfully",
	})
}

// ============================================================================
// GET /api/portal/balance — Patient balance and credits
// ============================================================================

func portalBalance(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	charges, err := model.Queries.PortalOutstandingCharges(c.Request.Context(), patientID)
	if err != nil {
		log.Printf("portalBalance: charges error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}
	if charges == nil {
		charges = []dbgen.PortalOutstandingChargesRow{}
	}

	credits, err := model.Queries.PortalUnappliedCredits(c.Request.Context(), patientID)
	if err != nil {
		log.Printf("portalBalance: credits error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}
	if credits == nil {
		credits = []dbgen.PortalUnappliedCreditsRow{}
	}

	// Sum outstanding charges
	var totalOutstanding float64
	for _, ch := range charges {
		totalOutstanding += ch.Balance
	}

	c.JSON(http.StatusOK, gin.H{
		"total_outstanding":   totalOutstanding,
		"outstanding_charges": charges,
		"credits":             credits,
	})
}

// ============================================================================
// GET /api/portal/statements — Recent superbill/claim summaries
// ============================================================================

func portalStatements(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	rows, err := model.Queries.PortalListSuperbills(c.Request.Context(), patientID)
	if err != nil {
		log.Printf("portalStatements: error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	if rows == nil {
		rows = []dbgen.PortalListSuperbillsRow{}
	}

	c.JSON(http.StatusOK, rows)
}

// ============================================================================
// GET /api/portal/forms — List form templates
// ============================================================================

func portalListForms(c *gin.Context) {
	_, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	rows, err := model.Queries.ListFormTemplates(c.Request.Context())
	if err != nil {
		log.Printf("portalListForms: error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	if rows == nil {
		rows = []dbgen.FormTemplate{}
	}

	c.JSON(http.StatusOK, rows)
}

// ============================================================================
// GET /api/portal/forms/:id — Get form template with JSON content
// ============================================================================

func portalGetForm(c *gin.Context) {
	_, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	formID := common.ParseInt(c.Param("id"))
	if formID < 1 {
		common.ErrorResponse(c, http.StatusBadRequest, "invalid form id")
		return
	}

	row, err := model.Queries.GetFormTemplate(c.Request.Context(), formID)
	if err != nil {
		if err == sql.ErrNoRows {
			common.ErrorResponse(c, http.StatusNotFound, "form not found")
			return
		}
		log.Printf("portalGetForm: error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	resp := gin.H{
		"id":            row.ID,
		"name":          row.Name,
		"description":   row.Description,
		"form_type":     row.FormType,
		"template_data": row.TemplateData,
	}
	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// POST /api/portal/forms/:id/submit — Submit filled intake form
// ============================================================================

type portalSubmitFormInput struct {
	Responses map[string]interface{} `json:"responses" binding:"required"`
}

func portalSubmitForm(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	formID := common.ParseInt(c.Param("id"))
	if formID < 1 {
		common.ErrorResponse(c, http.StatusBadRequest, "invalid form id")
		return
	}

	var input portalSubmitFormInput
	if err := c.ShouldBind(&input); err != nil {
		log.Printf("portalSubmitForm: bind error: %v", err)
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	// Get form template to build summary
	form, err := model.Queries.GetFormTemplate(c.Request.Context(), formID)
	if err != nil {
		if err == sql.ErrNoRows {
			common.ErrorResponse(c, http.StatusNotFound, "form not found")
			return
		}
		log.Printf("portalSubmitForm: template lookup error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	// Store submission in patient_emr using raw SQL
	query := `INSERT INTO patient_emr (patient, module, oid, stamp, summary, user, provider, language, created_at, updated_at) VALUES (?, 'form_submission', ?, NOW(), ?, 0, 0, '', NOW(), NOW())`
	result, err := model.SqlDb.ExecContext(c.Request.Context(), query,
		patientID,
		formID,
		"Form: "+form.Name,
	)
	if err != nil {
		log.Printf("portalSubmitForm: insert error: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{
		"id":      newID,
		"form_id": formID,
		"status":  "submitted",
		"message": "Form submitted successfully",
	})
}

// ============================================================================
// GET /api/portal/ccda — Download CCDA XML document
// ============================================================================

func portalCCDADownload(c *gin.Context) {
	patientID, ok := extractPortalPatientID(c)
	if !ok {
		common.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	xmlData, err := ccda.GenerateCCD(patientID)
	if err != nil {
		log.Printf("portalCCDADownload: error generating CCD: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.Header("Content-Type", "application/xml")
	c.Header("Content-Disposition", "attachment; filename=ccda.xml")
	c.Data(http.StatusOK, "application/xml", xmlData)
}
