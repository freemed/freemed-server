# GWT UI Differential Analysis

> FreeMED 0.9.x GWT UI (84+ screens, 626 files) vs current SvelteKit 5 frontend (13 routes, 10 pages)

## Scheduler Comparison (Primary Focus)

### GWT SchedulerWidget (2,806 lines)

**Views:**
- Day view (default, 6AM-12AM, 15-min intervals)
- Month view (via WholeDayField date renderer)
- Provider-specific views (by provider dropdown)

**Event Management:**
- Click event → detail popup with patient name, provider, time, duration, note, status
- SetAppointment (create new) — form with patient picklist, provider picklist, facility picklist, room picklist, appointment template picklist, date/time, duration, note, status, resource type (patient/resource)
- CopyAppointment — copy to new date (with or without notes)
- MoveAppointment — drag-drop to new time slot
- Cancel appointment — sets calstatus='cancelled'
- Recursive/Repeating appointments — SetRecurringAppointment with timestamps
- NextAvailable — find next available slot for a provider

**Group Appointments:**
- CopyGroupAppointment — copy entire group to new date
- MoveGroupAppointment — move entire group
- SetGroupAppointment — create group with multiple patient IDs
- FindGroupAppointments — list by group ID
- FindGroupAppointmentsDates — list dates with group appointments
- Group detail screen spawned via spawnGroupScreen()

**Block/Resource Management:**
- Retrieve/Render blocked time slots
- Room/Equipment scheduling via resource type
- Provider schedule view per-physician

**Templates:**
- AppointmentTemplates picklist for quick population of duration/type/status
- Template attribute: duration, appointment type, default status, color

**Import/Export:**
- ImportDate — import schedules from date to selected date
- Export/print capabilities (via custom CSS)

**Other Features:**
- CanBookAppointment — conflict checking before booking
- Scroll through dates (next/previous)
- Date picker for jumping to specific date
- Custom event colors from appointment templates
- Provider filter dropdown to show/hide providers
- Resource type: "Patient" vs "Resource" (room/equipment)

### Current SvelteKit Scheduler (`frontend/src/routes/scheduler/+page.svelte`)

**Implemented:**
- Week/month/day views (FullCalendar 6)
- Event loading via GET /scheduler/dailyapptrange/:from/:to
- Event detail modal with date/time/duration/provider/patient/note/status
- Event reschedule (drag-drop)
- Event resize (duration change)
- Create appointment modal (patient, provider, date, hour, minute, duration, note)
- Copy appointment (to new date/time)
- Cancel appointment (DELETE)
- Calendar refresh after actions

**Missing:**
- Recurring appointments (SetRecurringAppointment)
- NextAvailable slot finder
- Group appointments UI (create/copy/move groups with multiple patients)
- Room/Equipment scheduling
- Block time slot management
- Provider-specific views
- Import from date
- Export/print
- Conflict checking (CanBookAppointment)
- Appointment templates dropdown in create form
- Resource type selection

---

## Top-Level Screen Comparison

### System Category

| GWT Screen | Purpose | SvelteKit Status |
|-----------|---------|-----------------|
| Dashboard | Aggregated stats, action items, notifications | `/` — basic stats cards, no action items |
| Scheduler | Full calendar with booking | `/scheduler` — see above |
| Messages | Inbox with tags | `/messages` — inbox, compose |
| ClinicRegistration | Create patient from registration form | `/patients/new` — patient create form |
| Triage | Triage patient queue | MISSING |

### Patient Category

| GWT Screen | Purpose | SvelteKit Status |
|-----------|---------|-----------------|
| PatientSearch | Advanced search with picklists | `/patients` — search + results table |
| NewPatient | Create new patient | `/patients/new` — form |
| Groups | Patient groups management | MISSING |
| CallIn | Call-in log / phone triage | MISSING |
| RxRefill | Prescription refill requests | MISSING |
| TagSearch | Search patients by tags | Backend exists, no frontend |

### Patient Sub-Screens (inside PatientScreen tabs)

| GWT Screen | Purpose | SvelteKit Status |
|-----------|---------|-----------------|
| AllergyEntry | Add allergy | `/patients/:id/allergies` — list + add |
| DrugSampleEntry | Log drug sample | MISSING |
| EpisodeOfCare | Manage EOC | MISSING |
| EncounterScreen | Encounter notes | `/patients/:id/encounters` — data wired |
| FormEntry | Fill forms | MISSING |
| ImmunizationEntry | Add immunization | MISSING |
| LetterEntry | Generate letter | MISSING |
| PatientCorrespondenceEntry | Log correspondence | MISSING |
| PatientLinkEntry | Link patient records | MISSING |
| ProgressNoteEntry | Add progress note | `/patients/:id/progress-notes` |
| PrescriptionScreen | Manage Rx | Backend exists, no frontend |
| ReferralEntry | Add referral | MISSING |
| VitalsEntry | Record vitals | `/patients/:id/vitals` — list + add |
| ScannedDocumentsEntry | View scanned docs | MISSING |
| ProcedureScreen | Manage procedures | Backend exists, no frontend |
| AdvancePayment | Record payment | Backend exists, no frontend |
| PatientReportingScreen | Patient-specific reports | MISSING |
| GrowthChartScreen | Growth charts | MISSING |
| ClinicalOrdersEntry | Clinical orders | MISSING |
| ForeignId (PatientIdEntry) | External IDs | MISSING |

### Documents Category

| GWT Screen | Purpose | SvelteKit Status |
|-----------|---------|-----------------|
| UnfiledDocuments | Batch-split/route unfiled docs | MISSING |
| UnreadDocuments | Review/route unread docs | MISSING |

### Billing Category

| GWT Screen | Purpose | SvelteKit Status |
|-----------|---------|-----------------|
| AccountsReceivable | AR aging + collections | MISSING |
| ClaimsManager | Claim submission + tracking | MISSING |
| RemittBilling | Remittance processing | MISSING |
| SuperBills | Superbill generation + handling | MISSING |

### Reporting Category

| GWT Screen | Purpose | SvelteKit Status |
|-----------|---------|-----------------|
| ReportingEngine | Run reports | MISSING |
| ReportingLog | Report history | MISSING |

### Utilities Category

| GWT Screen | Purpose | SvelteKit Status |
|-----------|---------|-----------------|
| ToolsScreen | Execute server tools | MISSING |
| SupportData | Manage reference data tables | Partially via picklist system |
| FieldChecker | Module completeness checker | MISSING |
| UserManagement | Manage users/groups/permissions | Backend exists (GET /users), no frontend |
| SystemConfiguration | Admin config | `/settings` — read-only display |
| ACL | Access control management | MISSING |
| DBAdministration | Database maintenance | MISSING (commented out in GWT) |

---

## Key Widgets Present in GWT, Missing in SvelteKit

| GWT Widget | Purpose |
|-----------|---------|
| LedgerWidget | Full patient ledger with procedures + payments + writeoffs |
| LedgerPopup | Quick ledger view popup |
| RemittBillingWidget | Billing transport management |
| FinancialWidget | Financial demographics |
| PatientInfoBar | Patient info header bar on all patient screens |
| PatientWidget | Demographics form (edit patient) |
| PatientAddresses | Address management widget |
| PatientAuthorizations | Authorization list |
| PatientCoverages | Coverage/insurance list |
| PatientTagsWidget | Tag management |
| PatientTagWidget | Single tag entry |
| PatientProblemList | Problem list / Dx |
| EMRModuleWidget | Generic EMR data display |
| EncounterWidget | Encounter note entry/edit |
| EncounterTemplateWidget | Template-based encounters |
| NotesBox | Note entry with rich text |
| DocumentBox | Document viewer |
| DocumentThumbnailsWidget | Thumbnail strip |
| DrugWidget | Medication list + entry |
| RecentAllergiesList | Quick allergy reference |
| RecentMedicationsList | Quick medication reference |
| PharmacyWidget | Pharmacy lookup |
| ProviderWidget | Provider entry |
| SignatureWidget | Signature pad capture |
| ActionItemsBox | Action items panel |
| EventsWidget | System events/notifications |
| AgingSummaryWidget | AR aging summary |
| ReportingWidget | Report parameter entry |
| ClaimDetailsWidget | Claim detail view |
| CptEdit | CPT code picker |
| GeneratedFormWidget | Dynamic form renderer |
| ImageCompositedForm | Image + overlay form |
| MugshotWebcamWidget | Photo ID capture |
| DjvuViewer | Document format viewer |
| WorkList | Worklist generation |
| BlockScreenWidget | Block time slots |
| MessageView | Message reader |
| MessageBox | Message compose |
| Popup | Generic popup |
| PopupView | Popup with embedded content |

---

## Priority Gaps — Scheduler

1. **Recurring appointments** — SetRecurringAppointment (backend + frontend)
2. **NextAvailable** — find next free slot for provider (backend + frontend)
3. **Group appointments UI** — backend exists, needs frontend create/copy/move
4. **Provider-specific views** — filter calendar by provider
5. **Block time slots** — retrieve/render blocked slots
6. **Import/Export** — import schedules, export/print
7. **Appointment templates** — picklist in create form
8. **Conflict checking** — canBookAppointment before booking

## Priority Gaps — Overall UI

1. **Triage / Call In** — patient queue + phone triage
2. **Billing screens** — AR, Claims Manager, Remitt Billing, Superbills
3. **Patient sub-screens** — Drug Sample, Episode of Care, Immunization, Letters, Referrals, Scanned Documents, Growth Charts, Clinical Orders
4. **Documents** — Unfiled/Unread document routing
5. **Reporting** — Report engine + log
6. **User Management** — User CRUD frontend
7. **ACL** — Permission management frontend
8. **Tools** — Server tool execution
