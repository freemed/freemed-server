# FreeMED 0.9.x → FreeMED Server Gap Analysis

> Generated: 2026-08-07 | Source: `../freemed` (PHP 0.9.x) vs `freemed-server` (Go)

## Overview

| Metric | PHP 0.9.x | Go Server | Coverage |
|--------|-----------|-----------|----------|
| Module classes | 133 | ~68 equivalent handlers/queries | ~51% |
| Full parity modules | — | 68 | 51% |
| Partial parity | — | 5 | 4% |
| Missing modules | — | 60 | 45% |
| API endpoint classes | 24 | 14 | 58% |
| Missing APIs | — | 10 | 42% |

---

## Tier 0: MISSING PHP APIs (Backend Gap)

These are dedicated `lib/org/freemedsoftware/api/*.class.php` files that have no Go equivalent. They represent entire API subsystems.

| PHP API | Purpose | Priority | Notes |
|---------|---------|----------|-------|
| **Agata7** | Billing clearinghouse transport (Agata format) | DEPRECATE | Replace with native Go billing engine; no PHP transport dependency |
| **Transport** | Generic billing transport/EDI layer | DEPRECATE | Companion to Agata7; native Go billing replaces both |
| **ClaimLog** | Claims audit log / tracking | HIGH | Revenue integrity; needed regardless of billing backend |
| **Ledger** | Detailed patient financial ledger | HIGH | Billing; partially in payments module, needs full audit trail |
| **RxList** | Prescription listing, refill management | MEDIUM | Medication management |
| **Tickler** | Reminder/tickler/to-do system | MEDIUM | Clinical workflow |
| **Signatures** | Digital signature capture/verification | MEDIUM | Clinical documentation |
| **Fax** | Fax integration (send/receive) | LOW | Legacy comms; EHR certification |
| **Printing** | Server-side print/PDF generation | LOW | Reports already exist |
| **GraphInterface** | Charting/graphing data API | LOW | Frontend charts can use REST data |
| **FormTemplate** | Dynamic form template engine | LOW | Can be replaced by Svelte components |
| **FormTemplateList** | Form template catalog | LOW | Same as above |
| **ModuleInterface** | Generic CRUD module API | LOW | sqlc replaces this pattern |
| **TableMaintenance** | DB admin/table maintenance | LOW | golang-migrate handles schema |

---

## Tier 1: HIGH PRIORITY Missing Modules

Clinical and revenue modules with patient data or significant business logic (>2 methods).

| PHP Module | Methods | Category | Priority |
|------------|---------|----------|----------|
| **DicomModule** | 6 (patient-linked) | Radiology / DICOM imaging | CRITICAL |
| **PhotographicIdentification** | 7 (patient-linked) | Photo ID capture/storage | HIGH |
| **Forms** | 2 (patient-linked, CRUD) | Clinical form templates | HIGH |
| **FinancialDemographics** | 1 (patient-linked) | Insurance/financial data per patient | HIGH |
| **EpisodeOfCare** | 4 (patient-linked) | Episode-of-care tracking | HIGH |
| **CurrentProblems** | 1 (patient-linked) | Active problem list | HIGH |
| **ChronicProblems** | 1 (patient-linked) | Chronic condition tracking | HIGH |
| **Certifications** | 2 (patient-linked) | Provider certifications for patients | MEDIUM |
| **PreviousOperationsModule** | 1 (patient-linked) | Surgical history | MEDIUM |
| **AppointmentFollowup** | 1 (patient-linked) | Post-appointment follow-up | MEDIUM |

### PARTIAL Modules (Needs Completion)

| Module | What's Missing | Priority |
|--------|---------------|----------|
| **ProgressNotes** | Handler exists (patient_progress_notes.go) but no dedicated queries file; queries are in patient.sql | LOW — queries exist in patient.sql |
| **UserGroups** | Handler missing; ACL.go covers groups but UserGroups.class.php maps users→groups | MEDIUM |
| **UserPreferences** | No dedicated handler or queries; user.go has model only | MEDIUM |
| **BillKey** | Model + queries exist, no API handler | LOW |
| **DiagnosisFamily** | Queries exist (diagnoses.sql), no handler | LOW |

---

## Tier 2: REFERENCE DATA / PICKLIST Modules

These are thin table wrappers (0-2 methods, no patient data). Most serve as picklist/reference data. They have Go models but need API handlers for CRUD or at minimum read-only list endpoints.

### Has Go Model, Needs Handler (12 modules)

| Module | Go Model |
|--------|----------|
| Bccdc | `model/bccdc.go` |
| BillingClearinghouse | `model/clearinghouse.go` |
| BillingContact | `model/billingcontact.go` |
| BillingService | `model/billingservice.go` |
| BodySite | `model/bodysite.go` |
| ClaimTypes | `model/claimtype.go` |
| CptCodes | `model/cpt.go` |
| CptModifiers | `model/cptmodifier.go` |
| DrugForms | `model/drugforms.go` |
| DrugQuantityQualifiers | `model/drugquantityqualifier.go` |
| EnclosureTypes | `model/enclosuretype.go` |
| IcdCodes | `model/icdcode.go` |
| InsuranceCompanyModule | `model/insco.go` |
| InsuranceCompanyGroup | `model/inscogroup.go` |
| InsuranceModifiers | `model/insmod.go` |
| InternalServiceTypes | `model/internalservicetype.go` |
| Loinc | `model/loinc.go` |
| Pharmacy | `model/pharmacy.go` |
| PlaceOfService | `model/placeofservice.go` |
| Practices | `model/practice.go` |
| RouteOfAdministration | `model/routeofadmin.go` |

**Recommendation:** Create a generic `/api/support/:module/picklist/:query` endpoint using the existing picklist infrastructure (registered via `model/*.go` `init()` functions).

### No Go Model, No Handler (11 modules)

| Module | Notes |
|--------|-------|
| ClaimLogTable | Would need claimlog table migration |
| FaxStatus | Legacy fax tracking |
| Oids | Legacy OID registry |
| OrdersStock | Inventory tracking |
| OrdersTemplate | Order set templates |
| PhoneNumbers | Phone number management (separate from patient.phone) |
| RoomEquipment | Room/equipment tracking |
| RoomModule | Room management |
| TypeOfService | Service type codes |
| Taxonomy | Provider taxonomy codes |
| NPI | NPI registry (already in provider.npi field) |

---

## Tier 3: ADMIN / UTILITY Modules

Low priority — infrastructure, audits, translations.

| Module | Purpose | Priority |
|--------|---------|----------|
| CDRWBackup | CD/DVD backup system | OBSOLETE |
| LogModule | Audit logging | LOW — Prometheus covers this |
| ModuleFieldChecker | Schema validation tool | LOW — golang-migrate |
| ModuleFieldCheckerType | Field checker type registry | LOW |
| RecordLockModule | Record locking for concurrent edits | MEDIUM |
| UpdatesModule | Software update mechanism | OBSOLETE |
| WorkListsModule | Clinical worklists (task assignment) | MEDIUM |
| Translations | i18n translation management | LOW |
| i18nLanguages | Language registry | LOW |

---

## Tier 4: PHARMA / LEXICON Modules

Drug databases and clinical lexicons.

| Module | Purpose | Priority |
|--------|---------|----------|
| MultumDrugLexicon | Cerner Multum drug database | MEDIUM — drug interaction checking |
| NDCLexicon | National Drug Code registry | MEDIUM — e-prescribing |
| Xmr | Cross-MRN record linking | MEDIUM |
| XmrDefinition | XMR definition metadata | LOW |
| DiagnosisFamily | ICD diagnosis family groupings | LOW |

---

## Tier 5: Scheduling Sub-Systems

| Module | Purpose | Status |
|--------|---------|--------|
| AnesthesiologyCalendar | Anesthesia scheduling overlay | MISSING |
| CalendarGroup | Provider group calendars | MISSING (model removed in cleanup) |
| CalendarGroupAttendance | Group attendance tracking | MISSING |
| SchedulingRules | Rule engine for slot allocation | MISSING |
| ShimStation | Data shim/interface station | MISSING |

---

## ACTION PLAN: Recommended Implementation Order

### Batch A: Revenue Critical (Agata7 + Transport + ClaimLog)
- `Agata7` API: Implement Agata 7.0 billing format for claims export
- `Transport` API: Generic EDI transport layer
- `ClaimLog` module: Claims audit trail

### Batch B: Clinical Data Gaps (Patient Sub-Resources)
Pattern: Each becomes a patient sub-screen (like drug_samples, immunizations, etc.)
- `CurrentProblems` → `/api/patient/:id/problems` 
- `ChronicProblems` → `/api/patient/:id/chronic-problems`
- `EpisodeOfCare` → `/api/patient/:id/eoc` (already partially in eoc.go)
- `PreviousOperationsModule` → `/api/patient/:id/surgical-history`
- `Certifications` → `/api/patient/:id/certifications`
- `FinancialDemographics` → `/api/patient/:id/financial`
- `AppointmentFollowup` → `/api/patient/:id/followups`

### Batch C: Reference Data Endpoints
- Create generic picklist CRUD or at minimum read-list for 21 reference data models
- Priority: IcdCodes, CptCodes, Pharmacy, InsuranceCompanyModule

### Batch D: Frontend Gaps
- Patient sub-screens for each Batch B module
- Photo ID capture UI
- DICOM viewer (significant effort)
- Forms engine

### Batch E: Low Priority / Deprecate
- CDRWBackup: OBSOLETE
- FaxStatus: OBSOLETE (fax is legacy)
- UpdatesModule: OBSOLETE
- LogModule: Covered by Prometheus

---

## Summary Statistics

| Category | Count | Effort Estimate |
|----------|-------|-----------------|
| MISSING APIs (high impact) | 14 | 4-8 weeks |
| HIGH priority modules | 10 | 3-5 weeks |
| PARTIAL modules | 5 | 1-2 weeks |
| Reference data (needs handler) | 21 | 1-2 weeks (generic picklist) |
| Reference data (needs migration + handler) | 11 | 2-3 weeks |
| Admin/utility | 9 | 1 week (3 worth doing) |
| Pharma/lexicon | 5 | 2-4 weeks |
| Scheduling | 5 | 2-3 weeks |
| **TOTAL** | **~80 items** | **12-24 weeks** |
