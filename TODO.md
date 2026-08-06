# TODO

## Immediate (pre-production)

- [ ] Run integration tests against real MySQL database
- [ ] Verify stored procedure `CALL schedulerGenerateDailySchedule` works through sqlc
- [ ] Test picklist queries with `SqlDb.QueryContext()` pattern (support.go)
- [ ] Set `FREEMED_SESSION_KEY` in docker-compose.yml (currently uses config.yml default)
- [ ] Build frontend (`make frontend-build`) and test SPA on port 80 before removing legacy ui/

## Auth hardening

- [ ] Move token storage from localStorage to httpOnly cookie (XSS hardening)
- [ ] Add CSRF token requirement for POST /auth/login
- [ ] Implement token rotation tracking (reject tokens older than latest iat)
- [ ] Populate user_type claim from DB; add per-route role requirements using RequireRole()
- [ ] Add rate limiting on /auth/login (brute force prevention)
- [ ] Add account lockout after N failed attempts

## Database

- [ ] Write migration to add bcrypt password column; phase out MD5 userpassword column
- [ ] Add database indexes for common query patterns (patient.ptarchive, scheduler.caldateof, messages.msgfor/msgread)
- [ ] Run `make migrate-up` against a fresh database to verify migration works

## Frontend (SvelteKit)

- [ ] Replace localStorage token with httpOnly cookie (requires backend Cookie auth change)
- [ ] Add offline/loading/error states to all pages (some pages only have basic error handling)
- [ ] Replace `window.toast` global with proper Svelte store-based toast
- [ ] Add form validation library (e.g. sveltekit-superforms)
- [ ] Responsive audit: test all pages at mobile widths (375px)
- [ ] Add E2E tests (Playwright or Cypress)
- [ ] Remove legacy `ui/` directory once SPA deployment is verified working
- [ ] Add favicon and app manifest for PWA support

## Backend

- [ ] Implement remaining EMR module handlers (encounters, documents, billing, immunizations)
- [ ] Add pagination to list endpoints (patients, messages, scheduler)
- [ ] Add search/filter parameters to messages list endpoint
- [ ] Wire up PUT /api/emr/data_store/put for data store writes
- [ ] Implement HL7 FHIR interchange endpoints
- [ ] Add request ID logging middleware (tracing)
- [ ] Add health check endpoint (/api/health)
- [ ] Replace manners (graceful shutdown) with stdlib http.Server Shutdown

## DevOps

- [ ] Add Docker health checks for backend and frontend containers
- [ ] Set up docker-compose override for production (secrets, volumes, resource limits)
- [ ] Add database backup script / cron job
- [ ] Configure nginx gzip and caching properly for SPA assets
- [ ] Add structured logging (JSON format for log aggregation)
- [ ] Set up monitoring (Prometheus metrics endpoint)

## Testing

- [ ] Unit tests for model helpers (CheckUserPassword, HashPassword, upgradePasswordHash)
- [ ] Unit tests for RBAC middleware (RequireRole)
- [ ] Integration tests for patient create endpoint (transaction rollback scenarios)
- [ ] Integration tests for scheduler cancel/reschedule
- [ ] Test token blacklist flow (login → use token → logout → reuse token → rejected)
- [ ] Test bcrypt legacy upgrade path (MD5 hash → login → auto-upgraded to bcrypt)
- [ ] Load test scheduler daily range query with realistic data volume

## Cleanup

- [ ] Remove `common.Md5hash` if no non-session uses remain
- [ ] Remove unused model types after sqlc migration verified (UserModel, PatientModel, etc.)
- [ ] Remove `model/DbModuleWithHooks`, `model/DbTable` types if unused
- [ ] Remove `billing/remitt.go` commented-out handler or implement it
- [ ] Remove deprecated `//go:embed` blank imports if any remain
- [ ] Standardize error response format across all handlers
