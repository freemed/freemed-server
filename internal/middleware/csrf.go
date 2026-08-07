package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

const csrfCookieName = "csrf_token"
const csrfHeaderName = "X-CSRF-Token"

// GenerateCSRF sets a double-submit CSRF cookie and returns the token.
// The cookie is NOT httpOnly so JavaScript can read it.
func GenerateCSRF(c *gin.Context) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate CSRF token",
		})
		return
	}
	tokenStr := hex.EncodeToString(token)

	// Set cookie: readable by JS, SameSite=Strict, not httpOnly
	c.SetCookie(csrfCookieName, tokenStr, 0, "/", "", false, false) // secure=false for dev, SameSite not directly settable via gin's SetCookie
	// Manually set the cookie header with SameSite=Strict
	c.Header("Set-Cookie", csrfCookieName+"="+tokenStr+"; Path=/; SameSite=Strict; Max-Age=3600")

	c.JSON(http.StatusOK, gin.H{
		"token": tokenStr,
	})
}

// ValidateCSRF returns a Gin middleware that validates the CSRF double-submit cookie.
// The X-CSRF-Token header value must match the csrf_token cookie value.
func ValidateCSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookieToken, err := c.Cookie(csrfCookieName)
		if err != nil || cookieToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "CSRF token missing",
				"message": "CSRF token missing — request /auth/csrf first",
			})
			return
		}

		headerToken := c.GetHeader(csrfHeaderName)
		if headerToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "CSRF header missing",
				"message": "X-CSRF-Token header required",
			})
			return
		}

		if cookieToken != headerToken {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "CSRF token mismatch",
				"message": "CSRF token mismatch",
			})
			return
		}

		c.Next()
	}
}
