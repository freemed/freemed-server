package middleware

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

// AuditLog returns a Gin middleware that logs API actions to the audit_log table.
// action is the logical operation (e.g. "create", "update", "delete", "read").
// resourceType is the kind of resource being operated on (e.g. "patient", "encounter").
func AuditLog(action, resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		status := c.Writer.Status()
		success := status < 400

		var userID int64
		session, err := common.GetSession(c)
		if err == nil {
			userID = session.UserId
		}

		var patientID int64
		if pid := c.Param("id"); pid != "" {
			patientID = common.ParseInt(pid)
		}

		var resourceID int64
		if rid := c.Param("id"); rid != "" {
			resourceID = common.ParseInt(rid)
		}

		ip := c.ClientIP()
		ua := c.Request.UserAgent()
		details := fmt.Sprintf("method=%s path=%s status=%d", c.Request.Method, c.Request.URL.Path, status)

		_, err = model.Queries.InsertAuditLog(c.Request.Context(), dbgen.InsertAuditLogParams{
			Action:       action,
			UserID:       userID,
			PatientID:    patientID,
			ResourceType: resourceType,
			ResourceID:   resourceID,
			IpAddress:    ip,
			UserAgent:    ua,
			Details:      sql.NullString{String: details, Valid: true},
			Success:      success,
		})
		if err != nil {
			log.Printf("AuditLog: failed to insert audit entry: %s", err.Error())
		}
	}
}
