# Patient Portal Implementation Plan

> **Goal:** Build a HIPAA-compliant patient portal on a separate port with patient-specific auth, appointment requesting, and health information access.

**Architecture:** New SvelteKit SPA served by nginx on port 8080. Shares the Go backend which adds patient-specific JWT auth middleware and read-only API endpoints. Patients authenticate via portal credentials (DOB + patient ID + PIN, or portal password) with rate limiting and security hardening.

---

## Phase 1: Database — Patient Portal Auth

### Task 1.1: Add portal auth columns to patient table
- Migration 000033: `ALTER TABLE patient ADD COLUMN portal_password VARCHAR(255) NOT NULL DEFAULT ''` (bcrypt)
- Migration: `ADD COLUMN portal_pin VARCHAR(255) NOT NULL DEFAULT ''` (bcrypt, 4-6 digit PIN)
- Migration: `ADD COLUMN portal_enabled TINYINT(1) NOT NULL DEFAULT 0`
- Migration: `ADD COLUMN portal_last_login DATETIME`
- Migration: `ADD COLUMN portal_failed_attempts INT NOT NULL DEFAULT 0`
- Append to schema.sql, run sqlc generate

### Task 1.2: Create portal audit log table
- Migration 000034: `CREATE TABLE portal_audit_log`
- Columns: id, created_at, patient_id, action, ip_address, user_agent, success TINYINT(1)
- Purpose: HIPAA access audit trail

---

## Phase 2: Backend — Patient Portal Auth + API

### Task 2.1: Patient portal authentication middleware
- New file: `cmd/freemed-server/portal_auth.go`
- Patient login: POST /portal/auth/login (DOB + patient ID + PIN or password)
- Verify against portal_password or portal_pin on patient record
- Rate limiting (5 attempts per 15 min per IP + per patient)
- Account lockout after 5 failed attempts (portal_enabled = 0)
- JWT with claims: patient_id, role="patient", portal=true
- Audit logging to portal_audit_log
- Logout: blacklist JWT token

### Task 2.2: Patient portal API endpoints
All read-only, patient-scoped to their own patient_id from JWT.

- GET /portal/api/me — patient demographic info
- GET /portal/api/appointments — upcoming + past appointments
- POST /portal/api/appointments/request — request new appointment (within scheduling hours)
- GET /portal/api/medications — current medications
- GET /portal/api/allergies — allergy list
- GET /portal/api/vitals — recent vitals (last 12 months)
- GET /portal/api/problems — current + chronic problems
- GET /portal/api/labs — lab results
- GET /portal/api/documents — scanned/filed documents

Each handler: extract patient_id from JWT, scope all queries to that patient.

### Task 2.3: Scheduling hours validation
- Check config.Config.Scheduler.Start/End for valid hours
- Only allow appointment requests within scheduling hours
- Reject requests outside hours with clear message

---

## Phase 3: Frontend — SvelteKit Patient Portal

### Task 3.1: Scaffold new SvelteKit app
- Create `frontend-portal/` with SvelteKit 5 + Tailwind 4
- Same tech stack as main frontend (adapter-static, zod, Bootstrap)
- Separate npm project, separate build
- Vite proxy: /portal/* → Go backend

### Task 3.2: Portal pages
- `/login` — patient login (DOB + Patient ID + PIN/password)
- `/dashboard` — overview with upcoming appointments, alert badges
- `/appointments` — appointment list + request form
- `/medications` — current meds list
- `/health` — vitals, problems, allergies summary
- `/labs` — lab results
- `/documents` — documents list
- `/profile` — change PIN/password

### Task 3.3: Shared auth store
- `$lib/stores/portal-auth.svelte.ts` — httpOnly JWT cookie, checkAuth(), login(), logout()
- Rate limit feedback in UI (remaining attempts display)
- Session timeout warning (10 min inactivity)

---

## Phase 4: DevOps

### Task 4.1: Docker + nginx
- `frontend-portal/Dockerfile` — nginx serving portal SPA on port 8080
- `frontend-portal/nginx.conf` — proxy /portal/ to backend
- `docker-compose.yml` — add portal service on port 8080

### Task 4.2: Security hardening
- Security headers middleware (same as main app)
- Content-Security-Policy for portal
- SameSite=Strict cookies
- X-Forwarded-For for proper IP logging

---

## Execution Order

```
Phase 1:    1.1 → 1.2 (sequential — migrations)
Phase 2:    2.1 → 2.2 → 2.3 (sequential — auth before APIs)
Phase 3:    3.1 → 3.2 + 3.3 (parallel frontend tasks)
Phase 4:    4.1 + 4.2 (parallel DevOps)
```
