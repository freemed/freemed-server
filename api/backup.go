package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/freemed/freemed-server/common"
	"github.com/gin-gonic/gin"
)

// verifyBackup handles GET /api/verify-backup
// Checks the existence, size, and age of the database backup file.
// The backup path is configured via the FREEMED_BACKUP_PATH env var or defaults
// to backups/freemed-backup.sql in the configured base path.
func verifyBackup(c *gin.Context) {
	backupPath := os.Getenv("FREEMED_BACKUP_PATH")
	if backupPath == "" {
		// Try a configured path or fall back to a reasonable default
		backupPath = fmt.Sprintf("%s/backups/freemed-backup.sql", getBasePath())
	}

	result := gin.H{
		"path":    backupPath,
		"exists":  false,
		"status":  "missing",
	}

	info, err := os.Stat(backupPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusOK, result)
			return
		}
		log.Printf("verifyBackup: stat error: %s", err.Error())
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	result["exists"] = true
	result["size_bytes"] = info.Size()
	result["last_modified"] = info.ModTime().Format(time.RFC3339)

	ageHours := time.Since(info.ModTime()).Hours()
	result["age_hours"] = fmt.Sprintf("%.1f", ageHours)

	// Determine status based on age threshold from env or default 24h
	thresholdHours := 24.0
	if v := os.Getenv("FREEMED_BACKUP_MAX_AGE_HOURS"); v != "" {
		if n, err := time.ParseDuration(v); err == nil {
			thresholdHours = n.Hours()
		}
	}

	if ageHours > thresholdHours {
		result["status"] = "stale"
	} else {
		result["status"] = "ok"
	}

	c.JSON(http.StatusOK, result)
}

// getBasePath returns the configured base path, defaulting to ".".
func getBasePath() string {
	base := os.Getenv("FREEMED_BASE_PATH")
	if base != "" {
		return base
	}
	return "."
}
