package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// fhirCred is the credential context FHIRAuth installs on a request. A zero
// value means "nothing was set at all" — the fail-closed case.
type fhirCred struct {
	class     string // fhirClassStaff, fhirClassSMART, or "" for unset
	patientID int64  // 0 means no compartment in force
	scope     string
	setScope  bool
}

func newFhirTestContext(t *testing.T, cred fhirCred, method, target string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, nil)
	if cred.class != "" {
		c.Set(fhirCtxTokenClass, cred.class)
	}
	if cred.patientID > 0 {
		c.Set(fhirCtxPatientContext, cred.patientID)
	}
	if cred.setScope {
		c.Set(fhirCtxScope, cred.scope)
	}
	return c, w
}

func decodeOperationOutcome(t *testing.T, w *httptest.ResponseRecorder) fhirOperationOutcome {
	t.Helper()
	var oo fhirOperationOutcome
	if err := json.Unmarshal(w.Body.Bytes(), &oo); err != nil {
		t.Fatalf("response body is not a FHIR OperationOutcome: %v (body %q)", err, w.Body.String())
	}
	if oo.ResourceType != "OperationOutcome" {
		t.Fatalf("resourceType = %q, want OperationOutcome", oo.ResourceType)
	}
	if len(oo.Issue) == 0 {
		t.Fatal("OperationOutcome carries no issue")
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/fhir+json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want application/fhir+json; charset=utf-8", ct)
	}
	return oo
}

// TestFhirPatientAllowed is the core compartment rule: a delegated SMART token
// may only touch its launch patient, a staff session is not compartment-limited,
// and anything else is denied.
func TestFhirPatientAllowed(t *testing.T) {
	tests := []struct {
		name      string
		cred      fhirCred
		patientID int64
		want      bool
	}{
		{"smart token for 42 reads 42", fhirCred{class: fhirClassSMART, patientID: 42}, 42, true},
		{"smart token for 42 reads 43", fhirCred{class: fhirClassSMART, patientID: 42}, 43, false},
		{"smart token for 42 reads patient 0", fhirCred{class: fhirClassSMART, patientID: 42}, 0, false},
		{"smart token with no compartment reads 43", fhirCred{class: fhirClassSMART}, 43, false},
		{"smart token with no compartment reads 0", fhirCred{class: fhirClassSMART}, 0, false},
		{"staff session reads 43", fhirCred{class: fhirClassStaff}, 43, true},
		{"staff session with explicit compartment still reads 43", fhirCred{class: fhirClassStaff, patientID: 42}, 43, false},
		{"no credential class at all reads 43", fhirCred{}, 43, false},
		{"no credential class at all reads 0", fhirCred{}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := newFhirTestContext(t, tt.cred, http.MethodGet, "/api/fhir/Patient/43")
			if got := fhirPatientAllowed(c, tt.patientID); got != tt.want {
				t.Errorf("fhirPatientAllowed(%d) = %v, want %v", tt.patientID, got, tt.want)
			}
		})
	}
}

// TestFhirPatientScope covers the list-endpoint rule: no ?patient on a
// delegated token means "the launch patient", never "all patients".
func TestFhirPatientScope(t *testing.T) {
	tests := []struct {
		name      string
		cred      fhirCred
		requested int64
		want      int64
		wantOK    bool
	}{
		{"smart token for 42, no filter", fhirCred{class: fhirClassSMART, patientID: 42}, 0, 42, true},
		{"smart token for 42, filter 42", fhirCred{class: fhirClassSMART, patientID: 42}, 42, 42, true},
		{"smart token for 42, filter 43", fhirCred{class: fhirClassSMART, patientID: 42}, 43, 0, false},
		{"smart token with no compartment, no filter", fhirCred{class: fhirClassSMART}, 0, 0, false},
		{"staff, no filter", fhirCred{class: fhirClassStaff}, 0, 0, true},
		{"staff, filter 43", fhirCred{class: fhirClassStaff}, 43, 43, true},
		{"no credential class, no filter", fhirCred{}, 0, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := newFhirTestContext(t, tt.cred, http.MethodGet, "/api/fhir/Observation")
			got, ok := fhirPatientScope(c, tt.requested)
			if ok != tt.wantOK || got != tt.want {
				t.Errorf("fhirPatientScope(%d) = (%d, %v), want (%d, %v)", tt.requested, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

// TestFhirScopeAllows is the scope rule: a delegated token must carry a matching
// scope; a staff session is not scope-limited; a missing claim denies.
func TestFhirScopeAllows(t *testing.T) {
	tests := []struct {
		name string
		cred fhirCred
		need string
		want bool
	}{
		{"smart wildcard patient read", fhirCred{class: fhirClassSMART, scope: "patient/*.read openid fhirUser", setScope: true}, "read", true},
		{"smart bare wildcard read", fhirCred{class: fhirClassSMART, scope: "*.read", setScope: true}, "read", true},
		{"smart system read", fhirCred{class: fhirClassSMART, scope: "system/*.read", setScope: true}, "read", true},
		{"smart resource-qualified read", fhirCred{class: fhirClassSMART, scope: "patient/Observation.read", setScope: true}, "read", true},
		{"smart resource-qualified read, matching resource", fhirCred{class: fhirClassSMART, scope: "patient/Observation.read", setScope: true}, "Observation.read", true},
		{"smart resource-qualified read, other resource", fhirCred{class: fhirClassSMART, scope: "patient/Observation.read", setScope: true}, "Patient.read", false},
		{"smart identity-only scopes (the audit's 'no read verb')", fhirCred{class: fhirClassSMART, scope: "openid fhirUser", setScope: true}, "read", false},
		{"smart launch-only scope", fhirCred{class: fhirClassSMART, scope: "launch launch/patient", setScope: true}, "read", false},
		{"smart write-only scope", fhirCred{class: fhirClassSMART, scope: "patient/*.write", setScope: true}, "read", false},
		{"smart empty scope claim", fhirCred{class: fhirClassSMART, setScope: true}, "read", false},
		{"smart no scope claim at all", fhirCred{class: fhirClassSMART}, "read", false},
		{"staff with no scope claim", fhirCred{class: fhirClassStaff}, "read", true},
		{"staff with an explicit empty scope claim", fhirCred{class: fhirClassStaff, setScope: true}, "read", true},
		{"no credential class", fhirCred{}, "read", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := newFhirTestContext(t, tt.cred, http.MethodGet, "/api/fhir/Observation")
			if got := fhirScopeAllows(c, tt.need); got != tt.want {
				t.Errorf("fhirScopeAllows(%q) = %v, want %v", tt.need, got, tt.want)
			}
		})
	}
}

func TestFhirScopeStringAllows(t *testing.T) {
	tests := []struct {
		scopes string
		need   string
		want   bool
	}{
		{"patient/*.read openid fhirUser", "read", true},
		{"fhirUser openid patient/*.read", "read", true},
		{"openid fhirUser", "read", false},
		{"", "read", false},
		{"launch patient/*.read", "write", false},
		{"patient/*.write", "write", true},
		{"patient/Observation.read", "Observation.read", true},
		{"user/Patient.read", "Patient.read", true},
		{"system/*.read", "read", true},
		{"patient/*.read", "", false},
		{"patient/Observation.", "read", false},
	}
	for _, tt := range tests {
		if got := fhirScopeStringAllows(tt.scopes, tt.need); got != tt.want {
			t.Errorf("fhirScopeStringAllows(%q, %q) = %v, want %v", tt.scopes, tt.need, got, tt.want)
		}
	}
}

// TestFhirPatientGetDeniesForeignPatient drives the real handler: a token
// launched for patient 42 asking for Patient/43 must be refused with a 403
// OperationOutcome, before any database access.
func TestFhirPatientGetDeniesForeignPatient(t *testing.T) {
	c, w := newFhirTestContext(t, fhirCred{class: fhirClassSMART, patientID: 42, scope: "patient/*.read", setScope: true},
		http.MethodGet, "/api/fhir/Patient/43")
	c.Params = gin.Params{{Key: "id", Value: "43"}}

	fhirPatientGet(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 (body %q)", w.Code, w.Body.String())
	}
	oo := decodeOperationOutcome(t, w)
	if oo.Issue[0].Code != "forbidden" {
		t.Errorf("issue code = %q, want forbidden", oo.Issue[0].Code)
	}

	// Same for the C-CDA $document operation.
	c2, w2 := newFhirTestContext(t, fhirCred{class: fhirClassSMART, patientID: 42, scope: "patient/*.read", setScope: true},
		http.MethodGet, "/api/fhir/Patient/43/$document")
	c2.Params = gin.Params{{Key: "id", Value: "43"}}
	fhirPatientDocument(c2)
	if w2.Code != http.StatusForbidden {
		t.Fatalf("$document status = %d, want 403 (body %q)", w2.Code, w2.Body.String())
	}
}

// TestFhirObservationListDeniesForeignPatient drives the real handler: the list
// endpoint must reject a ?patient that is outside the token's compartment, and
// an unparseable one, without touching the database.
func TestFhirObservationListDeniesForeignPatient(t *testing.T) {
	c, w := newFhirTestContext(t, fhirCred{class: fhirClassSMART, patientID: 42, scope: "patient/*.read", setScope: true},
		http.MethodGet, "/api/fhir/Observation?patient=43")
	fhirObservationList(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 (body %q)", w.Code, w.Body.String())
	}
	decodeOperationOutcome(t, w)

	// A non-numeric patient is a 400, not a silent 0.
	c2, w2 := newFhirTestContext(t, fhirCred{class: fhirClassSMART, patientID: 42, scope: "patient/*.read", setScope: true},
		http.MethodGet, "/api/fhir/Observation?patient=abc")
	fhirObservationList(c2)
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body %q)", w2.Code, w2.Body.String())
	}
}

// TestFhirRequireReadScope exercises the group guard in isolation.
func TestFhirRequireReadScope(t *testing.T) {
	tests := []struct {
		name       string
		cred       fhirCred
		wantStatus int
	}{
		{"smart token without a read scope", fhirCred{class: fhirClassSMART, scope: "openid fhirUser", setScope: true}, http.StatusForbidden},
		{"smart token with no scope claim", fhirCred{class: fhirClassSMART}, http.StatusForbidden},
		{"smart token with a read scope", fhirCred{class: fhirClassSMART, scope: "patient/*.read", setScope: true}, http.StatusOK},
		{"staff session", fhirCred{class: fhirClassStaff}, http.StatusOK},
		{"unauthenticated context", fhirCred{}, http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			cred := tt.cred
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if cred.class != "" {
					c.Set(fhirCtxTokenClass, cred.class)
				}
				if cred.patientID > 0 {
					c.Set(fhirCtxPatientContext, cred.patientID)
				}
				if cred.setScope {
					c.Set(fhirCtxScope, cred.scope)
				}
				c.Next()
			})
			g := r.Group("/api/fhir")
			g.Use(fhirRequireReadScope())
			g.GET("/Observation", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/fhir/Observation", nil))
			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body %q)", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantStatus == http.StatusForbidden {
				decodeOperationOutcome(t, w)
			}
		})
	}
}

// TestFhirRoutesAreScopeGuarded enumerates every registered FHIR route and
// proves it is behind fhirRequireReadScope. This is the regression test for the
// "one unchecked handler is the whole vulnerability" problem: a route added
// before the guard fails here. /metadata is the only documented exemption.
func TestFhirRoutesAreScopeGuarded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	origAuth := fhirAuthFunc
	fhirAuthFunc = func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set(fhirCtxTokenClass, fhirClassSMART)
			c.Set(fhirCtxScope, "openid fhirUser") // authenticated, but no read verb
			c.Next()
		}
	}
	t.Cleanup(func() { fhirAuthFunc = origAuth })

	r := gin.New()
	fhirRegisterRoutes(r.Group("/api/fhir"))

	const exemptPath = "/api/fhir/metadata"
	guarded := 0
	for _, rt := range r.Routes() {
		if rt.Method != http.MethodGet {
			continue
		}
		target := strings.ReplaceAll(rt.Path, ":id", "42")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))

		if rt.Path == exemptPath {
			if w.Code != http.StatusOK {
				t.Errorf("exempt route %s = %d, want 200", rt.Path, w.Code)
			}
			continue
		}
		guarded++
		if w.Code != http.StatusForbidden {
			t.Errorf("route %s with a scopeless SMART token = %d, want 403 — it is not behind fhirRequireReadScope", rt.Path, w.Code)
		}
	}
	if guarded != 17 {
		t.Errorf("guarded FHIR route count = %d, want 17 (did a route move above the guard?)", guarded)
	}
}
