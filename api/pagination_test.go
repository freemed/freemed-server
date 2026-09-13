package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/freemed/freemed-server/dbgen"
	"github.com/gin-gonic/gin"
)

// TestPageParamsClamps covers the single bound every M7 list endpoint shares.
// Before this the handlers returned every row in the table; a caller can also
// ask for `?limit=1000000` explicitly, so the ceiling has to be enforced here
// and not only in SQL.
func TestPageParamsClamps(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		target     string
		wantOffset int32
		wantLimit  int32
	}{
		{"no parameters uses the defaults", "/api/x", 0, 50},
		{"explicit page", "/api/x?offset=10&limit=5", 10, 5},
		{"negative offset is repaired", "/api/x?offset=-5&limit=5", 0, 5},
		{"negative limit falls back to the default", "/api/x?offset=3&limit=-1", 3, 50},
		{"zero limit falls back to the default", "/api/x?limit=0", 0, 50},
		{"limit above the cap is clamped", "/api/x?limit=1000000", 0, int32(maxListLimit)},
		{"exactly the cap is allowed", "/api/x?limit=200", 0, int32(maxListLimit)},
		{"offset above the cap is clamped", "/api/x?offset=99999999999", int32(maxListOffset), 50},
		{"unparseable values fall back to the defaults", "/api/x?offset=abc&limit=xyz", 0, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodGet, tt.target, nil)

			offset, limit := pageParams(c)
			if offset != tt.wantOffset || limit != tt.wantLimit {
				t.Errorf("pageParams(%s) = (%d, %d), want (%d, %d)", tt.target, offset, limit, tt.wantOffset, tt.wantLimit)
			}
		})
	}
}

// docEnvelope is the pagination envelope the two document inboxes now return.
type docEnvelope struct {
	Data   json.RawMessage `json:"data"`
	Total  int64           `json:"total"`
	Offset int32           `json:"offset"`
	Limit  int32           `json:"limit"`
}

// callHandler runs a handler through a real gin test context and returns the
// recorder, so a test can assert on both the status and the body.
func callHandler(t *testing.T, method, target string, h gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, nil)
	h(c)
	return w
}

// TestUnfiledDocsListIsBoundedAndEnveloped covers M7 for
// GET /api/documents/unfiled: the clamped page bound reaches the query and the
// response is the standard envelope rather than a bare array of every row in
// the table.
func TestUnfiledDocsListIsBoundedAndEnveloped(t *testing.T) {
	origPage, origTotal := unfiledDocsPageQuery, unfiledDocsTotalQuery
	var page dbgen.ListUnfiledDocsParams
	unfiledDocsPageQuery = func(ctx context.Context, arg dbgen.ListUnfiledDocsParams) ([]dbgen.UnfiledDoc, error) {
		page = arg
		return []dbgen.UnfiledDoc{{ID: 1}}, nil
	}
	unfiledDocsTotalQuery = func(ctx context.Context) (int64, error) { return 4242, nil }
	t.Cleanup(func() { unfiledDocsPageQuery, unfiledDocsTotalQuery = origPage, origTotal })

	w := callHandler(t, http.MethodGet, "/api/documents/unfiled?offset=-3&limit=1000000", unfiledDocsList)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", w.Code, w.Body.String())
	}
	if page.Limit != int32(maxListLimit) || page.Offset != 0 {
		t.Errorf("query page = (offset %d, limit %d), want (0, %d)", page.Offset, page.Limit, maxListLimit)
	}

	var env docEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("body is not the pagination envelope: %v (body %q)", err, w.Body.String())
	}
	if env.Total != 4242 {
		t.Errorf("total = %d, want 4242 (the count, not the page length)", env.Total)
	}
	if env.Offset != 0 || env.Limit != int32(maxListLimit) {
		t.Errorf("envelope page = (offset %d, limit %d), want (0, %d)", env.Offset, env.Limit, maxListLimit)
	}
}

// TestUnfiledDocsListEmptyPageIsAnArray pins the empty-page shape: `data` must
// serialise as [] rather than null, so a client can iterate it unconditionally.
func TestUnfiledDocsListEmptyPageIsAnArray(t *testing.T) {
	origPage, origTotal := unfiledDocsPageQuery, unfiledDocsTotalQuery
	unfiledDocsPageQuery = func(ctx context.Context, arg dbgen.ListUnfiledDocsParams) ([]dbgen.UnfiledDoc, error) {
		return nil, nil
	}
	unfiledDocsTotalQuery = func(ctx context.Context) (int64, error) { return 0, nil }
	t.Cleanup(func() { unfiledDocsPageQuery, unfiledDocsTotalQuery = origPage, origTotal })

	w := callHandler(t, http.MethodGet, "/api/documents/unfiled", unfiledDocsList)
	if body := w.Body.String(); !strings.Contains(body, `"data":[]`) {
		t.Errorf("empty page body = %s, want data to be an empty array", body)
	}
}

// TestUnreadDocsListIsBoundedAndEnveloped is the same contract for
// GET /api/documents/unread.
func TestUnreadDocsListIsBoundedAndEnveloped(t *testing.T) {
	origPage, origTotal := unreadDocsPageQuery, unreadDocsTotalQuery
	var page dbgen.ListUnreadDocsParams
	unreadDocsPageQuery = func(ctx context.Context, arg dbgen.ListUnreadDocsParams) ([]dbgen.UnreadDoc, error) {
		page = arg
		return []dbgen.UnreadDoc{{ID: 2}}, nil
	}
	unreadDocsTotalQuery = func(ctx context.Context) (int64, error) { return 7, nil }
	t.Cleanup(func() { unreadDocsPageQuery, unreadDocsTotalQuery = origPage, origTotal })

	w := callHandler(t, http.MethodGet, "/api/documents/unread?offset=20&limit=5", unreadDocsList)

	if page.Limit != 5 || page.Offset != 20 {
		t.Errorf("query page = (offset %d, limit %d), want (20, 5)", page.Offset, page.Limit)
	}

	var env docEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("body is not the pagination envelope: %v (body %q)", err, w.Body.String())
	}
	if env.Total != 7 {
		t.Errorf("total = %d, want 7", env.Total)
	}
	if env.Offset != 20 || env.Limit != 5 {
		t.Errorf("envelope page = (offset %d, limit %d), want (20, 5)", env.Offset, env.Limit)
	}
}
