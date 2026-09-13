package common

import (
	"errors"
	"fmt"
	"log"
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

// SafeMessageError wraps an error whose text is deliberately written for the
// client and is therefore safe to echo even on a 5xx response.
type SafeMessageError struct {
	msg string
}

// Error implements the error interface.
func (e *SafeMessageError) Error() string { return e.msg }

// SafeMessage returns the client-safe text of the error.
func (e *SafeMessageError) SafeMessage() string { return e.msg }

// SafeMessageErrorf builds a SafeMessageError from a format string.
func SafeMessageErrorf(format string, args ...interface{}) error {
	return &SafeMessageError{msg: fmt.Sprintf(format, args...)}
}

// genericInternalMessage is what a client sees for a 5xx whose cause is not an
// explicitly client-safe error.
const genericInternalMessage = "internal server error"

// ErrorResponseFromError is a convenience wrapper that calls ErrorResponse with
// a message derived from err. Falls back to http.StatusText(code) when err is nil.
//
// Client errors (4xx) keep err.Error(): those messages are written for the user
// ("invalid id", "form not found", bind failures) and carry no server detail.
//
// Server errors (5xx) do NOT: err.Error() on a 5xx is routinely a driver or SQL
// message — `Error 1146 (42S02): Table 'freemed.audit_log' doesn't exist` —
// which leaked schema names, driver internals and query text to unauthenticated
// reconnaissance. The real error is logged server-side with the request id
// instead, and the client gets a generic message. A handler that *does* have a
// safe, useful message for a 5xx can pass common.SafeMessageErrorf(...) (or any
// error implementing SafeMessage() string) and it will be echoed verbatim.
func ErrorResponseFromError(c *gin.Context, code int, err error) {
	if err == nil {
		ErrorResponse(c, code, http.StatusText(code))
		return
	}

	if code < http.StatusInternalServerError {
		ErrorResponse(c, code, err.Error())
		return
	}

	var safe interface{ SafeMessage() string }
	if errors.As(err, &safe) {
		ErrorResponse(c, code, safe.SafeMessage())
		return
	}

	requestID := ""
	method := ""
	path := ""
	if c != nil {
		requestID = c.GetString("request_id")
		if c.Request != nil {
			method = c.Request.Method
			path = c.Request.URL.Path
		}
	}
	// Log the real cause server-side so the detail is not lost with the
	// client-visible text.
	log.Printf("ErrorResponseFromError: %d %s %s (request_id=%q): %v", code, method, path, requestID, err)

	msg := genericInternalMessage
	if requestID != "" {
		msg = fmt.Sprintf("%s (request id %s)", genericInternalMessage, requestID)
	}
	ErrorResponse(c, code, msg)
}
