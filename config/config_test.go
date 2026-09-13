package config

import (
	"bytes"
	"strings"
	"testing"
)

// TestDomainKeyIsDistinctPerDomain is the core property that makes token
// confusion impossible: a token signed with one domain's key must not verify
// against another's. Before this split every credential domain shared one key,
// so a patient's portal token authenticated as a staff FHIR session.
func TestDomainKeyIsDistinctPerDomain(t *testing.T) {
	c := &AppConfig{}
	c.Session.Key = strings.Repeat("k", 48)

	domains := []string{KeyDomainStaff, KeyDomainPortal, KeyDomainFHIR, KeyDomainOAuth2}
	seen := make(map[string]string, len(domains))
	for _, d := range domains {
		k := c.DomainKey(d)
		if len(k) != 32 {
			t.Errorf("DomainKey(%q) returned %d bytes, want 32 (SHA-256)", d, len(k))
		}
		if prev, dup := seen[string(k)]; dup {
			t.Errorf("DomainKey(%q) collides with DomainKey(%q)", d, prev)
		}
		seen[string(k)] = d
		if bytes.Equal(k, []byte(c.Session.Key)) {
			t.Errorf("DomainKey(%q) returned the master secret itself", d)
		}
	}
}

// TestDomainKeyIsDeterministic pins the property that makes this safe to deploy:
// restarting the server must not invalidate live sessions.
func TestDomainKeyIsDeterministic(t *testing.T) {
	c := &AppConfig{}
	c.Session.Key = strings.Repeat("k", 48)

	if !bytes.Equal(c.DomainKey(KeyDomainStaff), c.DomainKey(KeyDomainStaff)) {
		t.Error("DomainKey is not deterministic across calls")
	}

	// A different master secret must produce different domain keys.
	other := &AppConfig{}
	other.Session.Key = strings.Repeat("j", 48)
	if bytes.Equal(c.DomainKey(KeyDomainStaff), other.DomainKey(KeyDomainStaff)) {
		t.Error("DomainKey ignores the master secret")
	}
}

// TestValidateStartupRejectsWeakKeys covers the fail-fast contract: these
// conditions previously only produced a log warning while the server kept
// running with a key anyone could find in the repository.
func TestValidateStartupRejectsWeakKeys(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"unset", "", true},
		{"built-in default", "freemed", true},
		{"too short", strings.Repeat("a", minSessionKeyLen-1), true},
		{"exactly at the floor", strings.Repeat("a", minSessionKeyLen), false},
		{"strong random-looking", "YXJiaXRyYXJpbHkgbG9uZyBkZXYga2V5IGZvciB0ZXN0aW5n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &AppConfig{}
			c.Session.Key = tt.key
			err := c.ValidateStartup()
			if tt.wantErr && err == nil {
				t.Errorf("ValidateStartup() = nil, want an error for key %q", tt.key)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateStartup() = %v, want nil for key %q", err, tt.key)
			}
		})
	}
}

// TestValidateProductionNoLongerWarnsAboutTheKey documents the move of the key
// check from advisory to fatal: it must not appear in both places.
func TestValidateProductionNoLongerWarnsAboutTheKey(t *testing.T) {
	c := &AppConfig{}
	c.Session.Key = defaultSessionKey
	for _, w := range c.ValidateProduction() {
		if strings.Contains(w, "SESSION_KEY") || strings.Contains(w, "session key") {
			t.Errorf("ValidateProduction still warns about the session key (%q); it is now fatal in ValidateStartup", w)
		}
	}
}
