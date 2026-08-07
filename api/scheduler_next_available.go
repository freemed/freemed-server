package api

import (
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

// nextAvailableInput is the JSON body for POST /api/scheduler/next-available.
type nextAvailableInput struct {
	ProviderID int64  `json:"provider_id" binding:"required"`
	Date       string `json:"date"        binding:"required"`
	Duration   int64  `json:"duration"    binding:"required"`
}

// nextAvailableResponse is the JSON response from the next-available endpoint.
type nextAvailableResponse struct {
	Available bool  `json:"available"`
	Hour      int64 `json:"hour"`
	Minute    int64 `json:"minute"`
}

const (
	clinicStartHour   = 6
	clinicEndHour     = 18
	slotIncrementMins = 15
)

// schedulerNextAvailable handles POST /api/scheduler/next-available.
// It finds the first available time slot for a given provider on a given date
// that can accommodate the requested duration without overlapping existing appointments.
func schedulerNextAvailable(c *gin.Context) {
	var in nextAvailableInput
	if err := c.ShouldBind(&in); err != nil {
		log.Printf("schedulerNextAvailable: bind error: %v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	calDate, err := common.ParseDate(in.Date)
	if err != nil {
		log.Printf("schedulerNextAvailable: date parse error: %v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Fetch all existing appointments for this provider on this date.
	appts, err := model.Queries.FindNextAvailable(c.Request.Context(), dbgen.FindNextAvailableParams{
		ProviderID: in.ProviderID,
		ReqDate:    calDate,
	})
	if err != nil {
		log.Printf("schedulerNextAvailable: query error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Iterate over time slots from clinicStartHour to clinicEndHour
	// in slotIncrementMins increments.
	for hour := int64(clinicStartHour); hour < int64(clinicEndHour); hour++ {
		for minute := int64(0); minute < 60; minute += int64(slotIncrementMins) {
			candidateStart := hour*60 + minute
			candidateEnd := candidateStart + in.Duration

			// Don't go past clinic end.
			if candidateEnd > int64(clinicEndHour)*60 {
				break
			}

			// Check for conflicts with existing appointments.
			conflict := false
			for _, a := range appts {
				apptStart := a.Calhour*60 + a.Calminute
				apptEnd := apptStart + a.Calduration

				// Overlap: candidate_start < appt_end AND candidate_end > appt_start
				if candidateStart < apptEnd && candidateEnd > apptStart {
					conflict = true
					break
				}
			}

			if !conflict {
				c.JSON(http.StatusOK, nextAvailableResponse{
					Available: true,
					Hour:      hour,
					Minute:    minute,
				})
				return
			}
		}
	}

	// No slot found.
	c.JSON(http.StatusOK, nextAvailableResponse{
		Available: false,
	})
}
