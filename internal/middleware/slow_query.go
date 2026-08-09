package middleware

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// SlowQueryLog returns a Gin middleware that logs requests exceeding the given
// duration threshold. thresholdMs is in milliseconds.
func SlowQueryLog(thresholdMs int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		ms := duration.Milliseconds()
		if ms > thresholdMs {
			log.Printf("SLOW QUERY [%dms] %s %s", ms, c.Request.Method, c.Request.URL.Path)
			// Also log as structured message for downstream parsing
			log.Printf("slow_query method=%s path=%s duration_ms=%d status=%d",
				c.Request.Method, c.Request.URL.Path, ms, c.Writer.Status())
		}
	}
}

// SlowQueryLogf is like SlowQueryLog but accepts a format string for additional
// context. Use when you want to include resource-specific info.
func SlowQueryLogf(thresholdMs int64, format string, args ...interface{}) gin.HandlerFunc {
	prefix := fmt.Sprintf(format, args...)
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		ms := duration.Milliseconds()
		if ms > thresholdMs {
			log.Printf("SLOW QUERY [%dms] %s %s %s", ms, c.Request.Method, c.Request.URL.Path, prefix)
		}
	}
}
