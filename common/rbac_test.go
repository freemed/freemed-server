package common

import (
	"net/http"
	"net/http/httptest"
	"testing"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRequireRole_Allowed(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set user_type=admin, allowed=admin -> should get 200
	c.Set("JWT_PAYLOAD", jwt.MapClaims{"user_type": "admin"})

	handler := RequireRole("admin")
	handler(c)

	// Status should be 200 (no abort called)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	// Verify c.IsAborted() is false (Next was called, not Abort)
	if c.IsAborted() {
		t.Error("expected context NOT to be aborted")
	}
}

func TestRequireRole_Denied(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set user_type=user, allowed=admin -> should get 403
	c.Set("JWT_PAYLOAD", jwt.MapClaims{"user_type": "user"})

	handler := RequireRole("admin")
	handler(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
	if !c.IsAborted() {
		t.Error("expected context to be aborted")
	}
}

func TestRequireRole_MultipleRoles(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set user_type=user, allowed=admin,user -> should get 200
	c.Set("JWT_PAYLOAD", jwt.MapClaims{"user_type": "user"})

	handler := RequireRole("admin", "user")
	handler(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if c.IsAborted() {
		t.Error("expected context NOT to be aborted")
	}
}

func TestRequireRole_NoClaim(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// No JWT_PAYLOAD set -> user_type = "" -> not in allowed list -> 403
	handler := RequireRole("admin")
	handler(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
	if !c.IsAborted() {
		t.Error("expected context to be aborted")
	}
}
