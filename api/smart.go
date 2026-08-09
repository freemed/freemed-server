package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() {
	common.ApiMap["smart"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.GET("/clients", listFhirClients)
			r.POST("/clients", createFhirClient)
			r.DELETE("/clients/:id", deactivateFhirClient)
		},
	}
}

// smartConfiguration returns the SMART on FHIR metadata at /.well-known/smart-configuration.
// This is registered separately in main.go, not via ApiMap.
func SmartConfiguration(c *gin.Context) {
	issuer := fmt.Sprintf("http://%s", c.Request.Host)
	c.JSON(http.StatusOK, gin.H{
		"issuer":                issuer,
		"authorization_endpoint": fmt.Sprintf("%s/oauth2/authorize", issuer),
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
		rows = make([]dbgen.FhirClient, 0)
	}
	c.JSON(http.StatusOK, rows)
}

// createFhirClientRequest is the JSON input for creating a FHIR client.
type createFhirClientRequest struct {
	ClientName     string `json:"client_name" binding:"required"`
	RedirectUris   string `json:"redirect_uris" binding:"required"`
	GrantTypes     string `json:"grant_types"`
	Scopes         string `json:"scopes"`
	IsConfidential bool   `json:"is_confidential"`
}

// createFhirClient registers a new FHIR client (admin only).
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

	result, err := model.Queries.CreateFhirClient(c.Request.Context(), dbgen.CreateFhirClientParams{
		ClientID:       clientID,
		ClientName:     in.ClientName,
		RedirectUris:   in.RedirectUris,
		GrantTypes:     in.GrantTypes,
		Scopes:         in.Scopes,
		IsConfidential: in.IsConfidential,
	})
	if err != nil {
		log.Printf("createFhirClient: %v", err)
		common.ErrorResponseFromError(c, http.StatusInternalServerError, err)
		return
	}

	newID, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{
		"id":        newID,
		"client_id": clientID,
	})
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
