# GWT UI Gap Implementation Plan

> Implement scheduler and overall UI gaps from GWT differential analysis in priority order.

## Phase A: Scheduler Completeness (backend exists for most)

### A.1 Recurring Appointments
- Backend: `POST /api/scheduler/recurring` — accept `{id, timestamps[], description}`
- sqlc: `CreateRecurringAppointments :exec` — batch insert
- Frontend: "Make Recurring" button in event modal, weekly/biweekly/monthly picker

### A.2 NextAvailable Slot Finder
- Backend: `POST /api/scheduler/next-available` — accept `{date, provider_id, duration}`
- sqlc: query existing appointments, compute first free slot
- Frontend: "Find Next Available" button in create form

### A.3 Group Appointments UI
- Backend: `POST /api/scheduler/group` already exists
- Frontend: "Create Group" button, multi-patient picker, group detail view
- Frontend: Copy/Move group in scheduler page

### A.4 Provider Filter + Templates
- Frontend: provider dropdown in scheduler toolbar (show/hide providers)
- Frontend: appointment template picklist in create form
- Backend: `GET /api/scheduler/templates` — list appointment templates (picklist already exists)

### A.5 Block Time Slots
- Frontend: render blocked slots on calendar
- Backend: `GET /api/scheduler/blocks/:date` — list blocked slots

---

## Phase B: Patient Sub-Screens (backend exists for most)

### B.1 Drug Sample Entry
- Backend: `POST /api/patient/:id/drug-samples`, `GET /api/patient/:id/drug-samples`
- Frontend: drug sample entry form, list page

### B.2 Immunization Entry
- Backend: `POST /api/patient/:id/immunizations`, `GET /api/patient/:id/immunizations`
- Frontend: immunization form + history list

### B.3 Referral Entry
- Backend: `POST /api/patient/:id/referrals`, `GET /api/patient/:id/referrals`
- Frontend: referral form + list

### B.4 Episode of Care
- Backend: `GET /api/patient/:id/episodes-of-care`, `POST`
- Frontend: EOC management page

### B.5 Growth Charts
- Frontend: BMI/weight/height chart visualization component

---

## Phase C: Admin + Documents + Billing

### C.1 User Management Frontend
- Frontend: `/admin/users` — list, create, edit users
- Uses existing `GET /api/users` endpoint, add PUT/POST/DELETE

### C.2 Unfiled/Unread Documents
- Backend: wire up existing unfiled/unread tables
- Frontend: document routing queues

### C.3 Billing Screens
- Frontend: `/billing/claims` — claims manager
- Frontend: `/billing/ar` — accounts receivable aging
- Uses existing backend endpoints
