# Gap Analysis Implementation Plan

> **For Hermes:** Execute in batched phases with parallel subagent dispatch. Revenue batch first, then clinical + reference data in parallel where possible.

**Goal:** Implement the 6 gap-analysis batches (A-F) from `.hermes/plans/2026-08-07_gap-analysis-freemed-0.9.x.md`.

**Architecture:** Follow existing patterns — sqlc queries → Go handlers in `api/` → register in `common.ApiMap`. Patient sub-resources use the checklist at `references/patient-sub-screen-checklist.md`.

---

## Phase A: Revenue (ClaimLog + Ledger) — 2 tasks, parallel

### Task A.1: ClaimLog
- Table `claimlog` exists in schema (table def at schema.sql)
- Create sqlc queries: InsertClaimLog, ListClaimLogByClaim
- Create `api/claimlog.go` with admin-only CRUD
- Register as authenticated endpoint

### Task A.2: Ledger API
- `api/payments.go` has `patientLedger` handler — verify completeness
- Add `GET /api/ledger/:patientId` with date-range filtering
- Return structured ledger entries (charges, payments, adjustments)
- Create sqlc query joining procrec, payrec tables

---

## Phase B: Clinical Patient Sub-Resources — 7 tasks, 3+2+2 dispatch

All follow the patient sub-screen pattern: sqlc (List + Create) → handler → route in `api/patient.go`

### Batch B1 (3 parallel)
- **B.1** CurrentProblems (table: `current_problems` or `patient_problems`)
- **B.2** ChronicProblems (same table, different filter)
- **B.3** EpisodeOfCare (extend `api/eoc.go`, table: `eoc`)

### Batch B2 (2 parallel)
- **B.4** SurgicalHistory (table: `previous_operations` or generic)
- **B.5** Certifications (table: `provider_certifications` or `patient_certifications`)

### Batch B3 (2 parallel)
- **B.6** FinancialDemographics (table: `patient_financial` or `financial_demographics`)
- **B.7** AppointmentFollowup (table: `appointment_followup`)

---

## Phase C: Reference Data Endpoints — picklist infrastructure

### Task C.1: Generic Picklist Endpoint
- Audit 21 reference-data models with picklist registrations (IcdCodes, CptCodes, Pharmacy, etc.)
- Create or verify `GET /api/support/:module/picklist/:query` works
- Add missing picklist registrations where needed

### Task C.2: BillKey Handler
- Model + queries exist, no API handler
- Create `api/billkey.go` with list + create

---

## Phase D: Additional Modules — 3 tasks, 3 parallel

### Task D.1: PhotoID (PhotographicIdentification)
- Table `photoid` exists (migration 000009)
- Create sqlc queries + handler for photo ID CRUD
- Patient sub-resource + standalone admin

### Task D.2: RxList API
- Prescription listing with refill status
- Extend `api/prescriptions.go` or create new endpoint
- JOIN pharmacy table for pharmacy info

### Task D.3: Tickler (Reminders)
- Table `tickler` (check if exists in schema)
- Create sqlc + handler for reminders CRUD

---

## Phase E: Frontend Gaps — 2 tasks

### Task E.1: Patient Sub-Screens
- Add SvelteKit pages for each Batch B clinical module
- Follow existing patterns (medications, vitals, etc.)

### Task E.2: Photo ID UI
- Photo ID capture/upload UI on patient page
- Camera integration or file upload

---

## Execution Order

```
Phase A (Revenue):     A.1 + A.2 parallel
    ↓
Phase B (Clinical):    B1 (3) → B2 (2) → B3 (2)
    ↓
Phase C (Reference):   C.1 + C.2 parallel (or C.1 then C.2)
Phase D (Modules):     D.1 + D.2 + D.3 parallel
    ↓ (can overlap)
Phase E (Frontend):    E.1 + E.2 (after backend endpoints exist)
```

## Risks

- **ClaimLog/ledger tables**: Must verify table schemas in PHP source `../freemed/data/schema/mysql/` before writing sqlc queries
- **Patient sub-resources**: Check if tables exist in schema.sql or need new migrations
- **Generic picklist endpoint**: The support endpoint may already exist — verify first
- **Native billing**: Deferred out of this plan (Agata7 replacement is a separate design effort)
