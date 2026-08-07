# SQL Schema Cross-Reference Gap Analysis

145 SQL schemas from ../freemed/data/schema/mysql/ vs 89 PHP modules with methods vs 27 Go API modules + 24 patient sub-routes.

## 66 Missing Modules by Priority

### Tier 1 — High Impact (have complex methods, SQL tables exist)

**RemittBillingTransport (11 methods, billkey table):**
GetClaimInformation, getMonthlyReportsDetails, getMonthsInfo, GetRebillList, GetStatus,
MarkAsBilled, PatientsToBill, ProceduresToBill, ProcessClaims, ProcessStatement, rebillkeys

**PaymentModule (12 methods, payrec table):**
attachProcedure, CoverageIdFromType, CoverageToInsuranceName, getAdvancePaymentInfo,
getLastRecord, GetLedger, getUnAttachedCopays, getUnAttachedDeductables,
getUnAttachedPayments, IsAuthorized, PayerSelection, RemoveProcedureAsMistake

**ACL (24 methods, acl table):**
Full RBAC system — groups, permissions, user assignments, ACO management

**PatientModule (5 methods, patient table):**
DeleteAddressById, DeleteAddresses, GetAddresses, Search, SetAddresses

**FacilityModule (5 methods), CallIn (4 methods), Reporting (4 methods)**

### Tier 2 — Picklist/Display (1 method, SQL table, need list endpoint)

**Reference data (picklist-only, tables exist in Go but no list API):**
bodysite, cptcodes, cptmodifiers, documentcategory, drugforms, enclosuretypes,
i18nlanguages, icdcodes, loinc, placeofservice, routeofadministration,
coveragetypes, claimtypes, insurancecompanymodule, internalservicetypes,
pharmacy, providerspecialties

**Already served via /api/support/:module/picklist/:query** — these don't need dedicated handlers.
The DbSupportPicklists entries cover them.

### Tier 3 — Modular CRUD (SQL table exists, needs basic CRUD)

Annotations (6), AppointmentTemplates (3), CalendarGroup (2), Events (1),
GrowthCharts (1), Holiday (2), LabsModule (1), PatientReporting (3),
PhoneNumbers (4), PhotographicID (6), ProgressNotesTemplates (1),
ProviderGroups (2), ProviderModule (3), RxRefillRequest (1),
SchedulerBlockSlots (2), SchedulerPatientStatus (1),
SchedulerStatusType (1), SMSProvider (1), SuperbillTemplate (1),
SystemNotifications (6), Tools (3), UserPreferences (3),
WorkflowStatus (3), Xmr (1), XmrDefinition (2)

### Tier 4 — External Integration (complex, optional)

DicomModule (5), CDRWBackup (4), ShimStation (3), BillingClearinghouse (1),
BillingContact (1), BillingService (1)

### Tier 5 — Stub/Empty (1 method, table may not have meaningful data)

BillKey (1), Certifications (1), EncounterNotesTemplate (2), Forms (1),
FormTemplates (1), ModuleFieldChecker (2), ModuleFieldCheckerType (1),
PatientIds (1), Translations (1), MultumDrugLexicon (2), NDCLexicon (3),
LettersTemplates (1)

## Already Migrated (23 of 89 PHP modules)

Prescriptions, Medications, Allergies, Procedures, Payments, PatientCoverages,
PatientTag, Superbills, Referrals, Immunizations, DrugSamples, EpisodeOfCare,
EncounterNotes, ProgressNotes, Vitals, ScannedDocuments, Messages,
SystemNotifications (partial), UserPreferences (read-only), SchedulerBlocks (partial),
ClinicRegistration (patients/new), UnfiledDocuments, UnreadDocuments

## Remaining Picklist-Only Tables (no handler needed, use /api/support/...)

These already have DbSupportPicklists entries:
bccdc, billingcontact, billingservice, bodysite, cpt, cptmodifier,
documentcategory, drugforms, drugquantityqualifier, drugsampleinventory,
enclosuretype, facility, i18nlanguages, icdcode, insco, inscogroup, insmod,
internalservicetype, loinc, placeofservice, practice, provider, routeofadmin,
schedulerstatustype, coveragetypes, claimtype

## Net Gap: ~25 modules worth implementing

After filtering out picklist-only, stub/empty, and already-migrated modules,
approximately 25 modules remain that would benefit from dedicated Go handlers:

- **High** (6): RemittBillingTransport, PaymentModule full methods, ACL, PatientModule, FacilityModule, CallIn
- **Medium** (12): Annotations, AppointmentTemplates, CalendarGroup, Reporting, PhoneNumbers, PhotographicID, ProviderModule, RxRefillRequest, SystemNotifications (full), UserPreferences (write), WorkflowStatus, LabsModule
- **Low** (7): Events, GrowthCharts, Holiday, SchedulerBlockSlots, SMSProvider, SuperbillTemplate, Tools
