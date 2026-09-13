package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	ginjwt "github.com/appleboy/gin-jwt/v2"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/gin-gonic/gin"
)

// sessionContext builds a gin context carrying the JWT claim payload that
// gin-jwt's middleware installs on an authenticated request, so
// common.GetSession resolves to the given staff user. userID == 0 leaves the
// payload unset, which is the unauthenticated case.
func sessionContext(t *testing.T, userID int64, method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, reader)
	c.Request.Header.Set("Content-Type", "application/json")
	if userID != 0 {
		c.Set("JWT_PAYLOAD", ginjwt.MapClaims{"id": float64(userID), "user_type": "admin"})
	}
	return c, w
}

// rowsResult is the minimal sql.Result that reports a fixed affected-row count,
// so a delete handler can be exercised without a database.
type rowsResult int64

func (r rowsResult) LastInsertId() (int64, error) { return 0, nil }
func (r rowsResult) RowsAffected() (int64, error) { return int64(r), nil }

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var out map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("response is not a JSON object: %v (body %q)", err, w.Body.String())
	}
	return out
}

// ============================================================================
// M5.1 — POST /api/messages/delete
// ============================================================================

// stubMessageDelete replaces the delete seam and records every call.
func stubMessageDelete(t *testing.T, affected int64) *[]dbgen.DeleteMessagesForUserParams {
	t.Helper()
	orig := messagesDeleteRows
	calls := &[]dbgen.DeleteMessagesForUserParams{}
	messagesDeleteRows = func(ctx context.Context, arg dbgen.DeleteMessagesForUserParams) (sql.Result, error) {
		*calls = append(*calls, arg)
		return rowsResult(affected), nil
	}
	t.Cleanup(func() { messagesDeleteRows = orig })
	return calls
}

// TestMessagesDeleteIsScopedToTheSessionUser is the M5 regression test for the
// mass-delete hole: the handler used to hand the raw request-body ids to an
// unscoped `DELETE FROM messages WHERE id IN (...)`, so any authenticated user
// could delete another user's secure messages. The ids must now travel with the
// session user, which the SQL uses as `msgfor`.
func TestMessagesDeleteIsScopedToTheSessionUser(t *testing.T) {
	calls := stubMessageDelete(t, 1)

	c, w := sessionContext(t, 7, http.MethodPost, "/api/messages/delete", `{"ids":[11,12]}`)
	messagesDelete(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", w.Code, w.Body.String())
	}
	if len(*calls) != 1 {
		t.Fatalf("delete calls = %d, want 1", len(*calls))
	}
	got := (*calls)[0]
	if got.UserID != 7 {
		t.Errorf("delete scoped to user %d, want the session user 7 — a caller can name any id in the body", got.UserID)
	}
	if !reflect.DeepEqual(got.Ids, []int64{11, 12}) {
		t.Errorf("delete ids = %v, want [11 12]", got.Ids)
	}

	body := decodeBody(t, w)
	if body["deleted"] != float64(1) {
		t.Errorf("deleted = %v, want 1 (the rows actually removed)", body["deleted"])
	}
	if body["requested"] != float64(2) {
		t.Errorf("requested = %v, want 2", body["requested"])
	}
}

// TestMessagesDeleteReportsSkippedNonOwnedIDs pins the chosen semantics: a
// non-owned id is silently skipped (the scoped DELETE simply does not match it)
// and the response is honest about how many rows were really removed, instead
// of reporting success for the whole batch.
func TestMessagesDeleteReportsSkippedNonOwnedIDs(t *testing.T) {
	stubMessageDelete(t, 0)

	c, w := sessionContext(t, 7, http.MethodPost, "/api/messages/delete", `{"ids":[999]}`)
	messagesDelete(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	if body["deleted"] != float64(0) {
		t.Errorf("deleted = %v, want 0 — an id owned by another user must not be deleted", body["deleted"])
	}
	if body["requested"] != float64(1) {
		t.Errorf("requested = %v, want 1", body["requested"])
	}
}

// TestMessagesDeleteRequiresASession covers the missing session check: the
// handler had none at all, and it must fail closed before touching the query.
func TestMessagesDeleteRequiresASession(t *testing.T) {
	calls := stubMessageDelete(t, 1)

	c, w := sessionContext(t, 0, http.MethodPost, "/api/messages/delete", `{"ids":[11]}`)
	messagesDelete(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 (body %q)", w.Code, w.Body.String())
	}
	if len(*calls) != 0 {
		t.Fatalf("delete was called %d times without a session; want 0", len(*calls))
	}
}

func TestMessagesDeleteRejectsAnEmptyIDList(t *testing.T) {
	calls := stubMessageDelete(t, 1)

	c, w := sessionContext(t, 7, http.MethodPost, "/api/messages/delete", `{"ids":[]}`)
	messagesDelete(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body %q)", w.Code, w.Body.String())
	}
	if len(*calls) != 0 {
		t.Fatalf("delete was called %d times for an empty id list; want 0", len(*calls))
	}
}

// ============================================================================
// M5.2 — GET /api/notifications/from
// ============================================================================

func stubNotificationsFrom(t *testing.T) *[]dbgen.NotificationsFromTimestampParams {
	t.Helper()
	orig := notificationsFromTimestampRows
	calls := &[]dbgen.NotificationsFromTimestampParams{}
	notificationsFromTimestampRows = func(ctx context.Context, arg dbgen.NotificationsFromTimestampParams) ([]dbgen.Systemnotification, error) {
		*calls = append(*calls, arg)
		return []dbgen.Systemnotification{}, nil
	}
	t.Cleanup(func() { notificationsFromTimestampRows = orig })
	return calls
}

// TestNotificationsFromScopesToTheSessionUser is the M5 regression test for the
// poll endpoint. NotificationsFromTimestamp had no nuser predicate at all, and
// systemnotification carries npatient, so the response covered every other
// user's and every other patient's rows.
func TestNotificationsFromScopesToTheSessionUser(t *testing.T) {
	calls := stubNotificationsFrom(t)

	c, w := sessionContext(t, 7, http.MethodGet, "/api/notifications/from?timestamp=2026-03-04T05:06:07Z", "")
	notificationsFromTimestamp(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", w.Code, w.Body.String())
	}
	if len(*calls) != 1 {
		t.Fatalf("query calls = %d, want 1", len(*calls))
	}
	got := (*calls)[0]
	if got.UserID != 7 {
		t.Errorf("nuser filter = %d, want the session user 7", got.UserID)
	}
	want := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)
	if !got.Since.Equal(want) {
		t.Errorf("since = %s, want %s", got.Since, want)
	}
}

func TestNotificationsFromRequiresASession(t *testing.T) {
	calls := stubNotificationsFrom(t)

	c, w := sessionContext(t, 0, http.MethodGet, "/api/notifications/from?timestamp=2026-03-04T05:06:07Z", "")
	notificationsFromTimestamp(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 (body %q)", w.Code, w.Body.String())
	}
	if len(*calls) != 0 {
		t.Fatalf("query was called %d times without a session; want 0", len(*calls))
	}
}

func TestNotificationsFromRejectsABadTimestamp(t *testing.T) {
	calls := stubNotificationsFrom(t)

	c, w := sessionContext(t, 7, http.MethodGet, "/api/notifications/from?timestamp=nonsense", "")
	notificationsFromTimestamp(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body %q)", w.Code, w.Body.String())
	}
	if len(*calls) != 0 {
		t.Fatalf("query was called %d times for an unparseable timestamp; want 0", len(*calls))
	}
}

// ============================================================================
// M5.3 — GET /api/search
// ============================================================================

// stubSearchQueries replaces all three concurrent search legs; the message leg
// is the one that must be scoped to the session user.
func stubSearchQueries(t *testing.T) *[]dbgen.SearchMessagesParams {
	t.Helper()
	origPatients, origMessages, origAppointments := searchPatientsQuery, searchMessagesQuery, searchAppointmentsQuery
	calls := &[]dbgen.SearchMessagesParams{}
	searchPatientsQuery = func(ctx context.Context, arg dbgen.SearchPatientsParams) ([]dbgen.SearchPatientsRow, error) {
		return []dbgen.SearchPatientsRow{}, nil
	}
	searchMessagesQuery = func(ctx context.Context, arg dbgen.SearchMessagesParams) ([]dbgen.SearchMessagesRow, error) {
		*calls = append(*calls, arg)
		return []dbgen.SearchMessagesRow{}, nil
	}
	searchAppointmentsQuery = func(ctx context.Context, query interface{}) ([]dbgen.SearchAppointmentsRow, error) {
		return []dbgen.SearchAppointmentsRow{}, nil
	}
	t.Cleanup(func() {
		searchPatientsQuery, searchMessagesQuery, searchAppointmentsQuery = origPatients, origMessages, origAppointments
	})
	return calls
}

// TestSearchScopesMessagesToTheSessionUser is the M5 regression test for global
// search: SearchMessages matched on msgsubject alone, so it returned every
// message subject in the system to any authenticated caller.
func TestSearchScopesMessagesToTheSessionUser(t *testing.T) {
	calls := stubSearchQueries(t)

	c, w := sessionContext(t, 9, http.MethodGet, "/api/search?q=lab", "")
	search(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", w.Code, w.Body.String())
	}
	if len(*calls) != 1 {
		t.Fatalf("message search calls = %d, want 1", len(*calls))
	}
	got := (*calls)[0]
	if got.UserID != 9 {
		t.Errorf("message search scoped to user %d, want the session user 9 — the subject filter alone leaks other users' messages", got.UserID)
	}
	if got.Query != "lab" {
		t.Errorf("message search query = %v, want lab", got.Query)
	}
}

func TestSearchRequiresASession(t *testing.T) {
	calls := stubSearchQueries(t)

	c, w := sessionContext(t, 0, http.MethodGet, "/api/search?q=lab", "")
	search(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 (body %q)", w.Code, w.Body.String())
	}
	if len(*calls) != 0 {
		t.Fatalf("message search ran %d times without a session; want 0", len(*calls))
	}
}
