# Top 6 + 25 Implementation Plan

## Priority Tier 1 — Top 6 (High Impact)

### 1. ACL/RBAC System (24 methods)
Backend: `POST/GET /api/acl/groups`, `POST/GET /api/acl/permissions`, `POST/DELETE /api/acl/users/:id/groups`
Tables: acl, usergroup, user ACL fields
Frontend: /admin/acl — permission matrix, group management

### 2. PaymentModule — Full Methods (12 methods)
Backend: `POST /api/payments/:id/attach-procedure`, `GET /api/payments/unattached`,
`POST /api/payments/:id/mistake`, `GET /api/coverages/copay-info`
Existing: /api/patient/:id/payments, /api/patient/:id/ledger

### 3. RemittBillingTransport (11 methods)
Backend: `POST /api/remitt/process-claims`, `POST /api/remitt/process-statement`,
`GET /api/remitt/patients-to-bill`, `GET /api/remitt/procedures-to-bill`,
`POST /api/remitt/mark-billed`
Existing: /api/remitt/status, /api/remitt/months

### 4. PatientModule — Address/Search Methods (5 methods)
Backend: `DELETE /api/patient/:id/addresses` (bulk), patient search refinements
Existing: individual address CRUD, basic search

### 5. FacilityModule (5 methods)
Backend: `GET /api/facilities`, `GET /api/facilities/default`
Existing: picklist only

### 6. CallIn (4 methods)
Backend: `GET /api/callin`, `GET /api/callin/:id`
Frontend: /patients/callin — phone triage log
Existing: none

## Priority Tier 2 — Medium (12 modules)

### Batch A: Annotations, Templates, Groups
- Annotations (6): `GET/POST /api/patient/:id/annotations`
- AppointmentTemplates (3): `GET /api/scheduler/templates/:id`
- CalendarGroup (2): `GET /api/scheduler/groups`, `GET /api/scheduler/groups/:id`

### Batch B: Reporting, Phone, Photo ID
- Reporting (4): `GET/POST /api/reports`, `GET /api/reports/:id`
- PhoneNumbers (4): `GET /api/patient/:id/phones`, `POST`
- PhotographicID (6): `GET /api/patient/:id/photo-id`, `POST`

### Batch C: Provider, RxRefill, Notifications
- ProviderModule (3): `GET /api/providers`, `GET /api/providers/lookup-npi`
- RxRefillRequest (1): `GET /api/rx-refill-requests`
- SystemNotifications full (6): timestamps, task inbox, counts

### Batch D: Labs, Workflow, Preferences, Tools
- LabsModule (1): `GET /api/patient/:id/labs`
- WorkflowStatus (3): `GET/POST /api/patient/:id/workflow-status`
- UserPreferences (3): `PUT /api/preferences` (write support)
- Tools (3): `GET /api/tools`, `POST /api/tools/:name`

## Priority Tier 3 — Low (7 modules)
GrowthCharts, Holiday, SchedulerBlockSlots, SMSProvider, SuperbillTemplate, Events, ProviderSpecialties
