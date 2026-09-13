package main

import "testing"

// TestIsJSONSurfacePath pins the M3 routing rule: paths on the machine-readable
// surface must never be answered with the SPA's index.html, while genuine
// client-side routes still get it.
func TestIsJSONSurfacePath(t *testing.T) {
	jsonPaths := []string{
		"/api",
		"/api/",
		"/api/does-not-exist",
		"/api/claims/generate",
		"/auth",
		"/auth/login",
		"/oauth2/token",
		"/.well-known/smart-configuration",
		"/portal",
		"/portal/auth/login",
	}
	for _, p := range jsonPaths {
		if !isJSONSurfacePath(p) {
			t.Errorf("isJSONSurfacePath(%q) = false, want true", p)
		}
	}

	spaPaths := []string{
		"/",
		"/patients/1",
		"/dashboard",
		"/apifoo",  // must not match the /api prefix
		"/authfoo", // must not match the /auth prefix
		"/oauth2foo/bar",
		"/portals",
		"/_app/start.js",
		"/favicon.ico",
	}
	for _, p := range spaPaths {
		if isJSONSurfacePath(p) {
			t.Errorf("isJSONSurfacePath(%q) = true, want false", p)
		}
	}
}
