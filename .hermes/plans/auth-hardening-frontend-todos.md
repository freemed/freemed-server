# Plan: Auth Hardening + Frontend TODO Implementation

## Phase 1: Auth Hardening (Backend + Frontend)

### 1.1 Security Headers Middleware (BACKEND)
**Status:** Missing entirely  
**Effort:** Small (1 file)  
**Risk:** Low

Add a Gin middleware that sets security headers on every response:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 0`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Permissions-Policy: camera=(), microphone=(), geolocation=()`
- `Content-Security-Policy` (tailored for jsDelivr-based Tailwind/Bootstrap)

Wire into main.go BEFORE gzip and route setup, applied to both `/api` and `/auth` groups.

### 1.2 Rate Limiting on /auth/login (BACKEND)
**Status:** Missing  
**Effort:** Small (1 middleware file + 1-2 lines in main.go)  
**Risk:** Low

Add a token-bucket rate limiter middleware for the login endpoint:
- 10 attempts per minute per IP
- Background cleanup goroutine
- Apply to `POST /auth/login`
- Return 429 with JSON error when rate exceeded

### 1.3 CSRF Token for POST /auth/login (BACKEND)
**Status:** Missing  
**Effort:** Medium (cross-cutting: middleware + endpoint + frontend)  
**Risk:** Medium (touches auth flow)

Implementation:
- Backend: Create CSRF middleware that sets a `csrf_token` cookie (SameSite=Strict, httpOnly=false so JS can read it) and validates `X-CSRF-Token` header on POST /auth/login
- Generate token via `crypto/rand` hex string, store in Redis with short TTL
- Validate: extract token from `X-CSRF-Token` header, compare against Redis value
- Frontend: Read `csrf_token` cookie on login page mount, send as `X-CSRF-Token` header in login request

### 1.4 localStorage → httpOnly Cookie Token Storage (FRONTEND + BACKEND)
**Status:** Currently uses `localStorage.getItem('freemed_token')`  
**Effort:** Medium (auth flow rewrite)  
**Risk:** Medium-High (changes auth contract between frontend and backend)

**Approach:**
- Backend already supports `cookie:jwt` in TokenLookup: `"header:Authorization,cookie:jwt"` — GOOD
- On successful login `/auth/login`, respond with `Set-Cookie: jwt={token}; HttpOnly; Secure; SameSite=Strict; Path=/; Max-Age=N` in addition to the JSON response body
- Remove `authToken.current = data.token` from frontend login flow
- The browser will automatically send the cookie with every request — no need to manually attach `Authorization` header
- Frontend: Remove `STORAGE_KEY` / `localStorage` logic from `auth.svelte.ts`
- On logout: server clears the jwt cookie via `Set-Cookie: jwt=; Max-Age=0`

**Pitfalls:**
- CSRF risk with cookie-based auth → mitigated by SameSite=Strict + CSRF token requirement (Phase 1.3)
- The `api.ts` currently reads `authToken.current` and attaches `Authorization: Bearer ...`; with httpOnly cookies, the browser sends the cookie automatically so the Authorization header becomes redundant
- Must keep `Bearer` header support for API clients that don't use cookies

### 1.5 Wire RequireRole() on Admin Endpoints (BACKEND)
**Status:** `RequireRole()` exists but only used in 1 place  
**Effort:** Small (route-level changes)  
**Risk:** Low

Endpoints needing `RequireRole("admin")`:

**api/users.go:**
- POST `/` (create user) → `RequireRole("admin")`
- PUT `/:id` (update user) → `RequireRole("admin")`
- PUT `/:id/password` (change password) → keep unguarded (self-service) or add RequireRole for other users
- DELETE `/:id` (delete user) → `RequireRole("admin")`
- GET `/` (list users) → `RequireRole("admin")`

**api/acl.go:**
- ALL write endpoints: POST/PUT/DELETE on groups, permissions, group-permissions, user-groups → `RequireRole("admin")`
- GET endpoints: list groups/permissions/user-groups → `RequireRole("admin")` (ACL data is sensitive)

**api/config.go:**
- ✅ Already guarded: `r.PUT("/:id", common.RequireRole("admin"), configUpdate)`

**Other admin-like endpoints to review:** `api/facilities.go`, `api/providers.go`, `api/tools.go`

---

## Phase 2: Frontend TODO Items (Prioritized)

### 2.1 Replace window.toast with Svelte Store-Based Toast
**Status:** `window.toast` global pattern exists in legacy code  
**Effort:** Small (1 toast store + replace global usages)  
**Risk:** Low

- Create `$lib/stores/toast.svelte.ts`: reactive store with `show(message, type)` and auto-dismiss
- Create `Toast.svelte` component rendering from store
- Add to `+layout.svelte`
- Grep for `window.toast` / `toast(` usages and replace with `import { toast } from '$lib/stores/toast.svelte'`

### 2.2 Add Offline/Loading/Error States to All Pages
**Status:** Most pages lack these states  
**Effort:** Large (touches ~25 page components)  
**Risk:** Low (purely additive)

Create reusable components:
- `LoadingSpinner.svelte` — centered spinner with optional message
- `ErrorBanner.svelte` — dismissible error with retry button
- `EmptyState.svelte` — "No data found" with optional action
- `OfflineBanner.svelte` — navigator.onLine listener

Apply to each page component's data-fetching lifecycle.

### 2.3 Add admin pages: ACL Management + User Management
**Status:** Backend endpoints exist, frontend pages missing  
**Effort:** Medium (2 new pages + API wiring)  
**Risk:** Low

- `frontend/src/routes/admin/users/+page.svelte` — already exists as placeholder, needs wiring
- ACL management: new route `frontend/src/routes/admin/acl/+page.svelte`
  - Group CRUD
  - Permission CRUD  
  - Group-permission assignment
  - User-group assignment

### 2.4 Add billing pages
**Status:** Backend endpoints exist, frontend placeholder routes exist  
**Effort:** Large (4-5 pages)  
**Risk:** Medium (complex data relationships)

Existing placeholder routes to wire:
- `routes/billing/+page.svelte` — billing dashboard/overview
- `routes/billing/claims/+page.svelte` — claims manager
- `routes/billing/ar/+page.svelte` — AR aging
- `routes/billing/remitt/+page.svelte` — remittances
- `routes/billing/superbills/+page.svelte` — superbills

### 2.5 Add call-in page
**Status:** Backend endpoint exists (`api/callin.go`), frontend missing  
**Effort:** Small (1 new page)  
**Risk:** Low

### 2.6 Schedule Recurring Appointments UI
**Status:** Backend has recurring appointment endpoints in scheduler  
**Effort:** Medium (UI complexity with recurrence patterns)  
**Risk:** Medium

### 2.7 Responsive Audit
**Status:** Not done  
**Effort:** Medium  
**Risk:** Low

Test all ~27 pages at mobile widths (375px), tablet (768px), desktop (1280px). Fix layout issues.

### 2.8 Form Validation Library
**Status:** Missing  
**Effort:** Small (install + adopt incrementally)  
**Risk:** Low

Install `sveltekit-superforms` + `zod`. Apply to login page first, then high-value forms (patient create, appointment create).

### 2.9 E2E Tests
**Status:** Missing  
**Effort:** Large  
**Risk:** Low

Choose Playwright (better SPA support). Start with critical paths: login → patient search → create encounter.

### 2.10 Remove legacy ui/ Directory
**Status:** Blocked on SPA deployment verification  
**Effort:** Small (delete + update README)  
**Risk:** Medium (must verify SPA covers all legacy functionality first)

---

## Phase 3: Backend Security Cleanup (from JWT Anti-Patterns)

### 3.1 Fix `Authorizator` Stub
**Status:** Returns `true` for all authenticated users — mitigated by per-route RequireRole  
**Effort:** Small  
**Risk:** Low

Remove the TODO comment; the current behavior is intentional since RequireRole handles per-route authorization. Add a comment explaining the design: "All authenticated users pass; route-level authorization is handled by RequireRole middleware."

### 3.2 Production Config Warnings
**Status:** Missing  
**Effort:** Small  
**Risk:** Low

Add `ValidateProduction()` to AppConfig that warns on default session key, default DB password, default Redis host. Call at startup.

### 3.3 Error Message Leakage Audit
**Status:** Widespread `log.Print(err.Error()) + r.AbortWithError(http.StatusInternalServerError, err)`  
**Effort:** Medium (touches ~40 handlers)  
**Risk:** Low

`r.AbortWithError(http.StatusInternalServerError, err)` sends the `err.Error()` string in the response body. Replace with `r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})`.

---

## Implementation Order (Dependency Chain)

```
Phase 1 (auth hardening):
  1.1 Security Headers ─── independent
  1.2 Rate Limiting   ─── independent
  1.3 CSRF Token      ─── requires 1.1 (headers for cookie settings)
  1.4 httpOnly Cookie ─── requires 1.3 (CSRF must be in place first)
  1.5 RequireRole     ─── independent

Phase 2 (frontend):
  2.1 Replace toast    ─── independent
  2.9 E2E Tests        ─── independent (can start immediately)
  2.2 Loading states   ─── independent (additive)
  2.3 Admin pages      ─── requires 1.5 (RequireRole on backend)
  2.4 Billing pages    ─── independent
  2.5 Call-in page     ─── independent
  2.8 Form validation  ─── independent
  2.6 Recurring UI     ─── independent
  2.7 Responsive audit ─── best done last (after all pages exist)
  2.10 Remove legacy   ─── blocked on 2.3-2.6 completion

Phase 3 (security cleanup):
  3.1 Authorizator fix ─── independent
  3.2 Config warnings  ─── independent
  3.3 Error leakage    ─── independent
```

## Dispatch Strategy

Phases 1.1, 1.2, 1.5 are fully independent — dispatch as parallel subagents (batch of 3).  
Phase 1.3 and 1.4 are sequential (CSRF first, then cookie).  
Phase 2 items 2.1, 2.5, 2.8 can go in parallel together.  
Phase 2 items 2.3, 2.4 are larger — dispatch as separate subagents.

## Key Files Modified Per Task

| Task | Files |
|------|-------|
| 1.1 Security Headers | `internal/middleware/security.go` (NEW), `cmd/freemed-server/main.go` |
| 1.2 Rate Limiting | `internal/middleware/ratelimit.go` (NEW), `cmd/freemed-server/main.go` |
| 1.3 CSRF | `internal/middleware/csrf.go` (NEW), `cmd/freemed-server/auth.go`, `frontend/src/lib/stores/auth.svelte.ts`, `frontend/src/routes/login/+page.svelte` |
| 1.4 httpOnly Cookie | `cmd/freemed-server/auth.go`, `frontend/src/lib/stores/auth.svelte.ts`, `frontend/src/lib/api.ts`, `cmd/freemed-server/main.go` |
| 1.5 RequireRole | `api/users.go`, `api/acl.go`, `api/facilities.go`, `api/providers.go`, `api/tools.go` |
| 2.1 Toast | `frontend/src/lib/stores/toast.svelte.ts` (NEW), `frontend/src/lib/Toast.svelte` (NEW), `frontend/src/routes/+layout.svelte`, ~5 page files |
| 2.2 Loading | 4 new components in `frontend/src/lib/`, ~25 page files |
| 2.3 Admin pages | `frontend/src/routes/admin/users/+page.svelte`, `frontend/src/routes/admin/acl/+page.svelte` (NEW) |
| 2.4 Billing pages | 5 pages under `frontend/src/routes/billing/` |
