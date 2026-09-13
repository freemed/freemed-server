package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// maxBodyTestRouter applies MaxBody(limit) to a handler that reads the whole
// body, mirroring the ingest handlers (dicom/hl7/era/ccda/eligibility).
func maxBodyTestRouter(limit int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/ingest", MaxBody(limit), func(c *gin.Context) {
		data, err := io.ReadAll(c.Request.Body)
		if err != nil {
			// This is what the real handlers do: surface the read error.
			if errors.As(err, new(*http.MaxBytesError)) {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "read": len(data)})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "read": len(data)})
	})
	return r
}

func decodeJSONBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not JSON: %v (%q)", err, w.Body.String())
	}
	return body
}

// TestMaxBodyRejectsOversizedContentLength covers the curl/browser/axios case:
// a declared Content-Length over the limit is refused before the handler runs.
func TestMaxBodyRejectsOversizedContentLength(t *testing.T) {
	const limit = 1024
	r := maxBodyTestRouter(limit)

	oversized := bytes.Repeat([]byte("A"), limit+1)
	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader(oversized))
	req.Header.Set("Content-Type", "application/octet-stream")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413 (body %q)", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q, want JSON", ct)
	}
	body := decodeJSONBody(t, w)
	if got, ok := body["code"].(float64); !ok || int(got) != http.StatusRequestEntityTooLarge {
		t.Errorf(`body["code"] = %v, want 413`, body["code"])
	}
}

// TestMaxBodyAllowsBodyUnderLimit guards against breaking normal requests.
func TestMaxBodyAllowsBodyUnderLimit(t *testing.T) {
	const limit = 1024
	r := maxBodyTestRouter(limit)

	req := httptest.NewRequest(http.MethodPost, "/ingest", strings.NewReader(strings.Repeat("B", 512)))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", w.Code, w.Body.String())
	}
	body := decodeJSONBody(t, w)
	if got, ok := body["read"].(float64); !ok || int(got) != 512 {
		t.Errorf(`body["read"] = %v, want 512`, body["read"])
	}
}

// TestMaxBodyCapsUnknownLengthBody covers chunked bodies (no Content-Length):
// the up-front check cannot see them, so http.MaxBytesReader must cap the read.
// This is the memory-exhaustion case measured at 30 MB -> 1.71 GB RSS.
func TestMaxBodyCapsUnknownLengthBody(t *testing.T) {
	const limit = 1024
	r := maxBodyTestRouter(limit)

	req := httptest.NewRequest(http.MethodPost, "/ingest", strings.NewReader(strings.Repeat("C", limit*8)))
	req.ContentLength = -1 // simulate Transfer-Encoding: chunked
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// The handler sees the read fail; the response is 400 with the driver-level
	// message (documented behaviour for chunked bodies) — what matters is that
	// the handler could not read past the cap.
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body %q)", w.Code, w.Body.String())
	}
	body := decodeJSONBody(t, w)
	read, _ := body["read"].(float64)
	if int(read) > limit {
		t.Errorf("handler read %v bytes, want at most %d", read, limit)
	}
	if msg, _ := body["message"].(string); !strings.Contains(msg, "too large") {
		t.Errorf("message = %q, want the request-body-too-large error", msg)
	}
}

// TestMaxBodyZeroLimitDisables documents the limit<=0 escape hatch.
func TestMaxBodyZeroLimitDisables(t *testing.T) {
	r := maxBodyTestRouter(0)
	req := httptest.NewRequest(http.MethodPost, "/ingest", strings.NewReader(strings.Repeat("D", 4096)))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 when the limit is disabled", w.Code)
	}
}
