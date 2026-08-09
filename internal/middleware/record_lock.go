package middleware

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
)

// CheckLock returns middleware that checks if a record is locked by another user.
// tableName: the database table name to check against.
// paramName: the URL/query param that holds the record ID (e.g. "id").
//
// If the record is locked by another user, returns 423 Locked.
// If the record is NOT locked (first access), it sets c.Set("record_lock_acquired", true).
func CheckLock(tableName, paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		recordID := common.ParseInt(c.Param(paramName))
		if recordID == 0 {
			// Also try query param
			recordID = common.ParseInt(c.Query(paramName))
		}
		if recordID == 0 {
			c.Next()
			return
		}

		session, err := common.GetSession(c)
		if err != nil {
			common.ErrorResponseFromError(c, http.StatusUnauthorized, err)
			return
		}

		lock, err := model.Queries.CheckRecordLock(c.Request.Context(), dbgen.CheckRecordLockParams{
			TableName: tableName,
			RecordID:  recordID,
		})
		if err == nil {
			// Lock exists — check if it's the same user
			if lock.UserID != session.UserId {
				c.AbortWithStatusJSON(http.StatusLocked, gin.H{
					"code":    423,
					"message": "Record is currently locked by another user",
					"locked_by": lock.UserID,
				})
				return
			}
		} else if err != sql.ErrNoRows {
			log.Printf("CheckLock: error checking lock: %s", err.Error())
			// Non-fatal; allow the request to proceed
		}

		c.Set("record_lock_acquired", true)
		c.Next()
	}
}
