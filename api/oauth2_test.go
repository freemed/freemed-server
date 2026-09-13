package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/freemed/freemed-server/config"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/gin-gonic/gin"
	jwtlib "github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

const testMasterSecret = "unit-test-master-signing-secret-0123456789"

// useTestSigningKey installs a master signing key for the duration of the test
// so staff cookies and FHIR access tokens can actually be minted and verified.
func useTestSigningKey(t *testing.T) {
	t.Helper()
	prev := config.Config.Session.Key
	config.Config.Session.Key = testMasterSecret
	t.Cleanup(func() { config.Config.Session.Key = prev })
}

// mintStaffToken produces a real staff-class session token signed with the
// STAFF domain key.
func mintStaffToken(t *testing.T, userID int64, extra jwtlib.MapClaims) string {
	t.Helper()
	claims := jwtlib.MapClaims{
		"id":        float64(userID),
		"user_type": "admin",
		"exp":       time.Now().Add(time.Hour).Unix(),
	}
	for k, v := range extra {
		claims[k] = v
	}
	signed, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).
		SignedString(config.Config.DomainKey(config.KeyDomainStaff))
	if err != nil {
		t.Fatalf("failed to mint staff token: %v", err)
	}
	return signed
}

func newAuthorizeRequest(t *testing.T, params url.Values, staffToken string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/oauth2/authorize?"+params.Encode(), nil)
	if staffToken != "" {
		req.AddCookie(&http.Cookie{Name: "jwt", Value: staffToken})
	}
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c, w
}

func newTokenRequest(t *testing.T, form url.Values) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c, w
}

// fakeResult is the minimal sql.Result the insert seams return.
type fakeResult struct{}

func (fakeResult) LastInsertId() (int64, error) { return 1, nil }
func (fakeResult) RowsAffected() (int64, error) { return 1, nil }

type capturedAuthCode struct {
	arg    dbgen.InsertFhirAuthCodeParams
	called bool
}

func stubAuthCodeInsert(t *testing.T) *capturedAuthCode {
	t.Helper()
	captured := &capturedAuthCode{}
	orig := fhirInsertAuthCode
	fhirInsertAuthCode = func(ctx context.Context, arg dbgen.InsertFhirAuthCodeParams) (sql.Result, error) {
		captured.arg = arg
		captured.called = true
		return fakeResult{}, nil
	}
	t.Cleanup(func() { fhirInsertAuthCode = orig })
	return captured
}

func stubClient(t *testing.T, client dbgen.FhirClient, err error) {
	t.Helper()
	orig := fhirGetClientByID
	fhirGetClientByID = func(ctx context.Context, clientID string) (dbgen.FhirClient, error) {
		if err != nil {
			return dbgen.FhirClient{}, err
		}
		if client.ClientID != clientID {
			return dbgen.FhirClient{}, sql.ErrNoRows
		}
		return client, nil
	}
	t.Cleanup(func() { fhirGetClientByID = orig })
}

func stubPatientExists(t *testing.T, exists bool) {
	t.Helper()
	orig := fhirPatientExists
	fhirPatientExists = func(ctx context.Context, patientID int64) (bool, error) { return exists, nil }
	t.Cleanup(func() { fhirPatientExists = orig })
}

func stubAuthCodeLookup(t *testing.T, code dbgen.FhirAuthCode, err error) {
	t.Helper()
	orig := fhirGetAuthCode
	fhirGetAuthCode = func(ctx context.Context, lookup string) (dbgen.FhirAuthCode, error) {
		if err != nil {
			return dbgen.FhirAuthCode{}, err
		}
		if lookup != code.Code {
			return dbgen.FhirAuthCode{}, sql.ErrNoRows
		}
		return code, nil
	}
	t.Cleanup(func() { fhirGetAuthCode = orig })
}

func stubAuthCodeConsume(t *testing.T, rows int64) *int64 {
	t.Helper()
	calls := new(int64)
	orig := fhirConsumeAuthCode
	fhirConsumeAuthCode = func(ctx context.Context, id int64) (int64, error) {
		*calls++
		return rows, nil
	}
	t.Cleanup(func() { fhirConsumeAuthCode = orig })
	return calls
}

func stubAccessTokenInsert(t *testing.T) *dbgen.InsertFhirAccessTokenParams {
	t.Helper()
	captured := new(dbgen.InsertFhirAccessTokenParams)
	orig := fhirInsertAccessToken
	fhirInsertAccessToken = func(ctx context.Context, arg dbgen.InsertFhirAccessTokenParams) (sql.Result, error) {
		*captured = arg
		return fakeResult{}, nil
	}
	t.Cleanup(func() { fhirInsertAccessToken = orig })
	return captured
}

func decodeJSON(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var out map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("response is not JSON: %v (body %q)", err, w.Body.String())
	}
	return out
}

// ============================================================================
// /oauth2/authorize
// ============================================================================

// TestOAuth2AuthorizeRequiresStaffSession is H4's central control: the
// authorization endpoint used to be reachable by anyone, so anyone could mint a
// code (and then a token) for any patient.
func TestOAuth2AuthorizeRequiresStaffSession(t *testing.T) {
	useTestSigningKey(t)

	params := url.Values{
		"response_type": {"code"},
		"client_id":     {"client-1"},
		"redirect_uri":  {"https://app.example.com/callback"},
	}

	t.Run("no session at all", func(t *testing.T) {
		c, w := newAuthorizeRequest(t, params, "")
		OAuth2Authorize(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401 (body %q)", w.Code, w.Body.String())
		}
		body := decodeJSON(t, w)
		if body["error"] != "login_required" {
			t.Errorf("error = %v, want login_required", body["error"])
		}
	})

	t.Run("garbage cookie", func(t *testing.T) {
		c, w := newAuthorizeRequest(t, params, "not-a-jwt")
		OAuth2Authorize(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401 (body %q)", w.Code, w.Body.String())
		}
	})

	t.Run("patient-class token in the jwt cookie (H2 regression)", func(t *testing.T) {
		portal := mintStaffToken(t, 1, jwtlib.MapClaims{"role": "patient", "portal": true})
		c, w := newAuthorizeRequest(t, params, portal)
		OAuth2Authorize(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401 — a portal-class token must never authorize (body %q)", w.Code, w.Body.String())
		}
	})
}

func TestOAuth2AuthorizeRequiresConsent(t *testing.T) {
	useTestSigningKey(t)
	staff := mintStaffToken(t, 7, nil)

	params := url.Values{
		"response_type": {"code"},
		"client_id":     {"client-1"},
		"redirect_uri":  {"https://app.example.com/callback"},
		"scope":         {"patient/*.read"},
	}
	c, w := newAuthorizeRequest(t, params, staff)
	OAuth2Authorize(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body %q)", w.Code, w.Body.String())
	}
	body := decodeJSON(t, w)
	if body["error"] != "consent_required" || body["consent_required"] != true {
		t.Errorf("body = %v, want error=consent_required consent_required=true", body)
	}
	// Nothing may be persisted before consent.
}

// TestOAuth2AuthorizeBindsUserAndLaunchPatient proves the code is bound to the
// staff user (not user_id = 0), that `launch` — not `aud` — supplies the patient
// compartment, and that the client cannot ask for scopes it does not hold.
func TestOAuth2AuthorizeBindsUserAndLaunchPatient(t *testing.T) {
	useTestSigningKey(t)
	staff := mintStaffToken(t, 7, nil)
	stubClient(t, dbgen.FhirClient{
		ClientID:     "client-1",
		RedirectUris: "https://app.example.com/callback",
		Scopes:       "launch patient/*.read openid fhirUser",
	}, nil)
	stubPatientExists(t, true)
	captured := stubAuthCodeInsert(t)

	params := url.Values{
		"response_type": {"code"},
		"client_id":     {"client-1"},
		"redirect_uri":  {"https://app.example.com/callback"},
		"scope":         {"patient/*.read system/*.write openid fhirUser"},
		"launch":        {"42"},
		"aud":           {"99"}, // the audience is NOT a patient id
		"state":         {"xyz"},
		"consent":       {"approve"},
	}
	c, w := newAuthorizeRequest(t, params, staff)
	OAuth2Authorize(c)

	if w.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302 (body %q)", w.Code, w.Body.String())
	}
	location := w.Header().Get("Location")
	if !strings.HasPrefix(location, "https://app.example.com/callback?code=") || !strings.HasSuffix(location, "&state=xyz") {
		t.Errorf("Location = %q, want a code and the echoed state", location)
	}
	redirected, err := url.Parse(location)
	if err != nil {
		t.Fatalf("Location is not a URL: %v", err)
	}
	code := redirected.Query().Get("code")
	if len(code) != 64 {
		t.Errorf("code = %q, want a 32-byte hex code", code)
	}

	if !captured.called {
		t.Fatal("no authorization code was stored")
	}
	got := captured.arg
	if got.UserID != 7 {
		t.Errorf("stored user_id = %d, want 7 (the authenticated staff user, not 0)", got.UserID)
	}
	if got.PatientID != 42 {
		t.Errorf("stored patient_id = %d, want 42 from `launch` (aud=99 must be ignored)", got.PatientID)
	}
	if got.ClientID != "client-1" {
		t.Errorf("stored client_id = %q, want client-1", got.ClientID)
	}
	if got.RedirectUri != "https://app.example.com/callback" {
		t.Errorf("stored redirect_uri = %q", got.RedirectUri)
	}
	if got.Scopes != "patient/*.read openid fhirUser" {
		t.Errorf("stored scopes = %q, want the registered subset (system/*.write must be dropped)", got.Scopes)
	}
	if got.Code != code {
		t.Errorf("stored code = %q, want the code in the redirect (%q)", got.Code, code)
	}
}

// TestOAuth2AuthorizeIgnoresAudWithoutLaunch pins down that `aud` alone never
// becomes a patient compartment.
func TestOAuth2AuthorizeIgnoresAudWithoutLaunch(t *testing.T) {
	useTestSigningKey(t)
	staff := mintStaffToken(t, 7, nil)
	stubClient(t, dbgen.FhirClient{
		ClientID:     "client-1",
		RedirectUris: "https://app.example.com/callback",
		Scopes:       "launch patient/*.read",
	}, nil)
	captured := stubAuthCodeInsert(t)

	params := url.Values{
		"response_type": {"code"},
		"client_id":     {"client-1"},
		"redirect_uri":  {"https://app.example.com/callback"},
		"scope":         {"patient/*.read"},
		"aud":           {"42"},
		"consent":       {"approve"},
	}
	c, w := newAuthorizeRequest(t, params, staff)
	OAuth2Authorize(c)

	if w.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302 (body %q)", w.Code, w.Body.String())
	}
	if captured.arg.PatientID != 0 {
		t.Errorf("stored patient_id = %d, want 0: `aud` is the audience, not a patient id", captured.arg.PatientID)
	}
}

func TestOAuth2AuthorizeRejectsBadLaunchContext(t *testing.T) {
	useTestSigningKey(t)
	staff := mintStaffToken(t, 7, nil)

	base := func() url.Values {
		return url.Values{
			"response_type": {"code"},
			"client_id":     {"client-1"},
			"redirect_uri":  {"https://app.example.com/callback"},
			"scope":         {"patient/*.read"},
			"consent":       {"approve"},
		}
	}
	stubClient(t, dbgen.FhirClient{
		ClientID:     "client-1",
		RedirectUris: "https://app.example.com/callback",
		Scopes:       "launch patient/*.read",
	}, nil)

	t.Run("non-numeric patient is rejected, not coerced to 0", func(t *testing.T) {
		stubPatientExists(t, true)
		params := base()
		params.Set("launch", "abc")
		c, w := newAuthorizeRequest(t, params, staff)
		OAuth2Authorize(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400 (body %q)", w.Code, w.Body.String())
		}
	})

	t.Run("zero patient is rejected", func(t *testing.T) {
		stubPatientExists(t, true)
		params := base()
		params.Set("launch", "0")
		c, w := newAuthorizeRequest(t, params, staff)
		OAuth2Authorize(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400 (body %q)", w.Code, w.Body.String())
		}
	})

	t.Run("unknown patient is rejected", func(t *testing.T) {
		stubPatientExists(t, false)
		params := base()
		params.Set("launch", "4242")
		c, w := newAuthorizeRequest(t, params, staff)
		OAuth2Authorize(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400 (body %q)", w.Code, w.Body.String())
		}
	})

	t.Run("patient parameter is the standalone fallback", func(t *testing.T) {
		stubPatientExists(t, true)
		captured := stubAuthCodeInsert(t)
		params := base()
		params.Set("patient", "12")
		c, w := newAuthorizeRequest(t, params, staff)
		OAuth2Authorize(c)
		if w.Code != http.StatusFound {
			t.Fatalf("status = %d, want 302 (body %q)", w.Code, w.Body.String())
		}
		if captured.arg.PatientID != 12 {
			t.Errorf("stored patient_id = %d, want 12 from `patient`", captured.arg.PatientID)
		}
	})
}

func TestOAuth2AuthorizeRejectsUnregisteredScopes(t *testing.T) {
	useTestSigningKey(t)
	staff := mintStaffToken(t, 7, nil)
	stubClient(t, dbgen.FhirClient{
		ClientID:     "client-1",
		RedirectUris: "https://app.example.com/callback",
		Scopes:       "launch patient/*.read",
	}, nil)
	stubAuthCodeInsert(t)

	params := url.Values{
		"response_type": {"code"},
		"client_id":     {"client-1"},
		"redirect_uri":  {"https://app.example.com/callback"},
		"scope":         {"system/*.write"},
		"consent":       {"approve"},
	}
	c, w := newAuthorizeRequest(t, params, staff)
	OAuth2Authorize(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body %q)", w.Code, w.Body.String())
	}
	if body := decodeJSON(t, w); body["error"] != "invalid_scope" {
		t.Errorf("error = %v, want invalid_scope", body["error"])
	}
}

func TestOAuth2AuthorizeRejectsUnknownRedirectURI(t *testing.T) {
	useTestSigningKey(t)
	staff := mintStaffToken(t, 7, nil)
	stubClient(t, dbgen.FhirClient{
		ClientID:     "client-1",
		RedirectUris: "https://app.example.com/callback",
		Scopes:       "launch patient/*.read",
	}, nil)

	params := url.Values{
		"response_type": {"code"},
		"client_id":     {"client-1"},
		"redirect_uri":  {"https://evil.example.com/callback"},
		"scope":         {"patient/*.read"},
		"consent":       {"approve"},
	}
	c, w := newAuthorizeRequest(t, params, staff)
	OAuth2Authorize(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body %q)", w.Code, w.Body.String())
	}
}

// ============================================================================
// /oauth2/token — authorization_code
// ============================================================================

// TestFhirAuthCodeRedeemable covers the code bindings that used to be skipped
// whenever the caller simply omitted client_id or redirect_uri.
func TestFhirAuthCodeRedeemable(t *testing.T) {
	now := time.Now()
	base := dbgen.FhirAuthCode{
		ID:          1,
		Code:        "code-a",
		ClientID:    "client-a",
		UserID:      7,
		PatientID:   42,
		Scopes:      "patient/*.read",
		RedirectUri: "https://app.example.com/callback",
		ExpiresAt:   now.Add(5 * time.Minute),
	}

	tests := []struct {
		name        string
		mutate      func(a *dbgen.FhirAuthCode)
		clientID    string
		redirectURI string
		wantOK      bool
		wantReason  string
	}{
		{"valid", func(a *dbgen.FhirAuthCode) {}, "client-a", "https://app.example.com/callback", true, ""},
		{"foreign client", func(a *dbgen.FhirAuthCode) {}, "client-b", "https://app.example.com/callback", false, "client_id mismatch"},
		{"foreign redirect uri", func(a *dbgen.FhirAuthCode) {}, "client-a", "https://evil.example.com/callback", false, "redirect_uri mismatch"},
		{"expired", func(a *dbgen.FhirAuthCode) { a.ExpiresAt = now.Add(-time.Second) }, "client-a", "https://app.example.com/callback", false, "Authorization code expired"},
		{"already used", func(a *dbgen.FhirAuthCode) { a.Used = true }, "client-a", "https://app.example.com/callback", false, "Authorization code already used"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := base
			tt.mutate(&code)
			ok, reason := fhirAuthCodeRedeemable(code, tt.clientID, tt.redirectURI, now)
			if ok != tt.wantOK || reason != tt.wantReason {
				t.Errorf("fhirAuthCodeRedeemable = (%v, %q), want (%v, %q)", ok, reason, tt.wantOK, tt.wantReason)
			}
		})
	}
}

func validAuthCode() dbgen.FhirAuthCode {
	return dbgen.FhirAuthCode{
		ID:          1,
		Code:        "code-a",
		ClientID:    "client-a",
		UserID:      7,
		PatientID:   42,
		Scopes:      "patient/*.read openid fhirUser",
		RedirectUri: "https://app.example.com/callback",
		ExpiresAt:   time.Now().Add(5 * time.Minute),
	}
}

func TestHandleAuthorizationCodeGrantRejectsBadCodes(t *testing.T) {
	useTestSigningKey(t)
	stubAccessTokenInsert(t)

	form := func() url.Values {
		return url.Values{
			"grant_type":   {"authorization_code"},
			"code":         {"code-a"},
			"client_id":    {"client-a"},
			"redirect_uri": {"https://app.example.com/callback"},
		}
	}

	tests := []struct {
		name      string
		mutate    func(a *dbgen.FhirAuthCode)
		form      func(f url.Values)
		lookupErr error
		wantCode  int
		wantErr   string
	}{
		{
			name:     "foreign client_id",
			form:     func(f url.Values) { f.Set("client_id", "client-b") },
			wantCode: http.StatusBadRequest, wantErr: "client_id mismatch",
		},
		{
			name:     "foreign redirect_uri",
			form:     func(f url.Values) { f.Set("redirect_uri", "https://evil.example.com/callback") },
			wantCode: http.StatusBadRequest, wantErr: "redirect_uri mismatch",
		},
		{
			name:     "expired code",
			mutate:   func(a *dbgen.FhirAuthCode) { a.ExpiresAt = time.Now().Add(-time.Minute) },
			wantCode: http.StatusBadRequest, wantErr: "Authorization code expired",
		},
		{
			name:     "already used code",
			mutate:   func(a *dbgen.FhirAuthCode) { a.Used = true },
			wantCode: http.StatusBadRequest, wantErr: "Authorization code already used",
		},
		{
			name:     "missing client_id",
			form:     func(f url.Values) { f.Del("client_id") },
			wantCode: http.StatusBadRequest, wantErr: "client_id is required",
		},
		{
			name:     "missing redirect_uri",
			form:     func(f url.Values) { f.Del("redirect_uri") },
			wantCode: http.StatusBadRequest, wantErr: "redirect_uri is required",
		},
		{
			name:      "unknown code",
			lookupErr: sql.ErrNoRows,
			wantCode:  http.StatusBadRequest, wantErr: "Invalid or expired authorization code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := validAuthCode()
			if tt.mutate != nil {
				tt.mutate(&code)
			}
			stubAuthCodeLookup(t, code, tt.lookupErr)
			stubAuthCodeConsume(t, 1)

			f := form()
			if tt.form != nil {
				tt.form(f)
			}
			c, w := newTokenRequest(t, f)
			handleAuthorizationCodeGrant(c)

			if w.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d (body %q)", w.Code, tt.wantCode, w.Body.String())
			}
			if body := decodeJSON(t, w); body["error_description"] != tt.wantErr {
				t.Errorf("error_description = %v, want %q", body["error_description"], tt.wantErr)
			}
		})
	}
}

// TestHandleAuthorizationCodeGrantRejectsReusedCode covers the atomic single-use
// consumption: the UPDATE only matches while used = 0, so a second redemption
// (including a concurrent one) sees 0 rows affected.
func TestHandleAuthorizationCodeGrantRejectsReusedCode(t *testing.T) {
	useTestSigningKey(t)
	stubAuthCodeLookup(t, validAuthCode(), nil)
	calls := stubAuthCodeConsume(t, 0)

	c, w := newTokenRequest(t, url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {"code-a"},
		"client_id":    {"client-a"},
		"redirect_uri": {"https://app.example.com/callback"},
	})
	handleAuthorizationCodeGrant(c)

	if *calls != 1 {
		t.Errorf("consume calls = %d, want 1", *calls)
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body %q)", w.Code, w.Body.String())
	}
	if body := decodeJSON(t, w); body["error"] != "invalid_grant" {
		t.Errorf("error = %v, want invalid_grant", body["error"])
	}
}

func TestHandleAuthorizationCodeGrantIssuesBoundToken(t *testing.T) {
	useTestSigningKey(t)
	code := validAuthCode()
	stubAuthCodeLookup(t, code, nil)
	stubAuthCodeConsume(t, 1)
	stored := stubAccessTokenInsert(t)

	c, w := newTokenRequest(t, url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {"code-a"},
		"client_id":    {"client-a"},
		"redirect_uri": {"https://app.example.com/callback"},
	})
	handleAuthorizationCodeGrant(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", w.Code, w.Body.String())
	}
	body := decodeJSON(t, w)
	tokenString, _ := body["access_token"].(string)
	if tokenString == "" {
		t.Fatal("no access_token in the response")
	}
	if body["patient"] != "42" {
		t.Errorf("patient = %v, want \"42\"", body["patient"])
	}

	claims, err := ValidateFHIRJWT(tokenString)
	if err != nil {
		t.Fatalf("issued token does not validate with the FHIR domain key: %v", err)
	}
	if claims["sub"] != "7" {
		t.Errorf("sub = %v, want 7 (the authorizing staff user)", claims["sub"])
	}
	if claims["patient"] != "42" {
		t.Errorf("patient claim = %v, want 42", claims["patient"])
	}
	if claims["scope"] != code.Scopes {
		t.Errorf("scope claim = %v, want %q", claims["scope"], code.Scopes)
	}

	// The stored hash must be the hash of the token that was returned, and the
	// token must remain bound to the same user and patient.
	if stored.TokenHash != sha256Hash(tokenString) {
		t.Error("stored token_hash does not match the issued token")
	}
	if stored.UserID != 7 || stored.PatientID != 42 || stored.ClientID != "client-a" {
		t.Errorf("stored token row = %+v, want user 7 / patient 42 / client-a", stored)
	}
}

// ============================================================================
// /oauth2/token — client_credentials
// ============================================================================

// TestHandleClientCredentialsGrantRequiresASecret is the H4 client-auth fix: a
// confidential client's client_id alone used to mint a one-hour token.
func TestHandleClientCredentialsGrantRequiresASecret(t *testing.T) {
	useTestSigningKey(t)

	secret := "s3cret-abcdefghijklmnop"
	hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	confidential := dbgen.FhirClient{
		ID:               1,
		ClientID:         "client-c",
		ClientName:       "Confidential",
		Scopes:           "system/*.read",
		IsConfidential:   true,
		Active:           true,
		ClientSecretHash: string(hash),
	}

	tests := []struct {
		name       string
		client     dbgen.FhirClient
		clientErr  error
		clientID   string
		secret     string
		useBasic   bool
		wantStatus int
		wantToken  bool
	}{
		{
			name: "client_id alone (the audit's repro)", client: confidential, clientID: "client-c",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "wrong secret", client: confidential, clientID: "client-c", secret: "wrong",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "public client", client: dbgen.FhirClient{ClientID: "client-c", Scopes: "launch"}, clientID: "client-c", secret: secret,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "confidential client with an empty hash", client: dbgen.FhirClient{ClientID: "client-c", Scopes: "system/*.read", IsConfidential: true},
			clientID: "client-c", secret: secret,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "unknown client", clientErr: sql.ErrNoRows, clientID: "client-c", secret: secret,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "correct secret", client: confidential, clientID: "client-c", secret: secret,
			wantStatus: http.StatusOK, wantToken: true,
		},
		{
			name: "correct secret over HTTP Basic", client: confidential, clientID: "", secret: secret, useBasic: true,
			wantStatus: http.StatusOK, wantToken: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubClient(t, tt.client, tt.clientErr)
			stored := stubAccessTokenInsert(t)

			form := url.Values{"grant_type": {"client_credentials"}}
			if tt.clientID != "" {
				form.Set("client_id", tt.clientID)
			}
			if tt.secret != "" {
				form.Set("client_secret", tt.secret)
			}
			c, w := newTokenRequest(t, form)
			if tt.useBasic {
				c.Request.SetBasicAuth(tt.client.ClientID, tt.secret)
			}
			handleClientCredentialsGrant(c)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body %q)", w.Code, tt.wantStatus, w.Body.String())
			}
			body := decodeJSON(t, w)
			if !tt.wantToken {
				if body["error"] != "invalid_client" {
					t.Errorf("error = %v, want invalid_client", body["error"])
				}
				if stored.TokenHash != "" {
					t.Error("a token was stored despite the grant being rejected")
				}
				return
			}
			tokenString, _ := body["access_token"].(string)
			if tokenString == "" {
				t.Fatal("no access_token in the response")
			}
			if body["scope"] != "system/*.read" {
				t.Errorf("scope = %v, want the client's registered scopes", body["scope"])
			}
			if _, err := ValidateFHIRJWT(tokenString); err != nil {
				t.Fatalf("issued token does not validate with the FHIR domain key: %v", err)
			}
			if stored.TokenHash != sha256Hash(tokenString) {
				t.Error("stored token_hash does not match the issued token")
			}
		})
	}
}

func TestFhirConsentApproved(t *testing.T) {
	for input, want := range map[string]bool{
		"approve": true, "APPROVE": true, " approved ": true, "yes": true, "true": true,
		"": false, "deny": false, "false": false, "maybe": false, "1": false,
	} {
		if got := fhirConsentApproved(input); got != want {
			t.Errorf("fhirConsentApproved(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestFhirGrantedScopes(t *testing.T) {
	tests := []struct {
		requested  string
		registered string
		want       string
	}{
		{"", "launch patient/*.read openid fhirUser", "launch patient/*.read openid fhirUser"},
		{"patient/*.read", "launch patient/*.read openid fhirUser", "patient/*.read"},
		{"patient/*.read system/*.write", "patient/*.read", "patient/*.read"},
		{"system/*.write", "patient/*.read", ""},
		{"openid fhirUser", "openid fhirUser", "openid fhirUser"},
	}
	for _, tt := range tests {
		if got := fhirGrantedScopes(tt.requested, tt.registered); got != tt.want {
			t.Errorf("fhirGrantedScopes(%q, %q) = %q, want %q", tt.requested, tt.registered, got, tt.want)
		}
	}
}
