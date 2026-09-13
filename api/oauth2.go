package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/freemed/freemed-server/config"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Query seams.
//
// api/ is a separate Go module and its tests have no live MySQL, so the handful
// of queries the OAuth2 grant handlers need are indirected through these
// variables (the same pattern api/dicom.go uses for the DICOM ownership guard).
// Handlers call the seams, tests substitute in-memory stubs, and every line of
// the authorization logic under test is the real one.
var (
	fhirGetClientByID = func(ctx context.Context, clientID string) (dbgen.FhirClient, error) {
		return model.Queries.GetFhirClientByID(ctx, clientID)
	}
	fhirGetAuthCode = func(ctx context.Context, code string) (dbgen.FhirAuthCode, error) {
		return model.Queries.GetFhirAuthCode(ctx, code)
	}
	fhirConsumeAuthCode = func(ctx context.Context, id int64) (int64, error) {
		return model.Queries.ConsumeFhirAuthCode(ctx, id)
	}
	fhirInsertAuthCode = func(ctx context.Context, arg dbgen.InsertFhirAuthCodeParams) (sql.Result, error) {
		return model.Queries.InsertFhirAuthCode(ctx, arg)
	}
	fhirInsertAccessToken = func(ctx context.Context, arg dbgen.InsertFhirAccessTokenParams) (sql.Result, error) {
		return model.Queries.InsertFhirAccessToken(ctx, arg)
	}
	fhirPatientExists = func(ctx context.Context, patientID int64) (bool, error) {
		_, err := model.Queries.FhirPatientById(ctx, patientID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	}
)

// OAuth2Authorize handles GET/POST /oauth2/authorize.
//
// It validates the authorization request, requires an authenticated staff
// session (the resource owner), requires explicit consent, resolves the launch
// patient from `launch`/`patient`, narrows the granted scopes to the client's
// registration, and only then issues a short-lived authorization code bound to
// the staff user, the client and the redirect URI.
//
// Before this, the endpoint checked nothing but client_id + an exact
// redirect_uri and read the patient straight out of `aud` — so an anonymous
// caller could mint a code for any patient, and the code was stored with
// user_id = 0 because no staff middleware is mounted on the /oauth2 group.
func OAuth2Authorize(c *gin.Context) {
	responseType := c.Query("response_type")
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	scope := c.Query("scope")
	state := c.Query("state")

	// SMART launch context. `launch` carries the EHR launch context and
	// `patient` the standalone-launch patient.
	//
	// `aud` is deliberately NOT read: it is the AUDIENCE (the FHIR base URL the
	// client intends to call, e.g. http://host/api/fhir). Treating it as a
	// patient id — as this handler did — let the caller choose any patient id.
	launch := strings.TrimSpace(c.Query("launch"))
	patientParam := strings.TrimSpace(c.Query("patient"))

	if responseType != "code" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_response_type", "error_description": "Only response_type=code is supported"})
		return
	}

	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "client_id is required"})
		return
	}

	if redirectURI == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "redirect_uri is required"})
		return
	}

	// 1. The resource owner must be an authenticated staff user. The /oauth2
	// group carries no staff middleware (it also serves the unauthenticated
	// token and introspect endpoints), so the session is resolved here from the
	// `jwt` cookie with the staff signing key.
	staffClaims, err := staffSessionFromCookie(c)
	if err != nil {
		log.Printf("OAuth2Authorize: rejected unauthenticated authorization request: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "login_required",
			"error_description": "An authenticated staff session is required to authorize a SMART client",
		})
		return
	}
	userID, err := fhirClaimInt64(staffClaims, "id")
	if err != nil || userID <= 0 {
		log.Printf("OAuth2Authorize: staff session carries no usable user id: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "login_required",
			"error_description": "The staff session does not identify a user",
		})
		return
	}

	// 2. Explicit consent. Nothing is inferred: the request must carry the
	// resource owner's acknowledgement of the requested scopes. There is no
	// consent UI yet (frontend follow-up), so an unacknowledged request is
	// refused with a renderable error instead of being treated as approved.
	consent := strings.TrimSpace(c.Query("consent"))
	if consent == "" && c.Request.Method == http.MethodPost {
		consent = strings.TrimSpace(c.PostForm("consent"))
	}
	if !fhirConsentApproved(consent) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "consent_required",
			"error_description": "The resource owner has not approved this authorization request. Show the requested scopes to the user and re-issue the request with consent=approve.",
			"consent_required":  true,
			"requested_scopes":  strings.TrimSpace(scope),
		})
		return
	}

	// 3. Client validation.
	client, err := fhirGetClientByID(c.Request.Context(), clientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_client", "error_description": "Unknown or inactive client"})
			return
		}
		log.Printf("OAuth2Authorize: failed to lookup client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	if !matchRedirectURI(client.RedirectUris, redirectURI) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "redirect_uri does not match registered URIs"})
		return
	}

	// 4. Launch context → patient compartment. An unparseable or unknown
	// patient is rejected outright; it is never coerced to 0 by
	// common.ParseInt, which silently turned "aud=abc" into "no patient".
	var patientID int64
	rawLaunch := launch
	if rawLaunch == "" {
		rawLaunch = patientParam
	}
	if rawLaunch != "" {
		pid, perr := strconv.ParseInt(rawLaunch, 10, 64)
		if perr != nil || pid <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":             "invalid_request",
				"error_description": "launch (or patient) must be a numeric patient identifier",
			})
			return
		}
		exists, derr := fhirPatientExists(c.Request.Context(), pid)
		if derr != nil {
			log.Printf("OAuth2Authorize: failed to validate launch patient %d: %v", pid, derr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
			return
		}
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":             "invalid_request",
				"error_description": "launch context names a patient that does not exist",
			})
			return
		}
		patientID = pid
	}

	// 5. Grant only what the client is registered for.
	scopes := fhirGrantedScopes(scope, client.Scopes)
	if scopes == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_scope",
			"error_description": "No requested scope is registered for this client",
		})
		return
	}

	// Generate a random auth code
	codeBytes := make([]byte, 32)
	if _, err := rand.Read(codeBytes); err != nil {
		log.Printf("OAuth2Authorize: failed to generate auth code: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}
	code := hex.EncodeToString(codeBytes)

	// Store the auth code.
	//
	// The code is opaque rather than a KeyDomainOAuth2-signed JWT: it only has
	// to be unguessable, bound to (client, redirect_uri, user, patient) and
	// spendable exactly once, and all four of those live on one DB row with a
	// SQL-enforced 5 minute expiry — revocation, replay protection and expiry
	// then have a single source of truth and there is no extra signing key to
	// rotate. The code is bound to the staff user here, so the token it later
	// yields cannot belong to an anonymous caller.
	_, err = fhirInsertAuthCode(c.Request.Context(), dbgen.InsertFhirAuthCodeParams{
		Code:        code,
		ClientID:    clientID,
		UserID:      userID,
		PatientID:   patientID,
		Scopes:      scopes,
		RedirectUri: redirectURI,
	})
	if err != nil {
		log.Printf("OAuth2Authorize: failed to insert auth code: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// Redirect with code and state
	sep := "?"
	if strings.Contains(redirectURI, "?") {
		sep = "&"
	}
	redirectURL := fmt.Sprintf("%s%scode=%s", redirectURI, sep, code)
	if state != "" {
		redirectURL += fmt.Sprintf("&state=%s", state)
	}

	c.Redirect(http.StatusFound, redirectURL)
}

// fhirConsentApproved reports whether the authorization request carries an
// explicit approval of the requested scopes.
func fhirConsentApproved(consent string) bool {
	switch strings.ToLower(strings.TrimSpace(consent)) {
	case "approve", "approved", "yes", "true":
		return true
	default:
		return false
	}
}

// fhirGrantedScopes intersects the scopes a client asked for with the scopes it
// is registered for. A requested scope the client does not hold is dropped
// rather than granted: the authorize endpoint used to store whatever the caller
// put in `scope`, so a client registered for `patient/*.read` could be handed
// `system/*.write`. An empty request means "everything registered".
func fhirGrantedScopes(requested, registered string) string {
	registeredTokens := strings.Fields(registered)
	if strings.TrimSpace(requested) == "" {
		return strings.Join(registeredTokens, " ")
	}
	allowed := make(map[string]bool, len(registeredTokens))
	for _, s := range registeredTokens {
		allowed[s] = true
	}
	granted := make([]string, 0, len(registeredTokens))
	for _, s := range strings.Fields(requested) {
		if allowed[s] {
			granted = append(granted, s)
		}
	}
	return strings.Join(granted, " ")
}

// fhirAuthCodeRedeemable reports whether the stored authorization code may be
// exchanged by this client for this redirect URI, and why not when it may not.
//
// Everything the code is bound to is re-checked here: the client it was issued
// to, the redirect URI it was issued for, its expiry, and single use. Both
// bindings were previously skipped whenever the caller simply omitted the
// parameter, so a leaked code was redeemable by anyone.
func fhirAuthCodeRedeemable(authCode dbgen.FhirAuthCode, clientID, redirectURI string, now time.Time) (bool, string) {
	if authCode.ClientID != clientID {
		return false, "client_id mismatch"
	}
	if authCode.RedirectUri != redirectURI {
		return false, "redirect_uri mismatch"
	}
	if !authCode.ExpiresAt.After(now) {
		return false, "Authorization code expired"
	}
	if authCode.Used {
		return false, "Authorization code already used"
	}
	return true, ""
}

// matchRedirectURI checks if the given redirect URI matches any of the registered URIs.
func matchRedirectURI(registeredURIs, candidate string) bool {
	for _, uri := range strings.Split(registeredURIs, "\n") {
		uri = strings.TrimSpace(uri)
		if uri == "" {
			continue
		}
		if uri == candidate {
			return true
		}
	}
	return false
}

// OAuth2Token handles POST /oauth2/token.
// Supports grant_type=authorization_code and grant_type=client_credentials.
func OAuth2Token(c *gin.Context) {
	grantType := c.PostForm("grant_type")

	switch grantType {
	case "authorization_code":
		handleAuthorizationCodeGrant(c)
	case "client_credentials":
		handleClientCredentialsGrant(c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_grant_type"})
	}
}

// handleAuthorizationCodeGrant exchanges an authorization code for an access token.
func handleAuthorizationCodeGrant(c *gin.Context) {
	code := c.PostForm("code")
	redirectURI := c.PostForm("redirect_uri")
	clientID := c.PostForm("client_id")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "code is required"})
		return
	}
	// client_id and redirect_uri are mandatory at redemption: while they were
	// optional, omitting either one skipped the binding check entirely.
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "client_id is required"})
		return
	}
	if redirectURI == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "redirect_uri is required"})
		return
	}

	// Look up the auth code. The query filters on used = 0 AND expires_at, so
	// an expired or spent code is already invalid_grant here.
	authCode, err := fhirGetAuthCode(c.Request.Context(), code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant", "error_description": "Invalid or expired authorization code"})
			return
		}
		log.Printf("handleAuthorizationCodeGrant: failed to lookup auth code: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// Re-check every binding (client, redirect URI, expiry, single use) so an
	// expired or foreign code is refused even if the lookup is ever loosened.
	if ok, reason := fhirAuthCodeRedeemable(authCode, clientID, redirectURI, time.Now()); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant", "error_description": reason})
		return
	}

	// Spend the code atomically (UPDATE ... WHERE used = 0): two concurrent
	// redemptions cannot both mint a token.
	consumed, err := fhirConsumeAuthCode(c.Request.Context(), authCode.ID)
	if err != nil {
		log.Printf("handleAuthorizationCodeGrant: failed to consume auth code: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}
	if consumed != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant", "error_description": "Authorization code already used"})
		return
	}

	// Parse scopes
	scopes := authCode.Scopes
	if scopes == "" {
		scopes = "patient/*.read openid fhirUser"
	}

	// Generate the JWT access token
	issuer := fmt.Sprintf("http://%s", c.Request.Host)

	claims := jwt.MapClaims{
		"sub":       fmt.Sprintf("%d", authCode.UserID),
		"patient":   fmt.Sprintf("%d", authCode.PatientID),
		"scope":     scopes,
		"iss":       issuer,
		"iat":       time.Now().Unix(),
		"exp":       time.Now().Add(1 * time.Hour).Unix(),
		"jti":       uuid.New().String(),
		"client_id": authCode.ClientID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// These are FHIR access tokens (validated by FHIRAuth and OAuth2Introspect
	// via ValidateFHIRJWT), so they are signed with the FHIR domain key.
	tokenString, err := token.SignedString(config.Config.DomainKey(config.KeyDomainFHIR))
	if err != nil {
		log.Printf("handleAuthorizationCodeGrant: failed to sign JWT: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// Store token hash in DB
	tokenHash := sha256Hash(tokenString)
	_, err = fhirInsertAccessToken(c.Request.Context(), dbgen.InsertFhirAccessTokenParams{
		TokenHash: tokenHash,
		ClientID:  authCode.ClientID,
		UserID:    authCode.UserID,
		PatientID: authCode.PatientID,
		Scopes:    scopes,
	})
	if err != nil {
		log.Printf("handleAuthorizationCodeGrant: failed to insert access token: %v", err)
		// Non-fatal — the token is still valid JWT
	}

	// Build response
	resp := gin.H{
		"access_token": tokenString,
		"token_type":   "bearer",
		"expires_in":   3600,
		"scope":        scopes,
	}
	if authCode.PatientID > 0 {
		resp["patient"] = fmt.Sprintf("%d", authCode.PatientID)
	}

	c.JSON(http.StatusOK, resp)
}

// handleClientCredentialsGrant handles the client_credentials grant for backend services.
//
// The client must now actually authenticate: a confidential client needs a
// client_secret matching the bcrypt hash stored on its row. The credential used
// to be discarded outright (`_ = clientSecretOrScope`) and there was no
// credential column at all, so a confidential client's client_id alone minted a
// one-hour token.
func handleClientCredentialsGrant(c *gin.Context) {
	clientID := c.PostForm("client_id")
	clientSecret := c.PostForm("client_secret")

	// RFC 6749 §2.3.1 also allows HTTP Basic client authentication.
	if clientID == "" || clientSecret == "" {
		if basicUser, basicPass, ok := c.Request.BasicAuth(); ok {
			if clientID == "" {
				clientID = basicUser
			}
			if clientSecret == "" {
				clientSecret = basicPass
			}
		}
	}

	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "client_id is required"})
		return
	}
	if clientSecret == "" {
		invalidClient(c, "client_secret is required")
		return
	}

	// Validate client
	client, err := fhirGetClientByID(c.Request.Context(), clientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			invalidClient(c, "Unknown or inactive client")
			return
		}
		log.Printf("handleClientCredentialsGrant: failed to lookup client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// For client_credentials, the client must be confidential
	if !client.IsConfidential {
		invalidClient(c, "Public clients cannot use client_credentials grant")
		return
	}

	// A confidential client with no stored secret cannot authenticate; reject
	// it rather than letting an empty hash stand in for a valid credential.
	if strings.TrimSpace(client.ClientSecretHash) == "" {
		log.Printf("handleClientCredentialsGrant: confidential client %s has no stored client_secret_hash", clientID)
		invalidClient(c, "Client has no registered secret; re-register the client")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(client.ClientSecretHash), []byte(clientSecret)); err != nil {
		log.Printf("handleClientCredentialsGrant: client authentication failed for %s", clientID)
		invalidClient(c, "Invalid client credentials")
		return
	}

	// Use the client's registered scopes, narrowed by any requested scope.
	scopes := fhirGrantedScopes(c.PostForm("scope"), client.Scopes)

	// Generate JWT
	issuer := fmt.Sprintf("http://%s", c.Request.Host)
	claims := jwt.MapClaims{
		"sub":       clientID,
		"scope":     scopes,
		"iss":       issuer,
		"iat":       time.Now().Unix(),
		"exp":       time.Now().Add(1 * time.Hour).Unix(),
		"jti":       uuid.New().String(),
		"client_id": clientID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// These are FHIR access tokens (validated by FHIRAuth and OAuth2Introspect
	// via ValidateFHIRJWT), so they are signed with the FHIR domain key.
	tokenString, err := token.SignedString(config.Config.DomainKey(config.KeyDomainFHIR))
	if err != nil {
		log.Printf("handleClientCredentialsGrant: failed to sign JWT: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// Store token hash in DB
	tokenHash := sha256Hash(tokenString)
	_, err = fhirInsertAccessToken(c.Request.Context(), dbgen.InsertFhirAccessTokenParams{
		TokenHash: tokenHash,
		ClientID:  clientID,
		UserID:    0,
		PatientID: 0,
		Scopes:    scopes,
	})
	if err != nil {
		log.Printf("handleClientCredentialsGrant: failed to insert access token: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": tokenString,
		"token_type":   "bearer",
		"expires_in":   3600,
		"scope":        scopes,
	})
}

// invalidClient writes an RFC 6749 invalid_client error.
func invalidClient(c *gin.Context, description string) {
	c.Header("WWW-Authenticate", `Basic realm="freemed"`)
	c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client", "error_description": description})
}

// OAuth2Introspect handles POST /oauth2/introspect.
// Validates a token and returns its metadata.
func OAuth2Introspect(c *gin.Context) {
	tokenString := c.PostForm("token")
	if tokenString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"active": false})
		return
	}

	// Try to validate as JWT
	claims, err := ValidateFHIRJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"active": false})
		return
	}

	// Also check DB store
	tokenHash := sha256Hash(tokenString)
	dbToken, err := model.Queries.GetFhirAccessToken(c.Request.Context(), tokenHash)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"active": false})
		return
	}

	// Build introspection response
	resp := gin.H{
		"active":     true,
		"scope":      dbToken.Scopes,
		"client_id":  dbToken.ClientID,
		"token_type": "bearer",
		"exp":        claims["exp"],
		"iat":        claims["iat"],
		"sub":        claims["sub"],
	}
	if dbToken.PatientID > 0 {
		resp["patient"] = fmt.Sprintf("%d", dbToken.PatientID)
	}

	c.JSON(http.StatusOK, resp)
}

// sha256Hash returns the hex-encoded SHA-256 hash of the input string.
func sha256Hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
