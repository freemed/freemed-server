package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ginjwt "github.com/appleboy/gin-jwt/v2"
	"github.com/freemed/freemed-server/common"
	"github.com/gin-gonic/gin"
)

// claimsRouterForRole builds the REAL claims route group from common.ApiMap (so
// these tests exercise api/claims.go's registration, not a copy of it) behind a
// middleware that installs the JWT claim payload gin-jwt would set for a staff
// session with the given user_type. An empty role means "no session at all".
func claimsRouterForRole(role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if role != "" {
			c.Set("JWT_PAYLOAD", ginjwt.MapClaims{
				"id":        float64(7),
				"user_type": role,
			})
		}
		c.Next()
	})
	common.ApiMap["claims"].RouterFunction(r.Group("/api/claims"))
	return r
}

// TestUpdateClaimStatusRequiresAdmin pins the M11 fix.
//
// PUT /api/claims/:id/status was registered with no guard, so any authenticated
// user - a nurse, a front-desk account, anyone - could mutate claim state. It was
// unguarded at base revision 873e1a too; the audit surfaced it, the fix adds
// common.RequireRole("admin").
//
// This asserts the NEGATIVE cases only, deliberately: a 403 proves the guard is
// in the chain, and if the guard is removed the request reaches the handler
// instead (which then fails on the absent DB, never returning 403), so the test
// still fails. That keeps the guard pinned without needing a database.
func TestUpdateClaimStatusRequiresAdmin(t *testing.T) {
	cases := []struct {
		name string
		role string
	}{
		{"non-admin staff account", "nurse"},
		{"provider account", "provider"},
		{"billing-adjacent account", "billing"},
		{"no session at all", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := claimsRouterForRole(tc.role)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/api/claims/1/status", nil)
			req.Header.Set("Content-Type", "application/json")

			func() {
				defer func() {
					// The unguarded path reaches updateClaimStatus with no database
					// configured and may panic; that is still "the guard is gone",
					// so swallow it and let the status assertion below decide.
					_ = recover()
				}()
				r.ServeHTTP(w, req)
			}()

			if w.Code != http.StatusForbidden {
				t.Fatalf("PUT /api/claims/:id/status as %s: status = %d, want %d "+
					"(403 means the admin guard is in the chain; anything else means "+
					"the route is unguarded)", tc.name, w.Code, http.StatusForbidden)
			}
		})
	}
}

// TestClaimRoutesGuardingIsUnchanged pins the neighbouring routes so a future
// edit that shuffles the claims group cannot silently drop their guards either.
func TestClaimRoutesGuardingIsUnchanged(t *testing.T) {
	// These two were already admin-only before the audit and must stay that way.
	for _, path := range []string{"/api/claims/generate", "/api/claims/1/x12"} {
		t.Run(path, func(t *testing.T) {
			r := claimsRouterForRole("nurse")
			method := http.MethodGet
			if path == "/api/claims/generate" {
				method = http.MethodPost
			}
			w := httptest.NewRecorder()
			req := httptest.NewRequest(method, path, nil)

			func() {
				defer func() { _ = recover() }()
				r.ServeHTTP(w, req)
			}()

			if w.Code != http.StatusForbidden {
				t.Fatalf("%s %s as nurse: status = %d, want %d", method, path, w.Code, http.StatusForbidden)
			}
		})
	}
}
