# FreeMED Server TODO Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan phase-by-phase. Parallel dispatch is explicitly called out where applicable.

**Goal:** Complete all remaining pre-production, backend, testing, DevOps, frontend, and cleanup items from TODO.md.

**Architecture:** Go 1.25 + Gin + sqlc + MySQL 8.0 + Redis 7 backend; SvelteKit 5 + Tailwind 4 SPA frontend; Docker Compose deployment with nginx reverse proxy.

**Tech Stack:** Go, Gin, gin-jwt, sqlc, pgx/MySQL driver, golang-migrate, BCrypt, Redis, SvelteKit 5, Tailwind 4, Docker, nginx, MySQL 8.0

---

## Phase 1: Immediate (Pre-Production) — 4 tasks

These items must be verified before any production deployment.

### Task 1.1: Set FREEMED_SESSION_KEY in docker-compose.yml

**Objective:** Add `FREEMED_SESSION_KEY` environment variable to the backend service so production deployments don't use the default `"freemed"` key.

**Files:** Modify: `docker-compose.yml`

**Step 1:** Add to the `backend.environment` block in docker-compose.yml:
```yaml
      FREEMED_SESSION_KEY: ${FREEMED_SESSION_KEY:-changeme-in-production}
```

**Step 2:** Optionally add the variable to `.env.example` or document in README.md.

**Verification:** Run `docker compose config | grep SESSION` and confirm the variable appears.

---

### Task 1.2: Build frontend and test SPA before removing legacy ui/

**Objective:** Run `make frontend-build`, verify the SvelteKit SPA build succeeds, serve it via the Go backend, and confirm it works end-to-end.

**Files:** `frontend/` (build only)

**Step 1:** Run `make frontend-deps && make frontend-build`
**Step 2:** Verify `frontend/build/index.html` exists
**Step 3:** Run `go build ./cmd/freemed-server/` and start with config pointing at a test DB
**Step 4:** Click through key routes (login, dashboard, patients, scheduler, admin, billing)
**Step 5:** Confirm no legacy `ui/` fallback is triggered

**Verification:** `ls frontend/build/index.html` exists, build has zero errors, all SPA routes work.

---

### Task 1.3: Run integration tests against real MySQL database

**Objective:** Write and execute integration tests that use a real MySQL instance (Docker or local) to validate core CRUD operations.

**Files:** Create: `internal/integration/` or `tests/integration/`

**Step 1:** Create a test setup that spins up MySQL via Docker or connects to an existing instance.
**Step 2:** Run migrations against the test database.
**Step 3:** Write Go integration tests for: patient create/list, user auth, scheduler create, encounter create.
**Step 4:** Run with `go test -tags=integration ./...`

**Verification:** All integration tests pass against a real MySQL instance.

---

### Task 1.4: Verify stored procedure schedulerGenerateDailySchedule through sqlc

**Objective:** Confirm the MySQL stored procedure `CALL schedulerGenerateDailySchedule(...)` works correctly via the sqlc-generated query.

**Files:** `internal/db/queries/scheduler.sql` (inspect), maybe tests

**Step 1:** Write a Go test that calls `model.Queries.GenerateDailySchedule(...)` with test data.
**Step 2:** Verify the stored procedure populates the scheduler table as expected.
**Step 3:** Check error handling edge cases (no provider, invalid date, etc.).

**Verification:** `go test` passes, stored procedure produces expected rows.

---

## Phase 2: Database — 3 tasks

### Task 2.1: Write migration to add bcrypt password column

**Objective:** Add a `userpassword_bcrypt VARCHAR(255)` column to the `user` table, allowing phased migration from MD5 to bcrypt without breaking existing logins.

**Files:**
- Create: `internal/db/migrations/000018_bcrypt_password.up.sql`
- Create: `internal/db/migrations/000018_bcrypt_password.down.sql`
- Modify: `internal/db/schema.sql` (append new column to user table)

**Step 1:** Write up migration:
```sql
ALTER TABLE `user` ADD COLUMN `userpassword_bcrypt` VARCHAR(255) NOT NULL DEFAULT '' AFTER `userpassword`;
```

**Step 2:** Write down migration:
```sql
ALTER TABLE `user` DROP COLUMN `userpassword_bcrypt`;
```

**Step 3:** Append the column to the `user` table definition in `internal/db/schema.sql` (use `cat >>` heredoc to avoid patch issues with replace_all).

**Step 4:** Regenerate sqlc (`make sqlc` or `sqlc generate -f internal/db/sqlc.yaml`).

**Step 5:** Update `model/user.go` `CheckUserPassword()` to prefer `userpassword_bcrypt` when non-empty, falling back to `userpassword` (MD5).

**Step 6:** Update `HashPassword()` / `upgradePasswordHash()` to write to the new column.

**Verification:** `make migrate-up` succeeds on fresh DB, login works with bcrypt and MD5 passwords.

---

### Task 2.2: Add database indexes for common query patterns

**Objective:** Analyze the existing queries and schema to identify missing indexes on frequently filtered columns.

**Files:**
- Create: `internal/db/migrations/000019_indexes.up.sql`
- Create: `internal/db/migrations/000019_indexes.down.sql`

**Step 1:** Audit query patterns across all `internal/db/queries/*.sql` files. Key candidates:
- `patient.ptlname, patient.ptfname` (name search)
- `patient.practice` (practice-scoped lists)
- `scheduler.caldateof` (date-range queries)
- `scheduler.calpatient` (patient schedule lookups)
- `messages.msgpatient, messages.msguser` (patient/message filtering)
- `procedures.procpatient` (patient procedure history)
- `user.username` (login lookups — may already be unique)

**Step 2:** Write CREATE INDEX statements for each.

**Step 3:** Write DROP INDEX down migration.

**Verification:** `make migrate-up && make migrate-down` cycles cleanly. `EXPLAIN` on common queries shows index usage.

---

### Task 2.3: Run make migrate-up against fresh database

**Objective:** Verify all 17 existing migrations + new ones apply cleanly to a fresh MySQL 8.0 instance.

**Files:** None (verification only)

**Step 1:** Start fresh MySQL container: `docker compose up -d db`
**Step 2:** Wait for healthy state
**Step 3:** Run `make migrate-up` with correct env vars
**Step 4:** Verify 19 migrations applied (17 existing + 2 new)

**Verification:** `migrate` reports all migrations applied, `show tables` lists all expected tables.

---

## Phase 3: Backend — 5 tasks

These can be dispatched in parallel (independent work).

### Task 3.1: Add health check endpoint (/api/health)

**Objective:** Add a `GET /api/health` endpoint that returns DB and Redis connectivity status.

**Files:**
- Create: `api/health.go`

**Step 1:** Create handler that pings `model.SqlDb` and `common.ActiveSession`.
**Step 2:** Register as unauthenticated route in `common.ApiMap["health"]`.
**Step 3:** Return JSON: `{"status": "ok", "database": true, "redis": true}` or `{"status": "degraded", ...}`.

**Pattern:**
```go
func init() {
    common.ApiMap["health"] = common.ApiMapping{
        Authenticated: false,
        RouterFunction: func(r *gin.RouterGroup) {
            r.GET("/", healthCheck)
        },
    }
}
```

**Verification:** `curl http://localhost:3000/api/health` returns 200 with status JSON.

---

### Task 3.2: Add request ID logging middleware (tracing)

**Objective:** Add middleware that assigns a unique request ID (UUID) to each request and includes it in log output.

**Files:**
- Create: `internal/middleware/request_id.go`
- Modify: `cmd/freemed-server/main.go` (wire into Gin chain)

**Step 1:** Create middleware that generates/reads `X-Request-ID` header and sets it on the Gin context.
**Step 2:** Wire BEFORE gin.Logger() in main.go so the logger picks it up.
**Step 3:** Set `X-Request-ID` on the response header.

**Verification:** `curl -v http://localhost:3000/api/health` shows `X-Request-ID` in response headers.

---

### Task 3.3: Replace manners with stdlib http.Server Shutdown

**Objective:** Remove the archived `github.com/braintree/manners` dependency and use Go's built-in graceful shutdown with `http.Server`.

**Files:**
- Modify: `cmd/freemed-server/main.go` (replace manners listen/serve calls)
- Modify: `cmd/freemed-server/go.mod` (drop manners, run go mod tidy)

**Step 1:** Replace:
```go
log.Fatal(manners.ListenAndServe(...))
```
with:
```go
srv := &http.Server{
    Addr:    fmt.Sprintf(":%d", config.Config.Web.Port),
    Handler: m,
}
go func() {
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal(err)
    }
}()
// Handle SIGINT/SIGTERM for graceful shutdown
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
srv.Shutdown(ctx)
```

**Step 2:** Same for TLS if configured.
**Step 3:** Run `go mod tidy` to remove manners.

**Verification:** `go build ./cmd/freemed-server/` succeeds, server starts and stops cleanly with Ctrl+C.

---

### Task 3.4: Implement HL7 FHIR interchange endpoints

**Objective:** Add FHIR R4 endpoints for exporting patient, encounter, and observation data.

**Files:**
- Create: `api/fhir.go`
- May need: `internal/db/queries/fhir.sql` (sqlc queries for FHIR resource mapping)

**Step 1:** Research FHIR R4 resource types: Patient, Encounter, Observation (vitals), Procedure.
**Step 2:** Create sqlc queries to fetch data in FHIR-mappable format.
**Step 3:** Implement `GET /api/fhir/Patient/:id`, `GET /api/fhir/Observation` (with patient param).
**Step 4:** Return proper FHIR JSON bundles.

**Note:** This is the most complex backend item. Consider a minimal viable subset first: Patient and Observation (vitals) resources.

**Verification:** `curl /api/fhir/Patient/1` returns valid FHIR Patient resource JSON.

---

### Task 3.5: Resolve billing module external dependencies

**Objective:** Fix or remove the `billing/` module's broken local `replace` directives (`remitt-server`, `ratago`).

**Files:**
- Modify: `billing/go.mod` (remove broken replace/require)
- Maybe: `go.work` (re-add billing if fixed)

**Step 1:** Check if billing module is actually imported anywhere: `grep -rn "freemed-server/billing" --include="*.go" .`
**Step 2:** If unused: remove `billing/` from go.work (already excluded), drop replaces, run `go mod tidy` in billing/.
**Step 3:** If needed: vendor or stub the external dependencies, or implement the missing functionality inline.
**Step 4:** Verify `go build ./...` passes from root.

**Verification:** `go build ./...` succeeds; no local replace paths point to missing directories.

---

## Phase 4: Testing — 6 tasks

All tests are independent and can be written in parallel. Use `delegate_task` with 3 tasks per batch.

### Task 4.1: Unit tests for model helpers (CheckUserPassword, HashPassword)

**Objective:** Test password hashing and verification, including the MD5 legacy fallback path.

**Files:** Create: `model/user_test.go`

**Tests:**
- `TestHashPassword` — generates bcrypt, verifies with `bcrypt.CompareHashAndPassword`
- `TestCheckUserPassword_Bcrypt` — stores bcrypt hash, verifies correct password works, wrong fails
- `TestCheckUserPassword_Md5Legacy` — stores MD5 hash, verifies correct password works, verifies auto-upgrade triggers
- `TestCheckUserPassword_UserNotFound` — returns false for nonexistent user
- `TestCheckUserPassword_EmptyPassword` — empty password never matches

**Setup:** Use `sqlmock` to mock sqlc Queries. Skill `go-sqlmock-testing` has patterns.

**Verification:** `go test ./model/ -v -run TestCheckUserPassword` passes.

---

### Task 4.2: Unit tests for RBAC middleware (RequireRole)

**Objective:** Test the `common.RequireRole` middleware with various claim scenarios.

**Files:** Create: `common/rbac_test.go`

**Tests:**
- `TestRequireRole_Allowed` — user with matching role → 200
- `TestRequireRole_Denied` — user with non-matching role → 403
- `TestRequireRole_MultipleRoles` — user matches one of several → 200
- `TestRequireRole_NoClaim` — JWT missing user_type claim → 403
- `TestRequireRole_AdminOnly` — admin route only allows admin

**Setup:** Use Gin test mode with `httptest.NewRecorder()`. Create JWT claims with `jwt.ExtractClaims` by setting up a test middleware that injects claims.

**Verification:** `go test ./common/ -v -run TestRequireRole` passes.

---

### Task 4.3: Integration tests for patient create (transaction rollback)

**Objective:** Verify that patient creation correctly rolls back on error.

**Files:** Create: `tests/integration/patient_test.go` (build tag: `integration`)

**Tests:**
- `TestPatientCreate_Success` — creates patient, verifies DB row
- `TestPatientCreate_DuplicateRollback` — creates patient with duplicate data, verifies rollback
- `TestPatientCreate_RequiredFields` — missing required fields returns error

**Setup:** Needs real MySQL (use Docker compose). Seed test data, run test, clean up.

**Verification:** `go test -tags=integration ./tests/integration/ -v` passes.

---

### Task 4.4: Integration tests for scheduler cancel/reschedule

**Objective:** Verify scheduler appointment lifecycle: create → cancel → reschedule.

**Files:** Create: `tests/integration/scheduler_test.go`

**Tests:**
- `TestSchedulerCreate` — creates appointment, verifies DB
- `TestSchedulerCancel` — cancels appointment, verifies status change
- `TestSchedulerReschedule` — moves appointment, verifies new date/time and old slot freed
- `TestSchedulerDoubleBook` — attempts double-booking, verifies conflict detection
- `TestSchedulerGenerateDailySchedule` — calls stored procedure, verifies slots created

**Verification:** `go test -tags=integration ./tests/integration/ -v -run TestScheduler` passes.

---

### Task 4.5: Test token blacklist flow

**Objective:** Verify JWT token blacklisting works: logout blacklists token, subsequent use is rejected.

**Files:** Create: `tests/integration/auth_test.go` or `cmd/freemed-server/auth_test.go`

**Tests:**
- `TestTokenBlacklist_AfterLogout` — login, logout, try to use same token → 401
- `TestTokenBlacklist_DifferentToken` — login, logout user A, user B's token still works
- `TestTokenBlacklist_Expiry` — blacklisted token after TTL → should be expired anyway

**Verification:** `go test -tags=integration ...` passes.

---

### Task 4.6: Test bcrypt legacy upgrade path

**Objective:** Verify the auto-upgrade from MD5 to bcrypt works correctly.

**Files:** Create: `model/user_test.go` (add tests) or new `tests/upgrade_test.go`

**Tests:**
- `TestBcryptUpgrade_Triggers` — login with MD5 password triggers upgrade goroutine
- `TestBcryptUpgrade_WritesNewHash` — after MD5 login, `userpassword` is replaced with bcrypt
- `TestBcryptUpgrade_SubsequentLogin` — after upgrade, bcrypt path is used (should be transparent)
- `TestBcryptUpgrade_Idempotent` — upgrading an already-bcrypt password is a no-op

**Verification:** Tests pass.

---

## Phase 5: DevOps — 6 tasks

Most DevOps tasks are independent; dispatch in parallel batches.

### Task 5.1: Add Docker health checks to backend service

**Objective:** Add a health check to the Go backend container so Docker can monitor it.

**Files:** Modify: `docker-compose.yml`

**Step 1:** Add to `backend` service:
```yaml
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:3000/api/health"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 15s
```

**Step 2:** Update `depends_on` for `frontend` to use `condition: service_healthy` on backend.

**Verification:** `docker compose up -d && docker compose ps` shows backend as healthy.

---

### Task 5.2: Set up docker-compose override for production

**Objective:** Create `docker-compose.prod.yml` with production overrides (no exposed DB ports, pinned versions, resource limits).

**Files:** Create: `docker-compose.prod.yml`

**Content:**
- Remove `ports` from `db` and `redis` (internal-only)
- Add `restart: always` (already `unless-stopped`, which is fine)
- Add resource limits (memory, CPU)
- Set `FREEMED_SESSION_KEY` to a strong random value
- Add logging config (json-file, rotation)

**Verification:** `docker compose -f docker-compose.yml -f docker-compose.prod.yml config` shows merged config.

---

### Task 5.3: Add database backup script

**Objective:** Create a shell script that dumps the MySQL database and optionally uploads to S3.

**Files:** Create: `scripts/backup-db.sh`

**Step 1:** Script that runs `mysqldump` against the DB container.
**Step 2:** Compress with gzip.
**Step 3:** Timestamp the filename.
**Step 4:** Optionally rotate old backups (keep last 7 days).

**Verification:** Run script, verify `.sql.gz` file is created and can be restored.

---

### Task 5.4: Configure nginx gzip and caching for SPA assets

**Objective:** Optimize the nginx config for the frontend container with gzip compression and better caching.

**Files:** Modify: `frontend/nginx.conf`

**Step 1:** Add gzip directives:
```nginx
gzip on;
gzip_types text/css application/javascript application/json image/svg+xml;
gzip_min_length 256;
gzip_vary on;
```

**Step 2:** Enhance caching: already has `/_app/` with 1y expiry. Add for `.html` with `no-cache`.

**Verification:** Rebuild frontend container, `curl -H "Accept-Encoding: gzip" http://localhost/` shows compressed response.

---

### Task 5.5: Add structured logging (JSON format)

**Objective:** Add JSON-formatted structured logging support, optionally controlled by a config flag.

**Files:**
- Modify: `config/config.go` (add `LogFormat` field)
- Modify: `cmd/freemed-server/main.go` (switch log format based on config)

**Step 1:** Add `LogFormat string` to AppConfig (default: "text").
**Step 2:** When set to "json", configure `log.SetFlags(0)` and use `log/slog` for structured output.
**Step 3:** Or use a Gin middleware that outputs JSON access logs.

**Verification:** Start with `FREEMED_LOG_FORMAT=json`, verify log output is valid JSON lines.

---

### Task 5.6: Set up monitoring (Prometheus metrics)

**Objective:** Add a `/metrics` endpoint with Prometheus-formatted metrics (request count, latency, DB pool stats).

**Files:**
- Create: `api/metrics.go` or `internal/middleware/metrics.go`
- Possibly add: `prometheus/client_golang` dependency

**Step 1:** Add prometheus Go client library.
**Step 2:** Create middleware that tracks request count, duration, status code.
**Step 3:** Register `GET /api/metrics` endpoint.
**Step 4:** Add DB pool stats (open connections, in-use, idle).

**Verification:** `curl http://localhost:3000/api/metrics` returns Prometheus-formatted metrics.

---

## Phase 6: Frontend (SvelteKit) — 4 tasks

### Task 6.1: Responsive audit — test all pages at mobile widths

**Objective:** Test every frontend page at 375px (iPhone) and 768px (iPad) widths; fix layout issues.

**Files:** Potentially many frontend route files

**Step 1:** For each route, capture screenshots at mobile widths.
**Step 2:** Identify issues: horizontal scroll, overlapping elements, truncated text, unreadable tables.
**Step 3:** Fix with Tailwind responsive classes (`md:`, `lg:`, etc.).
**Step 4:** Tables: add horizontal scroll wrapper (`overflow-x-auto`).
**Step 5:** Navigation: ensure mobile hamburger menu works.

**Verification:** All pages are usable and readable at 375px width.

---

### Task 6.2: Add E2E tests (Playwright)

**Objective:** Set up Playwright for end-to-end testing of critical user flows.

**Files:**
- Create: `frontend/e2e/` directory
- Create: `frontend/playwright.config.ts`
- Create: tests for login, patient search, scheduler, admin

**Step 1:** Install Playwright: `cd frontend && npm install -D @playwright/test && npx playwright install`
**Step 2:** Configure playwright.config.ts for SPA (baseURL pointing to dev server).
**Step 3:** Write tests:
- Login flow (valid credentials → dashboard, invalid → error)
- Patient list/search
- Schedule appointment (create → verify in list)
- Admin users (list, create, delete)

**Verification:** `npx playwright test` passes against running dev server.

---

### Task 6.3: Schedule recurring appointments UI

**Objective:** Add UI for configuring recurring appointment templates in the scheduler.

**Files:**
- Create: `frontend/src/routes/scheduler/recurring/+page.svelte`
- May need: backend API endpoints for appttemplate CRUD
- Modify: `frontend/src/routes/scheduler/+page.svelte` (link to recurring)

**Step 1:** If no backend endpoint exists, create `api/appttemplate.go` with list/create/update/delete.
**Step 2:** Build Svelte page with form: name, duration, equipment selection, color picker.
**Step 3:** Add to scheduler page sidebar/nav.

**Verification:** Can create, edit, and delete recurring appointment templates; they appear in scheduler.

---

### Task 6.4: Remove legacy `ui/` directory once SPA verified

**Objective:** After Phase 1 Task 1.2 confirms SPA works, remove the legacy `ui/` directory.

**Files:** Delete: `ui/` directory (entire tree)
- Modify: `cmd/freemed-server/main.go` (remove legacy fallback)
- Modify: `.gitignore` (if ui/ is ignored)

**Step 1:** Verify SPA is serving correctly (Task 1.2 must be complete).
**Step 2:** Remove the `else` block in main.go that serves legacy `ui/`.
**Step 3:** Remove `ui/` directory: `git rm -r ui/`
**Step 4:** Simplify frontend-serving logic to just the SPA path.

**Verification:** `go build` passes, app serves only SPA.

---

## Phase 7: Cleanup — 4 tasks

### Task 7.1: Remove common.Md5hash if no non-session uses remain

**Objective:** Replace the single remaining use of `common.Md5hash` in `session.go` with a crypto/rand UUID, then remove the function.

**Files:**
- Modify: `common/session.go` (replace Md5hash call)
- Modify: `common/util.go` (remove Md5hash function)

**Step 1:** In `session.go:56`, replace:
```go
sid := fmt.Sprintf("%d-%s", time.Now().Unix(), Md5hash(...))
```
with:
```go
sid := uuid.New().String()
```
(already importing `uuid` in `cmd/freemed-server/auth.go` — here we'd use `github.com/google/uuid` in common)

**Step 2:** Remove `Md5hash` from `common/util.go` and the `crypto/md5` import if unused.

**Step 3:** Run `go build ./...` to verify no other callers.

**Verification:** `grep -rn "Md5hash" --include="*.go" .` returns zero results.

---

### Task 7.2: Remove unused model types after sqlc migration verified

**Objective:** Identify and remove model/*.go files that have been fully replaced by sqlc-generated code.

**Files:** Audit: `model/` directory

**Step 1:** Compare model/*.go files against dbgen/*.sql.go — any struct that exists in dbgen with full coverage can be removed from model.
**Step 2:** Check for references to each model struct: `grep -rn "model.TypeName" --include="*.go" .`
**Step 3:** Remove model files with zero references.
**Step 4:** Run `go build ./...` to verify.

**Note:** Be conservative — some model files provide picklist registrations or business logic that dbgen doesn't replace.

**Verification:** `go build ./...` passes after removals.

---

### Task 7.3: Standardize error response format across all handlers

**Objective:** Ensure all API handlers return a consistent error JSON structure: `{"error": "message"}` or `{"code": NNN, "message": "..."}`.

**Files:** Audit: all `api/*.go` files

**Step 1:** Create a helper in `common/`:
```go
func ErrorResponse(code int, message string) gin.H {
    return gin.H{"code": code, "message": message}
}
```

**Step 2:** Audit all `c.AbortWithError`, `c.AbortWithStatusJSON`, `c.JSON` error paths.
**Step 3:** Replace inconsistent patterns with the helper.
**Step 4:** Run `go build ./...` and verify tests still pass.

**Verification:** All error responses follow the same JSON shape.

---

### Task 7.4: Add pagination to list endpoints

**Objective:** Add offset/limit pagination to list endpoints that currently return all rows.

**Files:** Modify: `api/patient.go`, `api/messages.go`, `api/scheduler.go`, and their sqlc queries.

**Step 1:** Add `offset` and `limit` query parameters with defaults (offset=0, limit=50).
**Step 2:** Update sqlc queries to accept `LIMIT :limit OFFSET :offset` parameters.
**Step 3:** Return pagination metadata: `{"data": [...], "total": N, "offset": 0, "limit": 50}`.
**Step 4:** Require a COUNT query for total.

**Pattern:**
```go
offset := common.ParseInt(c.DefaultQuery("offset", "0"))
limit := common.ParseInt(c.DefaultQuery("limit", "50"))
```

**Verification:** `curl "/api/patient?offset=10&limit=5"` returns 5 results with correct total count.

---

## Execution Order

```
Phase 1 (Immediate): All 4 tasks can run in parallel
    ↓
Phase 2 (Database): 2.1 → 2.2 → 2.3 (sequential — migrations build on each other)
    ↓
Phase 3 (Backend): All 5 tasks parallel (dispatch 3 + 2 in two batches)
    ↓
Phase 4 (Testing): All 6 tasks parallel (dispatch 3 + 3 in two batches)
Phase 5 (DevOps): 5.1, 5.3, 5.6 parallel → 5.2, 5.4, 5.5 parallel
    ↓ (can overlap with 4+5)
Phase 6 (Frontend): 6.1, 6.3 parallel (6.2 after 6.3, 6.4 last)
    ↓
Phase 7 (Cleanup): 7.1, 7.2 parallel → 7.3 → 7.4 (sequential)
```

**Estimated total: 33 tasks across 7 phases.**

## Risks and Constraints

- **No existing test infrastructure**: The project has only `package_test.go`. Need to set up test fixtures (sqlmock, httptest, test DB) from scratch.
- **Billing module broken deps**: `remitt-server` and `ratago` are local `replace` paths that don't resolve. May need to stub or vendor.
- **sqlc generate may fail**: The `dbgen/documents.sql.go` has pre-existing type errors (`UnfiledDoc`, `UnreadDoc` undefined). Move it aside during generation.
- **FHIR is complex**: Task 3.4 is the highest-risk item. A full FHIR implementation is weeks of work; consider a minimal proof-of-concept.
- **Playwright setup**: Requires Node.js and browser binaries. Make sure CI can handle this.
- **Stored procedure testing**: `schedulerGenerateDailySchedule` is a MySQL stored procedure that sqlc wraps via `Exec`. Testing requires a real MySQL instance with the procedure loaded.
