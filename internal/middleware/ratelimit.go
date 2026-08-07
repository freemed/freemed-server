// Package middleware provides HTTP middleware for the FreeMED EMR server.
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// tokenBucket implements a simple token-bucket rate limiter.
type tokenBucket struct {
	tokens    int
	lastFill  time.Time
}

// LoginRateLimiter is a token-bucket rate limiter scoped by client IP,
// intended for use on login endpoints to prevent brute-force attacks.
type LoginRateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*tokenBucket
	maxPerIP int
	window   time.Duration
}

// NewLoginRateLimiter creates a new LoginRateLimiter.
//
//	maxPerIP: maximum attempts allowed per IP within the window.
//	window:   the sliding window duration.
func NewLoginRateLimiter(maxPerIP int, window time.Duration) *LoginRateLimiter {
	return &LoginRateLimiter{
		buckets:  make(map[string]*tokenBucket),
		maxPerIP: maxPerIP,
		window:   window,
	}
}

// Middleware returns a gin.HandlerFunc that rate-limits requests by client IP
// before forwarding them to the next handler in the Gin chain.
// It uses c.ClientIP() for accurate IP extraction behind proxies and returns
// a 429 JSON response when the limit is exceeded.
//
// Usage:
//
//	auth.POST("/login", limiter.Middleware(), loginHandler)
func (rl *LoginRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many login attempts, try again later",
			})
			return
		}
		c.Next()
	}
}

// allow checks whether the given IP is within its rate limit.
// It must be called while rl.mu is NOT held (it acquires the lock internally).
func (rl *LoginRateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[ip]
	now := time.Now()

	if !exists {
		// First request from this IP: create a full bucket.
		rl.buckets[ip] = &tokenBucket{
			tokens:   rl.maxPerIP - 1,
			lastFill: now,
		}
		return true
	}

	// Refill tokens based on elapsed time.
	elapsed := now.Sub(bucket.lastFill)
	refillRate := float64(rl.maxPerIP) / rl.window.Seconds()
	newTokens := int(elapsed.Seconds() * refillRate)
	if newTokens > 0 {
		bucket.tokens += newTokens
		if bucket.tokens > rl.maxPerIP {
			bucket.tokens = rl.maxPerIP
		}
		bucket.lastFill = now
	}

	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

// Cleanup periodically removes expired entries from the bucket map to
// prevent unbounded memory growth. It should be run in a background goroutine.
//
//	interval: how often to run the cleanup pass.
func (rl *LoginRateLimiter) Cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		// Remove entries that haven't been used in 2× the rate window.
		expiry := rl.window * 2
		for ip, bucket := range rl.buckets {
			if now.Sub(bucket.lastFill) > expiry {
				delete(rl.buckets, ip)
			}
		}
		rl.mu.Unlock()
	}
}
