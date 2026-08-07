package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

// schedulerListBlockedSlots handles GET /api/scheduler/blocks?date=YYYY-MM-DD&provider_id=N
func schedulerListBlockedSlots(c *gin.Context) {
	date, err := common.ParseDate(c.Query("date"))
	if err != nil {
		log.Printf("schedulerListBlockedSlots: bad date: %v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	providerID, err := strconv.ParseInt(c.Query("provider_id"), 10, 64)
	if err != nil || providerID < 1 {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	rows, err := model.Queries.ListBlockedSlots(c.Request.Context(), dbgen.ListBlockedSlotsParams{
		Sbsdate:     date,
		Sbsprovider: providerID,
	})
	if err != nil {
		log.Printf("schedulerListBlockedSlots: ERROR: %s", err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, rows)
}

// schedulerCreateBlockedSlot handles POST /api/scheduler/blocks
func schedulerCreateBlockedSlot(c *gin.Context) {
	session, err := common.GetSession(c)
	if err != nil {
		log.Printf("schedulerCreateBlockedSlot: failed to get session: %v", err)
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	var input struct {
		Date     string `json:"date"     binding:"required"`
		Hour     int64  `json:"hour"     binding:"required"`
		Minute   int64  `json:"minute"   binding:"required"`
		Duration int64  `json:"duration" binding:"required"`
		Provider int64  `json:"provider" binding:"required"`
		Reason   string `json:"reason"`
	}
	if err := c.ShouldBind(&input); err != nil {
		log.Printf("schedulerCreateBlockedSlot: bind error: %v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	sbsDate, err := common.ParseDate(input.Date)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	result, err := model.Queries.CreateBlockedSlot(c.Request.Context(), dbgen.CreateBlockedSlotParams{
		Sbsdate:     sbsDate,
		Sbshour:     input.Hour,
		Sbsminute:   input.Minute,
		Sbsduration: input.Duration,
		Sbsprovider: input.Provider,
		Sbsreason:   input.Reason,
		User:        session.UserId,
	})
	if err != nil {
		log.Printf("schedulerCreateBlockedSlot: ERROR: %s", err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": newID})
}

// schedulerDeleteBlockedSlot handles DELETE /api/scheduler/blocks/:id
func schedulerDeleteBlockedSlot(c *gin.Context) {
	id := common.ParseInt(c.Param("id"))
	if id < 1 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	err := model.Queries.DeleteBlockedSlot(c.Request.Context(), id)
	if err != nil {
		log.Printf("schedulerDeleteBlockedSlot(%d): ERROR: %s", id, err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
