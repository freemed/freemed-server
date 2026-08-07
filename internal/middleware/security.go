package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders returns a Gin middleware that sets security-related HTTP
// response headers on every request.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "0")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		csp := "script-src 'self' https://cdn.jsdelivr.net 'unsafe-inline' 'unsafe-eval'; " +
			"style-src 'self' https://cdn.jsdelivr.net 'unsafe-inline'; " +
			"img-src 'self' data: blob:; " +
			"font-src 'self' https://cdn.jsdelivr.net; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'"
		c.Header("Content-Security-Policy", csp)

		c.Next()
	}
}
