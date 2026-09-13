package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	ginjwt "github.com/appleboy/gin-jwt/v2"
	"github.com/freemed/freemed-server/config"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
	jwtlib "github.com/golang-jwt/jwt/v4"
)

// Gin context keys written by FHIRAuth and read by the FHIR authorization
// helpers (fhirPatientAllowed / fhirScopeAllows / fhirRequireReadScope).
const (
	fhirCtxPatientContext = "fhir_patient_context"
	fhirCtxScope          = "fhir_scope"
	fhirCtxTokenClass     = "fhir_token_class"
)

// Credential classes recorded in the Gin context by FHIRAuth.
const (
	// fhirClassStaff is a validated staff session cookie. Staff are not
	// patient-compartment- or scope-limited in this deployment.
	fhirClassStaff = "staff"
	// fhirClassSMART is a delegated SMART bearer access token. Such a token
	// is always confined to its launch patient and to its granted scopes.
	fhirClassSMART = "smart"
)

// fhirForbiddenMessage is deliberately identical for a wrong patient
// compartment and for a resource outside it, so the response does not reveal
// whether the requested record exists.
const fhirForbiddenMessage = "Request is outside the patient compartment authorized for this credential"

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

			// Set claims in context for downstream handlers.
			// The value MUST be gin-jwt's MapClaims: ExtractClaims (used by
			// common.GetSession / common.GetClaim / RequireRole) does a bare type
			// assertion to THAT named type, and gin-jwt's MapClaims is a distinct
			// defined type from golang-jwt's, so storing jwtlib.MapClaims here
			// panics any handler that reads the session.
			c.Set("JWT_PAYLOAD", ginjwt.MapClaims(claims))

			// Mark this credential as a delegated SMART token and record the
			// scopes it was issued with, so the authorization helpers below can
			// confine it. Before this, fhir_patient_context had no reader at
			// all and the `scope` claim was never parsed: a token minted for
			// patient 42 could read patient 43, and any token could read
			// anything the route exposed.
			c.Set(fhirCtxTokenClass, fhirClassSMART)
			c.Set(fhirCtxScope, fhirClaimString(claims, "scope"))

			// Set the patient compartment if this token was launched for one.
			//
			// The stored row is authoritative (it is what the token was issued
			// with, and the token cannot be used at all unless the row exists),
			// so the `patient` claim is deliberately NOT used as a fallback: a
			// token whose stored patient is 0 gets no compartment and is
			// therefore denied every patient-data route.
			if dbToken.PatientID > 0 {
				c.Set(fhirCtxPatientContext, dbToken.PatientID)
			}

			c.Next()
			return
		}

		// Fall back to the staff session cookie.
		//
		// This is a STAFF session, so it is validated with the staff key and must
		// be a staff-class token. Before the signing keys were split per credential
		// domain, this branch validated with the shared key and checked nothing
		// else, so a patient's own portal token replayed in the `jwt` cookie
		// authenticated as a staff FHIR session (proven live: a portal token got
		// HTTP 200 with PHI from /api/fhir/Patient/5).
		claims, err := staffSessionFromCookie(c)
		if err == nil {
			c.Set("JWT_PAYLOAD", ginjwt.MapClaims(claims))
			// Staff sessions are not compartment-limited in this deployment.
			c.Set(fhirCtxTokenClass, fhirClassStaff)
			c.Next()
			return
		}

		// No valid auth found
		fhirError(c, http.StatusUnauthorized, "error", "login", "Authentication required")
	}
}

// ============================================================================
// Authorization helpers (H3)
//
// One write and zero readers is how this went wrong before: FHIRAuth recorded
// fhir_patient_context and nothing ever consulted it. These helpers are the
// readers, and every FHIR route that exposes patient data calls one of them.
// ============================================================================

// fhirIsStaffSession reports whether the request was authenticated with a staff
// session cookie. Staff are not confined to a patient compartment nor to a
// scope in this deployment; delegated SMART tokens always are.
func fhirIsStaffSession(c *gin.Context) bool {
	v, ok := c.Get(fhirCtxTokenClass)
	if !ok {
		return false
	}
	s, ok := v.(string)
	return ok && s == fhirClassStaff
}

// fhirPatientContext returns the patient compartment bound to the credential on
// this request and whether one is in force at all.
func fhirPatientContext(c *gin.Context) (int64, bool) {
	v, ok := c.Get(fhirCtxPatientContext)
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	if !ok || id <= 0 {
		return 0, false
	}
	return id, true
}

// fhirPatientAllowed reports whether the credential on this request may touch
// patientID.
//
//   - A delegated SMART token with a launch patient: allowed only for that
//     patient. Any other patient — and a request with no patient at all — is
//     denied.
//   - A delegated SMART token with no launch patient (e.g. a
//     client_credentials token, which has no patient compartment): denied. Fail
//     closed rather than fall through to "all patients".
//   - A staff session: allowed (staff are not compartment-limited here).
//   - Anything else (no credential class in the context): denied.
func fhirPatientAllowed(c *gin.Context, patientID int64) bool {
	if ctxPatient, ok := fhirPatientContext(c); ok {
		return patientID > 0 && patientID == ctxPatient
	}
	return fhirIsStaffSession(c)
}

// fhirPatientScope resolves the effective patient filter for a search route.
// requested is the ?patient parameter value (0 when absent). ok is false when
// the request must be rejected because it lies outside the credential's
// compartment.
//
// For a delegated SMART token this collapses to the launch patient: naming a
// different patient is denied, and naming none filters on the launch patient
// instead of returning every patient's data.
func fhirPatientScope(c *gin.Context, requested int64) (int64, bool) {
	if ctxPatient, ok := fhirPatientContext(c); ok {
		if requested > 0 && requested != ctxPatient {
			return 0, false
		}
		return ctxPatient, true
	}
	if !fhirIsStaffSession(c) {
		return 0, false
	}
	return requested, true
}

// fhirRequirePatient resolves the effective patient filter for a search route
// and writes a 403 OperationOutcome when the request is outside the
// credential's compartment.
func fhirRequirePatient(c *gin.Context, requested int64) (int64, bool) {
	patientID, ok := fhirPatientScope(c, requested)
	if !ok {
		fhirForbidden(c)
		return 0, false
	}
	return patientID, true
}

// fhirRequestedPatient parses the ?patient search parameter. ok is false when
// the value is present but is not a positive integer; the 400 response has
// already been written in that case.
func fhirRequestedPatient(c *gin.Context, raw string) (int64, bool) {
	if strings.TrimSpace(raw) == "" {
		return 0, true
	}
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		fhirError(c, http.StatusBadRequest, "error", "value", "Invalid patient parameter")
		return 0, false
	}
	return id, true
}

// fhirRequireOwnedResource enforces that a resource read by its own id
// (Condition/:id, Encounter/:id, …) belongs to a patient the credential may
// see. It writes a 403 OperationOutcome and returns false when it does not.
func fhirRequireOwnedResource(c *gin.Context, ownerPatientID int64) bool {
	if fhirPatientAllowed(c, ownerPatientID) {
		return true
	}
	fhirForbidden(c)
	return false
}

// fhirForbidden writes the uniform 403 FHIR OperationOutcome used for every
// compartment violation.
func fhirForbidden(c *gin.Context) {
	fhirError(c, http.StatusForbidden, "error", "forbidden", fhirForbiddenMessage)
}

// fhirScopeAllows reports whether the credential on this request grants the
// permission named by need.
//
// need is a bare verb ("read") or a resource-qualified verb
// ("Observation.read"). A staff session is not scope-limited; a delegated SMART
// token must carry a matching scope, and an absent or empty `scope` claim
// grants nothing (fail closed).
func fhirScopeAllows(c *gin.Context, need string) bool {
	v, ok := c.Get(fhirCtxScope)
	if !ok {
		return fhirIsStaffSession(c)
	}
	scopes, ok := v.(string)
	if !ok || strings.TrimSpace(scopes) == "" {
		return fhirIsStaffSession(c)
	}
	return fhirScopeStringAllows(scopes, need)
}

// fhirScopeStringAllows matches need against a space-separated SMART scope
// string. Accepted forms are the ones a SMART server issues:
//
//	*.read  patient/*.read  patient/Observation.read  user/Observation.read
//	system/*.read  openid  fhirUser  launch  launch/patient
//
// Tokens without a verb (openid, fhirUser, launch, launch/patient) grant no
// FHIR access. A wildcard resource ('*') matches any resource.
func fhirScopeStringAllows(scopes, need string) bool {
	needResource, needVerb := "", need
	if i := strings.LastIndex(need, "."); i >= 0 {
		needResource, needVerb = need[:i], need[i+1:]
	}
	if needVerb == "" {
		return false
	}

	for _, field := range strings.Fields(scopes) {
		token := field
		// Strip the SMART context prefix (patient/…, user/…, system/…).
		if i := strings.Index(token, "/"); i >= 0 {
			token = token[i+1:]
		}
		dot := strings.LastIndex(token, ".")
		if dot <= 0 || dot == len(token)-1 {
			continue // no verb part: grants nothing
		}
		resource, verb := token[:dot], token[dot+1:]
		if verb != needVerb && verb != "*" {
			continue
		}
		if needResource == "" || resource == "*" || resource == needResource {
			return true
		}
	}
	return false
}

// fhirRequireReadScope is the route-group guard that enforces the SMART `scope`
// claim on every FHIR route that can return patient data.
//
// It is attached to the whole group rather than repeated inside each handler so
// that a newly added route cannot silently skip the check; /metadata is
// registered before it and is the single documented exemption. It checks for a
// read verb (the wildcard read scopes this server advertises, e.g.
// `patient/*.read` or `system/*.read`), which is the granularity the audit
// asked for; per-resource narrowing is a follow-up.
func fhirRequireReadScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !fhirScopeAllows(c, "read") {
			fhirError(c, http.StatusForbidden, "error", "forbidden",
				"Credential does not grant a FHIR read scope")
			return
		}
		c.Next()
	}
}

// ============================================================================
// Staff session resolution
// ============================================================================

// staffSessionFromCookie resolves the staff session carried by the `jwt`
// cookie. The /oauth2 group has no staff middleware attached — and must not
// grow one, since it also serves the unauthenticated token and introspect
// endpoints — so the authorization endpoint resolves the resource owner itself
// with this.
func staffSessionFromCookie(c *gin.Context) (jwtlib.MapClaims, error) {
	cookie, err := c.Cookie("jwt")
	if err != nil || cookie == "" {
		return nil, fmt.Errorf("no jwt session cookie present")
	}
	return ValidateStaffSessionJWT(cookie)
}

// fhirClaimString renders a claim as a string ("" when absent).
func fhirClaimString(claims jwtlib.MapClaims, key string) string {
	v, ok := claims[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// fhirClaimInt64 converts an identity claim to int64 without panicking.
// JSON numbers decode as float64, but a signer may also use a string or
// json.Number; common.GetSession learned this the hard way (a bare type
// assertion there 500'd every handler reading the session).
func fhirClaimInt64(claims jwtlib.MapClaims, key string) (int64, error) {
	v, ok := claims[key]
	if !ok {
		return 0, fmt.Errorf("claim %q not found", key)
	}
	switch t := v.(type) {
	case float64:
		return int64(t), nil
	case int64:
		return t, nil
	case int:
		return int64(t), nil
	case json.Number:
		n, err := t.Int64()
		if err != nil {
			return 0, fmt.Errorf("claim %q is not an integer: %w", key, err)
		}
		return n, nil
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("claim %q is not an integer: %w", key, err)
		}
		return n, nil
	default:
		return 0, fmt.Errorf("claim %q has unusable type %T", key, v)
	}
}

// GenerateFHIRJWT creates a signed JWT with the given claims using the FHIR
// signing key. FHIR access tokens are a distinct credential domain, so they must
// not verify against — nor be verifiable by — the staff or portal keys.
func GenerateFHIRJWT(claims map[string]interface{}) (string, error) {
	jwtClaims := jwtlib.MapClaims(claims)
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, jwtClaims)
	return token.SignedString(config.Config.DomainKey(config.KeyDomainFHIR))
}

// ValidateFHIRJWT parses and validates a SMART on FHIR access token using the
// FHIR signing key. Returns the claims on success.
func ValidateFHIRJWT(tokenString string) (jwtlib.MapClaims, error) {
	return validateHMACJWT(tokenString, config.Config.DomainKey(config.KeyDomainFHIR))
}

// ValidateStaffSessionJWT validates a staff (provider) session token with the
// STAFF signing key and additionally requires it to be a staff-class token.
// The `jwt` cookie carries a staff session, so a patient-portal token or an
// OAuth2/SMART token must never authenticate through this path.
func ValidateStaffSessionJWT(tokenString string) (jwtlib.MapClaims, error) {
	claims, err := validateHMACJWT(tokenString, config.Config.DomainKey(config.KeyDomainStaff))
	if err != nil {
		return nil, err
	}

	// A staff session always carries the staff identity claim.
	if _, ok := claims["id"]; !ok {
		return nil, fmt.Errorf("not a staff session token: missing id claim")
	}
	// ...and never carries a patient/portal role marker.
	if role, _ := claims["role"].(string); role == "patient" {
		return nil, fmt.Errorf("not a staff session token: role is patient")
	}
	if portal, ok := claims["portal"].(bool); ok && portal {
		return nil, fmt.Errorf("not a staff session token: portal token")
	}

	return claims, nil
}

// validateHMACJWT parses a token and verifies its HS256 signature against key,
// rejecting non-HMAC algorithms and expired tokens.
func validateHMACJWT(tokenString string, key []byte) (jwtlib.MapClaims, error) {
	token, err := jwtlib.Parse(tokenString, func(token *jwtlib.Token) (interface{}, error) {
		// Validate the signing algorithm
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
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
