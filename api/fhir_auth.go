package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/freemed/freemed-server/config"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
	jwtlib "github.com/golang-jwt/jwt/v4"
)

// FHIRAuth returns a Gin middleware that validates either:
//   - Authorization: Bearer <token> header (SMART on FHIR OAuth2)
//   - jwt cookie (existing JWT-based session auth)
//
// For bearer tokens, it validates the JWT signature, checks against
// fhir_access_token table, and sets claims in the Gin context.
// For cookie-based auth, it extracts JWT claims using the same key.
// On failure, returns 401 with a FHIR OperationOutcome.
func FHIRAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try bearer token first
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" {
				fhirError(c, http.StatusUnauthorized, "error", "login", "Missing bearer token")
				return
			}

			// Validate the JWT
			claims, err := ValidateFHIRJWT(tokenString)
			if err != nil {
				log.Printf("FHIRAuth: bearer token validation failed: %v", err)
				fhirError(c, http.StatusUnauthorized, "error", "login", "Invalid or expired bearer token")
				return
			}

			// Check against fhir_access_token table
			tokenHash := sha256Hash(tokenString)
			dbToken, err := model.Queries.GetFhirAccessToken(c.Request.Context(), tokenHash)
			if err != nil {
				log.Printf("FHIRAuth: token not found in DB: %v", err)
				fhirError(c, http.StatusUnauthorized, "error", "login", "Token not found or expired")
				return
			}

			// Set claims in context for downstream handlers
			c.Set("JWT_PAYLOAD", claims)

			// Set patient context if scoped
			if dbToken.PatientID > 0 {
				c.Set("fhir_patient_context", dbToken.PatientID)
			}

			c.Next()
			return
		}

		// Fall back to jwt cookie (existing session auth)
		cookie, err := c.Cookie("jwt")
		if err == nil && cookie != "" {
			claims, err := ValidateFHIRJWT(cookie)
			if err != nil {
				log.Printf("FHIRAuth: cookie JWT validation failed: %v", err)
				fhirError(c, http.StatusUnauthorized, "error", "login", "Invalid or expired session")
				return
			}

			c.Set("JWT_PAYLOAD", claims)
			c.Next()
			return
		}

		// No valid auth found
		fhirError(c, http.StatusUnauthorized, "error", "login", "Authentication required")
	}
}

// GenerateFHIRJWT creates a signed JWT with the given claims using the
// server's session signing key.
func GenerateFHIRJWT(claims map[string]interface{}) (string, error) {
	jwtClaims := jwtlib.MapClaims(claims)
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, jwtClaims)
	return token.SignedString([]byte(config.Config.Session.Key))
}

// ValidateFHIRJWT parses and validates a JWT token string using the
// server's session signing key. Returns the claims on success.
func ValidateFHIRJWT(tokenString string) (jwtlib.MapClaims, error) {
	token, err := jwtlib.Parse(tokenString, func(token *jwtlib.Token) (interface{}, error) {
		// Validate the signing algorithm
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Config.Session.Key), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	claims, ok := token.Claims.(jwtlib.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid JWT claims")
	}

	// Verify expiry
	if err := claims.Valid(); err != nil {
		return nil, fmt.Errorf("JWT validation failed: %w", err)
	}

	return claims, nil
}
