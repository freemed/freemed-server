package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/freemed/freemed-server/dbgen"
	"github.com/gin-gonic/gin"
)

// newClientParamRequest builds a gin context carrying the :client_id path
// parameter the SMART client metadata handler reads.
func newClientParamRequest(clientID string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/smart/clients/"+clientID+"/public", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "client_id", Value: clientID}}
	return c, w
}

// TestGetFhirClientPublicNeverLeaksSecretHash is the point of this handler: it
// exists so the consent screen can name the app, and `SELECT *` loads the stored
// bcrypt credential hash into the struct it reads from, so the projection — not
// the struct — is what must be serialized.
func TestGetFhirClientPublicNeverLeaksSecretHash(t *testing.T) {
	const hash = "$2a$10$SUPERSECRETHASHVALUETHATMUSTNEVERLEAVE"
	stubClient(t, dbgen.FhirClient{
		ClientID:         "client-1",
		ClientName:       "Acme Health",
		Scopes:           "launch patient/*.read",
		IsConfidential:   true,
		ClientSecretHash: hash,
	}, nil)

	c, w := newClientParamRequest("client-1")
	getFhirClientPublic(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), hash) {
		t.Fatalf("response leaked the client secret hash: %q", w.Body.String())
	}

	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("response is not JSON: %v (body %q)", err, w.Body.String())
	}
	if _, ok := out["client_secret_hash"]; ok {
		t.Error("response must not carry a client_secret_hash field")
	}
	if _, ok := out["redirect_uris"]; ok {
		t.Error("response must stay minimal: redirect_uris is not needed by the consent screen")
	}
	if out["client_name"] != "Acme Health" {
		t.Errorf("client_name = %v, want Acme Health", out["client_name"])
	}
	if out["scopes"] != "launch patient/*.read" {
		t.Errorf("scopes = %v", out["scopes"])
	}
	if out["is_confidential"] != true {
		t.Errorf("is_confidential = %v, want true", out["is_confidential"])
	}
}

// TestGetFhirClientPublicUnknownClient: an unknown or deactivated client is a
// 404 (GetFhirClientByID filters on active = 1), which is what the consent
// screen renders as "cannot be authorized".
func TestGetFhirClientPublicUnknownClient(t *testing.T) {
	stubClient(t, dbgen.FhirClient{ClientID: "client-1"}, nil)

	c, w := newClientParamRequest("client-2")
	getFhirClientPublic(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (body %q)", w.Code, w.Body.String())
	}
}

// TestGetFhirClientPublicDatabaseFailureIsNotEchoed pins the error path: an
// unexpected DB error must not be answered with 200 or with the driver text.
func TestGetFhirClientPublicDatabaseFailureIsNotEchoed(t *testing.T) {
	stubClient(t, dbgen.FhirClient{}, sql.ErrConnDone)

	c, w := newClientParamRequest("client-1")
	getFhirClientPublic(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500 (body %q)", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "connection is already closed") {
		t.Errorf("driver error text leaked to the client: %q", w.Body.String())
	}
}

// TestConsentScreenSubmissionIsAccepted pins the contract between the SPA
// consent screen (frontend/src/routes/smart/authorize/+page.svelte) and the
// authorization handler: the request parameters travel in the query string and
// the approval travels in the form body as `consent=approve`.
//
// The handler reads the parameters with c.Query() and only the consent with
// c.PostForm(), so a consent screen that posted everything in the body would
// produce invalid_request for every user. This test fails if either side of
// that split changes without the other.
func TestConsentScreenSubmissionIsAccepted(t *testing.T) {
	useTestSigningKey(t)
	staff := mintStaffToken(t, 7, nil)
	stubClient(t, dbgen.FhirClient{
		ClientID:     "client-1",
		RedirectUris: "https://app.example.com/callback",
		Scopes:       "launch patient/*.read openid fhirUser",
	}, nil)
	stubPatientExists(t, true)
	captured := stubAuthCodeInsert(t)

	// Exactly what the consent screen submits: the forwarded parameters (the
	// screen drops anything else, e.g. `aud`) plus consent=approve in the body.
	query := url.Values{
		"response_type": {"code"},
		"client_id":     {"client-1"},
		"redirect_uri":  {"https://app.example.com/callback"},
		"scope":         {"launch patient/*.read system/*.write openid fhirUser"},
		"state":         {"xyz"},
		"launch":        {"42"},
	}
	form := url.Values{"consent": {"approve"}}

	gin.SetMode(gin.TestMode)
	// Run through a real engine so the response is flushed exactly as it would
	// be in production (a bare gin.CreateTestContext never writes the status).
	engine := gin.New()
	engine.POST("/oauth2/authorize", OAuth2Authorize)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/oauth2/authorize?"+query.Encode(),
		strings.NewReader(form.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "jwt", Value: staff})

	engine.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302 so the browser follows the code back to the app (body %q)", w.Code, w.Body.String())
	}
	if !strings.HasPrefix(w.Header().Get("Location"), "https://app.example.com/callback?code=") {
		t.Fatalf("Location = %q", w.Header().Get("Location"))
	}
	if !captured.called {
		t.Fatal("consent=approve did not produce an authorization code")
	}
	// Approving must not widen anything: the scope the client is not registered
	// for is still dropped, and the launch context still selects the patient.
	if captured.arg.Scopes != "launch patient/*.read openid fhirUser" {
		t.Errorf("stored scopes = %q, want the registered subset", captured.arg.Scopes)
	}
	if captured.arg.PatientID != 42 {
		t.Errorf("stored patient_id = %d, want 42", captured.arg.PatientID)
	}
	if captured.arg.UserID != 7 {
		t.Errorf("stored user_id = %d, want the authenticated staff user 7", captured.arg.UserID)
	}
}

// TestSmartConfigurationAdvertisesTheConsentScreen keeps discovery and the SPA
// consent route in step: the consent screen is only reachable if discovery sends
// apps there, and the route must not sit under a JSON-surface prefix (those are
// answered with JSON, not the SPA index).
func TestSmartConfigurationAdvertisesTheConsentScreen(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/smart-configuration", nil)
	req.Host = "emr.example.com"
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	SmartConfiguration(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var doc map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil {
		t.Fatalf("discovery document is not JSON: %v", err)
	}
	got, _ := doc["authorization_endpoint"].(string)
	if got != "http://emr.example.com/smart/authorize" {
		t.Fatalf("authorization_endpoint = %q, want http://emr.example.com/smart/authorize", got)
	}
	// The route must be served by the SPA fallback, i.e. not JSON-only.
	for _, prefix := range []string{"/api", "/auth", "/oauth2", "/.well-known", "/portal"} {
		if strings.HasPrefix(got, "http://emr.example.com"+prefix) {
			t.Errorf("consent route %q sits under the JSON-surface prefix %q", got, prefix)
		}
	}
	if doc["token_endpoint"] != "http://emr.example.com/oauth2/token" {
		t.Errorf("token_endpoint changed unexpectedly: %v", doc["token_endpoint"])
	}
}
