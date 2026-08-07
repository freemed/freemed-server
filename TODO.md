# TODO

## Immediate (pre-production)

- [ ] Run integration tests against real MySQL database
- [ ] Verify stored procedure `CALL schedulerGenerateDailySchedule` works through sqlc
- [ ] Set `FREEMED_SESSION_KEY` in docker-compose.yml
- [ ] Build frontend (`make frontend-build`) and test SPA before removing legacy ui/

## Auth hardening

- [ ] Move token storage from localStorage to httpOnly cookie (XSS hardening)
- [ ] Add CSRF token requirement for POST /auth/login
- [ ] Add rate limiting on /auth/login (brute force prevention)
- [ ] Wire up `RequireRole()` on admin endpoints (users, acl, config write)

## Database

- [ ] Write migration to add bcrypt password column; phase out MD5 userpassword column
- [ ] Add database indexes for common query patterns
- [ ] Run `make migrate-up` against a fresh database to verify 17 migrations

## Frontend (SvelteKit)

- [ ] Wire frontend pages to new backend endpoints (many pages are placeholder/backend-only)
- [ ] Add offline/loading/error states to all pages
- [ ] Replace `window.toast` global with proper Svelte store-based toast
- [ ] Add form validation library (e.g. sveltekit-superforms)
- [ ] Responsive audit: test all pages at mobile widths
- [ ] Add E2E tests (Playwright or Cypress)
- [ ] Remove legacy `ui/` directory once SPA deployment is verified
- [ ] Add billing pages (claims manager, AR aging, remitt, superbills frontend)
- [ ] Add admin pages (ACL management, user management)
- [ ] Add call-in page
- [ ] Schedule recurring appointments UI

## Backend

- [ ] Implement HL7 FHIR interchange endpoints
- [ ] Add request ID logging middleware (tracing)
- [ ] Add health check endpoint (/api/health)
- [ ] Replace manners with stdlib http.Server Shutdown
- [ ] Resolve billing module external dependencies (remitt-server, ratago)

## DevOps

- [ ] Add Docker health checks
- [ ] Set up docker-compose override for production
- [ ] Add database backup script
- [ ] Configure nginx gzip and caching for SPA assets
- [ ] Add structured logging (JSON format)
- [ ] Set up monitoring (Prometheus metrics)

## Testing

- [ ] Unit tests for model helpers (CheckUserPassword, HashPassword)
- [ ] Unit tests for RBAC middleware (RequireRole)
- [ ] Integration tests for patient create (transaction rollback)
- [ ] Integration tests for scheduler cancel/reschedule
- [ ] Test token blacklist flow
- [ ] Test bcrypt legacy upgrade path

## Cleanup

- [ ] Remove `common.Md5hash` if no non-session uses remain
- [ ] Remove unused model types after sqlc migration verified
- [ ] Standardize error response format across all handlers
- [ ] Add pagination to list endpoints (patients, messages, scheduler)
