# Feature Parity Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan phase-by-phase.

**Goal:** Achieve feature parity with FreeMED 0.9.x PHP codebase, prioritized by clinical
necessity and revenue impact.

**Architecture:** Add Go API handlers + sqlc queries + SvelteKit pages for each module,
following existing patterns in `api/`, `internal/db/queries/`, `frontend/src/routes/`.

---

## Phase 0: Reference Data Picklists (Quick Wins — 22 endpoints)

All 22 tables already exist in `schema.sql` with Go model structs. Each needs:
1. A `DbSupportPicklist` entry in its model file (or a new sqlc picklist query)
2. Verify the picklist works via `/api/support/:module/picklist/:query`

**Modules:** BodySite, CptCodes, CptModifiers, DocumentCategory, DrugForms,
DrugQuantityQualifiers, EnclosureTypes, IcdCodes, InternalServiceTypes, Loinc,
PlaceOfService, RouteOfAdministration, CoverageTypes, ClaimTypes, InsuranceCompany,
InsuranceCompanyGroup, InsuranceModifiers, Facility, Practice, Provider,
WorkflowStatus, WorkflowStatusType

**Estimated:** 1 hour (bulk add picklist entries)

---

## Phase 1: Core Clinical (Tier 1 — Patient Safety)

### Task 1.1: Allergies
- Create `api/allergies.go` — `GET /api/patient/:id/allergies`, `POST /api/patient/:id/allergies`
- Create `internal/db/queries/allergies.sql` — list, create, update, delete
- Frontend: `src/routes/patients/[id]/allergies/+page.svelte`
- Link from patient view page

### Task 1.2: Medications
- Create `api/medications.go` — `GET /api/patient/:id/medications`, `POST`
- Create `internal/db/queries/medications.sql`
- Frontend: `src/routes/patients/[id]/medications/+page.svelte`

### Task 1.3: Vitals
- Create `model/vitals.go` + migration for vitals table (if schema missing)
- Create `api/vitals.go` — `GET /api/patient/:id/vitals`, `POST`
- Frontend: vitals display on patient view or dedicated page

### Task 1.4: Encounter Notes
- Backend: `api/encounters.go` — `GET /api/patient/:id/encounters`, `GET /api/patient/:id/encounters/:encounterId`
- sqlc queries for encounters table
- Frontend: replace placeholder `encounters/+page.svelte` with real data

### Task 1.5: Diagnosis Listing (DxForPatient)
- Add `GET /api/patient/:id/diagnoses` returning ICD codes linked to patient
- Frontend: diagnosis list on patient view

---

## Phase 2: Billing Foundation (Tier 0 — Revenue Critical)

### Task 2.1: Procedure Listing
- `api/procedures.go` — `GET /api/patient/:id/procedures`, `GET /api/procedures/:id`
- sqlc: procedure list with joins to CPT codes

### Task 2.2: Payments
- `api/payments.go` — `GET /api/patient/:id/payments`, `POST`
- `GET /api/patient/:id/ledger` — combined procedure + payment view

### Task 2.3: Authorizations
- `api/authorizations.go` — `GET /api/patient/:id/authorizations`
- sqlc: authorization list with insurance company joins

### Task 2.4: Patient Coverages
- `api/coverages.go` — `GET /api/patient/:id/coverages`, `POST`, `DELETE`
- Frontend: insurance tab on patient view

---

## Phase 3: Scheduler Completeness

### Task 3.1: Set/Create Appointment
- `POST /api/scheduler` — create new appointment
- Frontend: "New Appointment" button on scheduler page

### Task 3.2: Move/Copy Appointment
- `POST /api/scheduler/:id/move` — move to different date/time
- `POST /api/scheduler/:id/copy` — duplicate to new slot

### Task 3.3: Group Appointments
- `GET /api/scheduler/group/:groupId` — group detail
- `POST /api/scheduler/group` — create group appointment

---

## Phase 4: UX Infrastructure

### Task 4.1: Dashboard Aggregation (getDashBoardDetails)
- `GET /api/dashboard` — aggregated counts: patients, appointments today, unread messages, action items

### Task 4.2: Global Search (ModuleSearch)
- `GET /api/search?q=` — search across patients, appointments, messages

### Task 4.3: User Preferences Write
- `PUT /api/config/:id` — update config values (SystemConfig.SetValues)

---

## Phase 5: Remaining Modules

Lower priority — implement as needed:
- Printing/PDF, Signatures, Tickler, TableMaintenance, Letters, Reports
- DICOM, ScannedDocuments, UnfiledDocuments, UnreadDocuments
- Pharmacy: Prescriptions, RxRefill, NDC lookup
- WorkLists, Rules, EpisodeOfCare
