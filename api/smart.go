package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	common.ApiMap["smart"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			// The client registry mints the client_id that /oauth2/token turns
			// into access tokens, so it is admin-only — as this file's doc
			// comments always claimed, but nothing enforced: any authenticated
			// user could register a SMART client.
			r.GET("/clients", common.RequireRole("admin"), listFhirClients)
			r.POST("/clients", common.RequireRole("admin"), createFhirClient)
			r.DELETE("/clients/:id", common.RequireRole("admin"), deactivateFhirClient)

			// Read-only, non-secret metadata for the consent screen. It is
			// staff-authenticated but deliberately NOT admin-only: the resource
			// owner asked to consent is any signed-in staff member, not only an
			// administrator. The credential hash is never part of the response.
			r.GET("/clients/:client_id/public", getFhirClientPublic)
		},
	}
}

// smartConfiguration returns the SMART on FHIR metadata at /.well-known/smart-configuration.
// This is registered separately in main.go, not via ApiMap.
func SmartConfiguration(c *gin.Context) {
	issuer := fmt.Sprintf("http://%s", c.Request.Host)
	c.JSON(http.StatusOK, gin.H{
		"issuer": issuer,
		// The advertised authorization endpoint is the browser consent screen,
		// not the raw /oauth2/authorize handler: that handler is a machine
		// endpoint which refuses any request that does not already carry
		// consent=approve, so an app following discovery straight into it could
		// never be authorized by a human. The consent screen is an SPA route
		// (everything outside /api, /auth, /oauth2, /.well-known and /portal is
		// rendered by the SPA fallback); it shows the app and the scopes it is
		// asking for and then re-posts the identical request to
		// /oauth2/authorize with consent=approve. Every server-side check —
		// staff session, explicit consent, redirect_uri match, launch-patient
		// validation and scope narrowing — is still enforced there.
		"authorization_endpoint": fmt.Sprintf("%s/smart/authorize", issuer),
		"token_endpoint":         fmt.Sprintf("%s/oauth2/token", issuer),
		"introspection_endpoint": fmt.Sprintf("%s/oauth2/introspect", issuer),
		"scopes_supported": []string{
			"launch",
			"launch/patient",
			"patient/*.read",
			"patient/*.write",
			"openid",
			"fhirUser",
		},
		"capabilities": []string{
			"launch-ehr",
			"launch-standalone",
			"client-public",
			"client-confidential-symmetric",
			"permission-patient",
			"permission-user",
		},
	})
}

// listFhirClients returns all registered FHIR clients (admin only).
func listFhirClients(c *gin.Context) {
	rows, err := model.Queries.ListFhirClients(c.Request.Context())
	if err != nil {
		log.Printf("listFhirClients: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}
	if rows == nil {
		rows = make([]dbgen.ListFhirClientsRow, 0)
	}
	c.JSON(http.StatusOK, rows)
}

// getFhirClientPublic returns the non-secret part of a client registration so
// the consent screen can name the app and show which of the scopes it is asking
// for are actually registered to it.
//
// Only the fields the consent screen renders are returned — in particular the
// stored bcrypt credential hash (which `SELECT *` does fetch into the struct)
// must never leave the process, for the same reason ListFhirClients keeps it
// out of its column list. An unknown or inactive client is a 404, matching the
// query's `AND active = 1` filter; the consent UI mirrors that and refuses to
// authorize.
func getFhirClientPublic(c *gin.Context) {
	clientID := strings.TrimSpace(c.Param("client_id"))
	if clientID == "" {
		common.ErrorResponse(c, http.StatusBadRequest, "client_id is required")
		return
	}

	client, err := fhirGetClientByID(c.Request.Context(), clientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			common.ErrorResponse(c, http.StatusNotFound, "unknown or inactive client")
			return
		}
		log.Printf("getFhirClientPublic: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"client_id":       client.ClientID,
		"client_name":     client.ClientName,
		"scopes":          client.Scopes,
		"is_confidential": client.IsConfidential,
	})
}

// createFhirClientRequest is the JSON input for creating a FHIR client.
type createFhirClientRequest struct {
	ClientName     string `json:"client_name" binding:"required"`
	RedirectUris   string `json:"redirect_uris" binding:"required"`
	GrantTypes     string `json:"grant_types"`
	Scopes         string `json:"scopes"`
	IsConfidential bool   `json:"is_confidential"`
}

// generateFhirClientSecret returns a 256-bit URL-safe random client secret.
func generateFhirClientSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// createFhirClient registers a new FHIR client (admin only).
//
// A confidential client gets a freshly generated client secret. Only the bcrypt
// hash is stored; the plaintext is returned exactly once, in this response.
func createFhirClient(c *gin.Context) {
	var in createFhirClientRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		common.ErrorResponseFromError(c, http.StatusBadRequest, err)
		return
	}

	// Generate a UUID for the client_id
	clientID := uuid.New().String()

	// Set defaults
	if in.GrantTypes == "" {
		in.GrantTypes = "authorization_code"
	}
	if in.Scopes == "" {
		in.Scopes = "launch patient/*.read openid fhirUser"
	}

	var clientSecret string
	var clientSecretHash string
	if in.IsConfidential {
		secret, err := generateFhirClientSecret()
		if err != nil {
			log.Printf("createFhirClient: failed to generate client secret: %v", err)
			common.ErrorResponse(c, http.StatusInternalServerError, "server error")
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("createFhirClient: failed to hash client secret: %v", err)
			common.ErrorResponse(c, http.StatusInternalServerError, "server error")
			return
		}
		clientSecret = secret
		clientSecretHash = string(hash)
	}

	result, err := model.Queries.CreateFhirClient(c.Request.Context(), dbgen.CreateFhirClientParams{
		ClientID:         clientID,
		ClientName:       in.ClientName,
		RedirectUris:     in.RedirectUris,
		GrantTypes:       in.GrantTypes,
		Scopes:           in.Scopes,
		IsConfidential:   in.IsConfidential,
		ClientSecretHash: clientSecretHash,
	})
	if err != nil {
		log.Printf("createFhirClient: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	resp := gin.H{
		"id":        newID,
		"client_id": clientID,
	}
	if clientSecret != "" {
		resp["client_secret"] = clientSecret
		resp["client_secret_note"] = "This secret is shown once and cannot be retrieved. Store it now."
	}
	c.JSON(http.StatusCreated, resp)
}

// deactivateFhirClient disables a FHIR client (admin only).
func deactivateFhirClient(c *gin.Context) {
	id := common.ParseInt(c.Param("id"))
	if id == 0 {
		common.ErrorResponse(c, http.StatusBadRequest, "bad request")
		return
	}

	err := model.Queries.DeactivateFhirClient(c.Request.Context(), id)
	if err != nil {
		log.Printf("deactivateFhirClient: %v", err)
		if err == sql.ErrNoRows {
			common.ErrorResponse(c, http.StatusNotFound, "client not found")
			return
		}
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deactivated"})
}
