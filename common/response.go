package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse aborts the current request with a standardized JSON error body
// containing the HTTP status code and a human-readable message.
func ErrorResponse(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(code, gin.H{
		"code":    code,
		"message": message,
	})
}

// ErrorResponseFromError is a convenience wrapper that calls ErrorResponse with
// err.Error() as the message. Falls back to http.StatusText(code) when err is nil.
func ErrorResponseFromError(c *gin.Context, code int, err error) {
	msg := http.StatusText(code)
	if err != nil {
		msg = err.Error()
	}
	ErrorResponse(c, code, msg)
}
