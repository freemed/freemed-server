# Feature Gap Reduction — Round 2

> **For Hermes:** Use subagent-driven-development. Dispatch each phase in parallel where possible.

**Goal:** Implement the next 12-15 highest-impact modules from the PHP 0.9.x feature parity
analysis, focusing on billing completeness, patient management, and UX infrastructure.

---

## Phase A: Billing Completeness (4 modules)

### Task A.1: Authorizations
- `GET /api/patient/:id/authorizations` — list with insurance company name
- `GET /api/authorizations/:id` — single authorization detail
- sqlc: `ListAuthorizations :many`, `GetAuthorization :one`
- Frontend: authorizations tab on patient view

### Task A.2: Patient Coverages
- `GET /api/patient/:id/coverages` — list coverages with insurance details
- `POST /api/patient/:id/coverages` — add coverage
- `DELETE /api/patient/:id/coverages/:coverageId` — remove coverage
- sqlc: `ListCoverages :many`, `CreateCoverage :execresult`, `RemoveCoverage :exec`
- Frontend: coverages tab/section on patient view

### Task A.3: Claim Log Listing
- `GET /api/patient/:id/claims` — list claim history with status
- sqlc: `ListClaims :many`

### Task A.4: Billing Action Items (ClaimLog::aging, Authorizations::getActionItems)
- `GET /api/billing/action-items` — aggregate action items across modules
- Combines: unpaid claims, expiring authorizations, unbilled procedures

---

## Phase B: Patient Management (3 modules)

### Task B.1: Patient Tags
- `GET /api/patient/:id/tags` — list tags
- `POST /api/patient/:id/tags` — create tag
- `DELETE /api/patient/:id/tags/:tagId` — expire tag
- `GET /api/patients/tags/search?q=` — search patients by tag
- sqlc: `ListTags :many`, `CreateTag :execresult`, `ExpireTag :exec`, `SearchByTag :many`
- Frontend: tags display on patient view with search

### Task B.2: Patient Address Management
- `PUT /api/patient/:id/addresses` — update addresses (already have create in POST /patients)
- `DELETE /api/patient/:id/addresses/:addressId` — remove address
- sqlc: `UpdateAddress :exec`, `DeleteAddress :exec`

### Task B.3: Patient Track History
- `GET /api/patient/:id/history` — combined timeline of encounters, procedures, notes
- sqlc: UNION ALL across patient_emr, procrec, pnotes

---

## Phase C: UX Infrastructure (3 modules)

### Task C.1: Dashboard Aggregation
- `GET /api/dashboard` — returns:
  - Patient count, appointments today, unread messages, pending action items
  - sqlc: individual COUNT queries aggregated in handler
- Frontend: update `+page.svelte` dashboard with real aggregated data

### Task C.2: Global Search
- `GET /api/search?q=` — search across patients, messages, scheduler
- sqlc: `SearchPatients :many`, `SearchMessages :many`
- Frontend: search bar in navbar

### Task C.3: Config Write Support + EMR Configuration
- `PUT /api/config/:id` — update config value (SystemConfig.SetValues)
- `GET /api/config/sections` — list config sections
- `GET /api/emr-configuration` — return EMR module configuration

---

## Phase D: Messaging & Notifications (3 modules)

### Task D.1: Message Tags + Bulk Operations
- `GET /api/messages/tags` — list unique tags
- `GET /api/messages/tag/:tag` — filter messages by tag
- `DELETE /api/messages` — bulk delete messages
- sqlc: `ListMessageTags :many`, `MessagesByTag :many`, `DeleteMessages :exec`

### Task D.2: System Notifications
- `GET /api/notifications` — list for current user
- `GET /api/notifications/unread-count` — unread count
- `GET /api/notifications/patient/:id` — notifications for a patient
- sqlc: `ListNotifications :many`, `NotificationCount :one`

### Task D.3: User List (GetUsers)
- `GET /api/users` — list all users (admin only via RequireRole)
- sqlc: `ListUsers :many`

---

## Phase E: Pharmacy Foundation (2 modules)

### Task E.1: Prescriptions
- `GET /api/patient/:id/prescriptions` — list prescriptions
- `POST /api/patient/:id/prescriptions` — create prescription
- sqlc: `ListPrescriptions :many`, `CreatePrescription :execresult`
- Frontend: prescriptions tab on patient view

### Task E.2: Pharmacy Picklist
- `GET /api/support/pharmacy/picklist/:query` — already registered via DbSupportPicklists? Verify.
- If not, add picklist entry

---

## Implementation Order

```
Batch 1 (parallel):
  A.1 Authorizations
  A.2 Patient Coverages
  B.1 Patient Tags

Batch 2 (parallel):
  A.3 Claim Log
  B.2 Patient Addresses
  C.1 Dashboard Aggregation

Batch 3 (parallel):
  C.2 Global Search
  C.3 Config Write + EMR Config
  D.1 Message Tags + Bulk

Batch 4 (parallel):
  D.2 System Notifications
  D.3 User List
  E.1 Prescriptions
  E.2 Pharmacy Picklist
  B.3 Track History

Batch 5 (sequential — depends on all above):
  A.4 Billing Action Items (aggregates data from A.1-A.3)
```

**Estimated:** 4 parallel batches × ~3 min each + 1 sequential batch = ~15 minutes total
