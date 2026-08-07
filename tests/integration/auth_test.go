//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

const authServerURL = "http://localhost:3000"

// TestTokenBlacklist tests the full token blacklist flow:
//  1. Insert a test user into the user table with a known bcrypt password.
//  2. Login via POST /auth/login (with CSRF cookie+header) and capture the jwt cookie.
//  3. Call GET /auth/me with the cookie to verify authenticated.
//  4. Call DELETE /auth/logout with the cookie.
//  5. Call GET /auth/me again to verify 401 (token blacklisted).
func TestTokenBlacklist(t *testing.T) {
	db := openDB(t)
	defer db.Close()

	// Generate bcrypt hash for known test password.
	const testPassword = "testpassword123"
	hash, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword: %v", err)
	}

	// Insert a test user with the bcrypt hash in userpassword_bcrypt column.
	now := time.Now()
	username := fmt.Sprintf("authtest_%d", now.UnixNano())
	_, err = db.Exec(
		"INSERT INTO user (created_at, updated_at, username, userpassword_bcrypt, userfname, userlname, usertype, userrealphy) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		now, now, username, string(hash), "Auth", "Test", "admin", 1,
	)
	if err != nil {
		t.Fatalf("INSERT test user: %v", err)
	}

	// Fetch the user ID for cleanup.
	var userID int64
	err = db.QueryRow("SELECT id FROM user WHERE username = ?", username).Scan(&userID)
	if err != nil {
		t.Fatalf("SELECT user id: %v", err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM user WHERE id = ?", userID)
	})
	t.Logf("Created test user: id=%d username=%s", userID, username)

	client := &http.Client{
		// Do not follow redirects – we want to inspect 303/307 codes directly.
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// --- Step 1: Get CSRF token ---
	csrfCookie, csrfToken, err := getCSRFToken(client)
	if err != nil {
		t.Fatalf("getCSRFToken: %v", err)
	}
	t.Logf("Got CSRF token: %s", csrfToken)

	// --- Step 2: Login ---
	loginPayload := map[string]string{
		"username": username,
		"password": testPassword,
	}
	bodyBytes, _ := json.Marshal(loginPayload)
	req, _ := http.NewRequest("POST", authServerURL+"/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrfToken)
	req.Header.Set("Cookie", csrfCookie)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST /auth/login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("POST /auth/login returned %d: %s", resp.StatusCode, string(respBody))
	}

	// Capture the jwt cookie from Set-Cookie header.
	jwtCookie := extractCookie(resp, "jwt")
	if jwtCookie == "" {
		t.Fatal("no jwt cookie in Set-Cookie header of login response")
	}
	t.Logf("Got JWT cookie: %s...", jwtCookie[:min(30, len(jwtCookie))])

	// --- Step 3: Verify authenticated via GET /auth/me ---
	req, _ = http.NewRequest("GET", authServerURL+"/auth/me", nil)
	req.Header.Set("Cookie", "jwt="+jwtCookie)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("GET /auth/me: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("GET /auth/me returned %d, want 200: %s", resp.StatusCode, string(respBody))
	}
	t.Log("GET /auth/me returned 200 — authenticated")

	// Verify the response body contains expected fields.
	var meResponse struct {
		UserID   interface{} `json:"user_id"`
		Username string      `json:"username"`
		UserType string      `json:"user_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meResponse); err == nil {
		t.Logf("auth/me response: user_id=%v username=%s user_type=%s",
			meResponse.UserID, meResponse.Username, meResponse.UserType)
	}
	// Re-read body since we consumed it above
	resp.Body.Close()

	// --- Step 4: Logout via DELETE /auth/logout ---
	req, _ = http.NewRequest("DELETE", authServerURL+"/auth/logout", nil)
	req.Header.Set("Cookie", "jwt="+jwtCookie)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("DELETE /auth/logout: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("DELETE /auth/logout returned %d, want 200: %s", resp.StatusCode, string(respBody))
	}
	t.Log("DELETE /auth/logout returned 200 — logged out")

	// --- Step 5: Verify rejection on /auth/me after logout (token blacklisted) ---
	// The gin-jwt Authorizator returns false for blacklisted tokens, which
	// yields 403 (Forbidden). Missing/invalid tokens yield 401 (Unauthorized).
	// Both codes indicate the token is no longer usable.
	req, _ = http.NewRequest("GET", authServerURL+"/auth/me", nil)
	req.Header.Set("Cookie", "jwt="+jwtCookie)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("GET /auth/me after logout: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("GET /auth/me after logout returned %d, want 401 or 403: %s", resp.StatusCode, string(respBody))
	}
	t.Logf("GET /auth/me after logout returned %d — token blacklisted ✓", resp.StatusCode)
}

// getCSRFToken fetches a CSRF token from /auth/csrf and returns the cookie
// header value and the token string.
func getCSRFToken(client *http.Client) (cookie string, token string, err error) {
	resp, err := client.Get(authServerURL + "/auth/csrf")
	if err != nil {
		return "", "", fmt.Errorf("GET /auth/csrf: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("GET /auth/csrf returned %d: %s", resp.StatusCode, string(body))
	}

	// Parse the JSON response for the token.
	var csrfResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&csrfResp); err != nil {
		return "", "", fmt.Errorf("decoding /auth/csrf response: %w", err)
	}

	// Extract the csrf_token cookie from Set-Cookie.
	cookieVal := extractCookie(resp, "csrf_token")
	if cookieVal == "" {
		return "", "", fmt.Errorf("no csrf_token cookie in response")
	}

	return "csrf_token=" + cookieVal, csrfResp.Token, nil
}

// extractCookie extracts a named cookie value from the Set-Cookie response header.
func extractCookie(resp *http.Response, name string) string {
	for _, c := range resp.Cookies() {
		if c.Name == name {
			return c.Value
		}
	}
	// Fallback: parse the raw Set-Cookie header manually, since
	// http.Cookies() sometimes misses cookies with certain flags.
	prefix := name + "="
	for _, h := range resp.Header["Set-Cookie"] {
		for _, part := range strings.Split(h, ";") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, prefix) {
				return strings.TrimPrefix(part, prefix)
			}
		}
	}
	return ""
}
