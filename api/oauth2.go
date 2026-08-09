package api

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/config"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// OAuth2Authorize handles GET/POST /oauth2/authorize.
// Validates the auth request, generates an auth code, and redirects to the client's redirect_uri.
func OAuth2Authorize(c *gin.Context) {
	responseType := c.Query("response_type")
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	scope := c.Query("scope")
	state := c.Query("state")
	// aud is the launch context (patient or encounter)
	aud := c.Query("aud")

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

	// Validate client
	client, err := model.Queries.GetFhirClientByID(c.Request.Context(), clientID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_client", "error_description": "Unknown or inactive client"})
			return
		}
		log.Printf("OAuth2Authorize: failed to lookup client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// Validate redirect URI matches registered URIs
	if !matchRedirectURI(client.RedirectUris, redirectURI) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "redirect_uri does not match registered URIs"})
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

	// Determine user_id and patient_id from launch context
	var userID int64
	var patientID int64

	// Try to extract user from existing JWT session (cookie-based auth)
	sess, err := common.GetSession(c)
	if err == nil {
		userID = sess.UserId
	}

	// Parse launch context for patient
	if aud != "" {
		patientID = common.ParseInt(aud)
	}
	_ = aud

	// Store the auth code
	_, err = model.Queries.InsertFhirAuthCode(c.Request.Context(), dbgen.InsertFhirAuthCodeParams{
		Code:        code,
		ClientID:    clientID,
		UserID:      userID,
		PatientID:   patientID,
		Scopes:      scope,
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

	// Look up the auth code
	authCode, err := model.Queries.GetFhirAuthCode(c.Request.Context(), code)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant", "error_description": "Invalid or expired authorization code"})
			return
		}
		log.Printf("handleAuthorizationCodeGrant: failed to lookup auth code: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// Verify client_id matches (if provided)
	if clientID != "" && authCode.ClientID != clientID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant", "error_description": "client_id mismatch"})
		return
	}

	// Verify redirect_uri matches (if provided)
	if redirectURI != "" && authCode.RedirectUri != redirectURI {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant", "error_description": "redirect_uri mismatch"})
		return
	}

	// Mark auth code as used
	if err := model.Queries.MarkFhirAuthCodeUsed(c.Request.Context(), authCode.ID); err != nil {
		log.Printf("handleAuthorizationCodeGrant: failed to mark auth code used: %v", err)
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
	tokenString, err := token.SignedString([]byte(config.Config.Session.Key))
	if err != nil {
		log.Printf("handleAuthorizationCodeGrant: failed to sign JWT: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// Store token hash in DB
	tokenHash := sha256Hash(tokenString)
	_, err = model.Queries.InsertFhirAccessToken(c.Request.Context(), dbgen.InsertFhirAccessTokenParams{
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
func handleClientCredentialsGrant(c *gin.Context) {
	clientID := c.PostForm("client_id")
	clientSecretOrScope := c.PostForm("scope")

	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "client_id is required"})
		return
	}

	// Validate client
	client, err := model.Queries.GetFhirClientByID(c.Request.Context(), clientID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_client", "error_description": "Unknown or inactive client"})
			return
		}
		log.Printf("handleClientCredentialsGrant: failed to lookup client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// For client_credentials, the client must be confidential
	if !client.IsConfidential {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_client", "error_description": "Public clients cannot use client_credentials grant"})
		return
	}
	_ = clientSecretOrScope

	// Use client's registered scopes
	scopes := client.Scopes

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
	tokenString, err := token.SignedString([]byte(config.Config.Session.Key))
	if err != nil {
		log.Printf("handleClientCredentialsGrant: failed to sign JWT: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// Store token hash in DB
	tokenHash := sha256Hash(tokenString)
	_, err = model.Queries.InsertFhirAccessToken(c.Request.Context(), dbgen.InsertFhirAccessTokenParams{
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
