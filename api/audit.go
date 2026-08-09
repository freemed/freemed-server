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

func init() {
	common.ApiMap["admin"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/audit-log", common.RequireRole("admin"), listAuditLog)
		},
	}
}

func listAuditLog(c *gin.Context) {
	offset := int32(0)
	limit := int32(50)

	if v := c.Query("offset"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			offset = int32(n)
		}
	}
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil && n > 0 && n <= 200 {
			limit = int32(n)
		}
	}

	var entries interface{}
	var err error

	if userIDStr := c.Query("user"); userIDStr != "" {
		userID := common.ParseInt(userIDStr)
		entries, err = model.Queries.ListAuditLogByUser(c.Request.Context(), dbgen.ListAuditLogByUserParams{
			UserID: userID,
			Limit:  limit,
			Offset: offset,
		})
	} else if patientIDStr := c.Query("patient"); patientIDStr != "" {
		patientID := common.ParseInt(patientIDStr)
		entries, err = model.Queries.ListAuditLogByPatient(c.Request.Context(), dbgen.ListAuditLogByPatientParams{
			PatientID: patientID,
			Limit:     limit,
			Offset:    offset,
		})
	} else {
		entries, err = model.Queries.ListAuditLog(c.Request.Context(), dbgen.ListAuditLogParams{
			Limit:  limit,
			Offset: offset,
		})
	}

	if err != nil {
		log.Printf("listAuditLog: %s", err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"audit_log": entries,
		"offset":    offset,
		"limit":     limit,
	})
}
