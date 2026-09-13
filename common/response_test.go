package common

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// runErrorHandler drives ErrorResponseFromError through a real gin context so
// the assertions cover the middleware/context plumbing, not just the helper.
func runErrorHandler(t *testing.T, code int, err error, requestID string) (*httptest.ResponseRecorder, map[string]interface{}) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		if requestID != "" {
			c.Set("request_id", requestID)
		}
		c.Next()
	})
	r.GET("/boom", func(c *gin.Context) {
		ErrorResponseFromError(c, code, err)
	})

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var body map[string]interface{}
	if w.Body.Len() > 0 {
		if jerr := json.Unmarshal(w.Body.Bytes(), &body); jerr != nil {
			t.Fatalf("response body is not JSON: %v (%q)", jerr, w.Body.String())
		}
	}
	return w, body
}

// TestErrorResponseFromErrorSanitizesServerErrors is the M6 regression test:
// a driver/SQL error must never reach the client verbatim on a 5xx.
func TestErrorResponseFromErrorSanitizesServerErrors(t *testing.T) {
	dbErr := errors.New("Error 1146 (42S02): Table 'freemed.audit_log' doesn't exist")

	w, body := runErrorHandler(t, http.StatusInternalServerError, dbErr, "req-123")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
	// Existing JSON shape must be preserved: {"code": N, "message": "..."}
	if got, ok := body["code"].(float64); !ok || int(got) != http.StatusInternalServerError {
		t.Errorf(`body["code"] = %v, want 500`, body["code"])
	}
	msg, ok := body["message"].(string)
	if !ok {
		t.Fatalf(`body["message"] is not a string: %v`, body["message"])
	}
	if strings.Contains(msg, "1146") || strings.Contains(msg, "audit_log") || strings.Contains(msg, "Error ") {
		t.Errorf("client message leaked the internal error: %q", msg)
	}
	if msg != genericInternalMessage+" (request id req-123)" {
		t.Errorf("message = %q, want the generic message plus request id", msg)
	}
	// The actionable detail must be recoverable from the request id.
	if !strings.Contains(msg, "req-123") {
		t.Errorf("message = %q, want it to carry the request id for log correlation", msg)
	}
}

// TestErrorResponseFromErrorKeepsClientErrorMessages guards the 4xx contract:
// client errors are written for the user and must be echoed unchanged.
func TestErrorResponseFromErrorKeepsClientErrorMessages(t *testing.T) {
	w, body := runErrorHandler(t, http.StatusBadRequest, errors.New("template_id is required"), "")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	if body["message"] != "template_id is required" {
		t.Errorf("message = %v, want the original client error text", body["message"])
	}
}

// TestErrorResponseFromErrorNilErrorUsesStatusText covers the documented
// fallback.
func TestErrorResponseFromErrorNilErrorUsesStatusText(t *testing.T) {
	_, body := runErrorHandler(t, http.StatusNotFound, nil, "")
	if body["message"] != http.StatusText(http.StatusNotFound) {
		t.Errorf("message = %v, want %q", body["message"], http.StatusText(http.StatusNotFound))
	}
}

// TestErrorResponseFromErrorHonoursSafeMessage proves the opt-in escape hatch:
// a handler that has a deliberately user-facing 5xx message can still pass it
// through, without weakening the default.
func TestErrorResponseFromErrorHonoursSafeMessage(t *testing.T) {
	w, body := runErrorHandler(t, http.StatusServiceUnavailable,
		SafeMessageErrorf("e-prescribing is not configured: set ncpdp.sender-id in config.yml"), "")

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", w.Code)
	}
	if body["message"] != "e-prescribing is not configured: set ncpdp.sender-id in config.yml" {
		t.Errorf("message = %v, want the wrapped safe message", body["message"])
	}
}

// TestErrorResponseShapeUnchanged pins the JSON contract clients depend on.
func TestErrorResponseShapeUnchanged(t *testing.T) {
	w, body := runErrorHandler(t, http.StatusForbidden, errors.New("nope"), "")
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
	if len(body) != 2 {
		t.Errorf("body has %d keys (%v), want exactly code+message", len(body), body)
	}
}
