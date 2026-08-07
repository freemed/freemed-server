package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID returns a Gin middleware that ensures every request has a unique
// request ID. It reads the X-Request-ID header from the incoming request and
// reuses it when present; otherwise it generates a new UUID. The ID is set on
// the Gin context as "request_id" and echoed back in the X-Request-ID response
// header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}
