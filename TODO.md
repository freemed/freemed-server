# TODO

## Immediate (pre-production)

- [x] Run integration tests against real MySQL database
- [x] Verify stored procedure `CALL schedulerGenerateDailySchedule` works through sqlc
- [x] Set `FREEMED_SESSION_KEY` in docker-compose.yml
- [x] Build frontend (`make frontend-build`) and test SPA before removing legacy ui/

## Auth hardening

- [x] Move token storage from localStorage to httpOnly cookie (XSS hardening)
- [x] Add CSRF token requirement for POST /auth/login
- [x] Add rate limiting on /auth/login (brute force prevention)
- [x] Wire up `RequireRole()` on admin endpoints (users, acl, config write)
- [x] Add security headers middleware (CSP, X-Frame-Options, etc.)
- [x] Add production config warnings (ValidateProduction)

## Database

- [x] Write migration to add bcrypt password column; phase out MD5 userpassword column
- [x] Add database indexes for common query patterns
- [x] Run `make migrate-up` against a fresh database to verify 28 migrations
- [x] Fix duplicate migration numbers (000004, 000012, 000013 → 000021-000023)
- [x] Fix scheduler proc migration (DELIMITER → multiStatements=true)
- [x] Add clinical tables: current_problems (000024), chronic_problems (000025), previous_operations (000026), certifications (000027), financial_demographics (000028)

## Frontend (SvelteKit)

- [x] Wire frontend pages to new backend endpoints (admin users, ACL, billing, call-in)
- [x] Add offline/loading/error states to all pages (reusable components created)
- [x] Replace `window.toast` global with proper Svelte store-based toast
- [x] Add form validation library (zod on login page)
- [x] Responsive audit: wrap all tables in overflow-x-auto (8 files fixed)
- [x] Add E2E tests (Playwright — 11/12 passing)
- [x] Remove legacy `ui/` directory once SPA deployment is verified
- [x] Add billing pages (claims manager, AR aging, remitt, superbills frontend)
- [x] Add admin pages (ACL management, user management)
- [x] Add call-in page
- [x] Schedule recurring appointments UI (templates CRUD + frontend)
- [x] Patient sub-screens for new clinical modules (problems, surgical-history, certifications, financial)

## Backend

- [x] Implement HL7 FHIR interchange endpoints (Patient, Observation, CapabilityStatement)
- [x] FHIR R4 audit: add profile, narrative, address, deceased, generalPractitioner, UCUM fix, OperationOutcome
- [x] Add request ID logging middleware (tracing)
- [x] Add health check endpoint (/api/health)
- [x] Replace manners with stdlib http.Server Shutdown
- [x] Resolve billing module external dependencies (billing not imported — no-op)
- [x] ClaimLog audit trail (queries + handler)
- [x] Patient financial ledger (date range filtering, pagination, standalone /api/ledger)
- [x] CurrentProblems API → `/api/patient/:id/current-problems`
- [x] ChronicProblems API → `/api/patient/:id/chronic-problems`
- [x] PreviousOperations (surgical history) → `/api/patient/:id/surgical-history`
- [x] Certifications → `/api/patient/:id/certifications`
- [x] FinancialDemographics → `/api/patient/:id/financial`
- [x] BillKey handler (admin-only)
- [x] RxList API → `/api/rxlist/:patientId` (pharmacy JOIN + refill info)
- [x] Reference data picklists (30 modules registered; routeofadmin dupe fixed)
- [x] EpisodeOfCare → extend `/api/patient/:id/eoc` (existing eoc.go partial)

## DevOps

- [x] Add Docker health checks
- [x] Set up docker-compose override for production
- [x] Add database backup script
- [x] Configure nginx gzip and caching for SPA assets
- [x] Add structured logging (JSON format via FREEMED_LOG_FORMAT=json)
- [x] Set up monitoring (Prometheus metrics at /api/metrics)

## Testing

- [x] Unit tests for model helpers (CheckUserPassword, HashPassword) — 9 tests
- [x] Unit tests for RBAC middleware (RequireRole) — 4 tests
- [x] Integration tests for patient create (transaction rollback) — 2 tests
- [x] Integration tests for scheduler cancel/reschedule — 4 tests
- [x] Test token blacklist flow
- [x] Test bcrypt legacy upgrade path — 3 tests

## Cleanup

- [x] Remove `common.Md5hash` (replaced with uuid.New in session ID generation)
- [x] Remove unused model types (14 files removed)
- [x] Standardize error response format across all handlers (31 files, 230+ patterns)
- [x] Add pagination to list endpoints (messages, scheduler)

---

## Remaining Gaps (prioritized)

### HIGH: Clinical
- [x] EpisodeOfCare → extend existing `api/eoc.go` with list/create handlers
- [x] UserGroups handler (user→group assignments; ACL.go covers groups)
- [x] UserPreferences endpoint (model/user.go exists)

### MEDIUM: Modules
- [ ] Forms module (clinical form templates)
- [ ] Signatures API (digital signature capture)
- [ ] DicomModule (radiology — significant effort)
- [ ] Tickler API (reminders — needs background job infrastructure)

### LOW: Frontend
- [ ] Patient sub-screens for clinical modules (SvelteKit pages)
- [ ] Native Go billing claims creation (Agata7 replacement — design task)
- [ ] SchedulingRules module

### OBSOLETE
- [x] CDRWBackup, FaxStatus, UpdatesModule
- [x] AppointmentFollowup (no dedicated table; scheduler flag)
