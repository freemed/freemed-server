# FreeMED Feature Parity Analysis — PHP 0.9.x vs Go SvelteKit

> Generated from 133 PHP module files in `../freemed/lib/org/freemedsoftware/module/`
> compared against current Go backend (`api/`, `model/`) + SvelteKit frontend (`frontend/src/routes/`)

## Summary

| Category | PHP Modules | Go Implemented | Gap |
|----------|------------|----------------|-----|
| Core/Admin | 12 | 4 | 8 |
| Patient | 14 | 4 | 10 |
| Clinical/EMR | 20 | 2 | 18 |
| Billing/Claims | 18 | 0 | 18 |
| Scheduling | 10 | 3 | 7 |
| Pharmacy/Rx | 8 | 0 | 8 |
| Reporting | 5 | 0 | 5 |
| DICOM/Imaging | 2 | 0 | 2 |
| Messaging/Notifications | 6 | 1 | 5 |
| Infrastructure/Utils | 12 | 3 | 9 |
| Reference Data | 26 | 22 | 4 |
| **TOTAL** | **133** | **39** | **94** |

---

## Detailed Gap Analysis

### 1. Core / Administration (12 modules, 4 implemented)

| PHP Module | Public Methods | Go Status | Priority |
|------------|---------------|-----------|----------|
| **ACL** | `acl`, `AddAllowedACOs`, `AddBlockedACOs`, `AddGroupWithPermissions`, `AddUserToGroup`, `DelAllowedACOs`, `DelBlockedACOs`, `DelGroupWithPermissions`, `GetAllowedACOs`, `GetAllPermissions`, `GetBlockedACOs`, `GetGroupPermissions`, `getIDByName`, `getModulesPermissionsBits`, `GetUserGroupNames`, `GetUserGroups`, `GetUserPermissions`, `ModGroupWithPermissions`, `RemoveUserFromGroup`, `UserAdd`, `UserDel`, `UserGroups`, `UserInGroup` | MISSING | HIGH |
| **UserPreferences** | `GetAll`, `GetConfigSections`, `SetValues` | `api/config.go` (read-only) | MEDIUM |
| **UserGroups** | (empty) | MISSING | MEDIUM |
| **Certifications** | `getCertifications` | MISSING | LOW |
| **ProviderCertifications** | (empty) | MISSING | LOW |
| **ProviderGroups** | `getGroupIds`, `getProviderIds` | MISSING | MEDIUM |
| **ProviderSpecialties** | (empty) | MISSING | LOW |
| **ProviderStatus** | (empty) | MISSING | LOW |
| **UpdatesModule** | `GetFeed` | MISSING | LOW |
| **Tools** | `ExecuteTool`, `GetToolParameters`, `GetTools` | MISSING | LOW |
| **Translations** | `mod` | MISSING | LOW |
| **CDRWBackup** | `device`, `driver`, `RunBackup` | MISSING | LOW |

### 2. Patient Management (14 modules, 4 implemented)

| PHP Module | Public Methods | Go Status | Priority |
|------------|---------------|-----------|----------|
| **PatientModule** | `DeleteAddressById`, `DeleteAddresses`, `GetAddresses`, `Search`, `SetAddresses` | `api/patients.go` (search, create), `api/patient.go` (view) | MEDIUM |
| **PatientIds** | (empty) | `model/patient_ids.go` (table only) | LOW |
| **PatientTag** | `AdvancedTagSearch`, `ChangeTag`, `CreateTag`, `ExpireTag`, `ListTags`, `SimpleTagSearch`, `TagsForPatient` | MISSING | MEDIUM |
| **PatientLink** | (empty) | MISSING | LOW |
| **PatientStatus** | (empty) | MISSING | LOW |
| **PatientLocation** | (empty) | MISSING | LOW |
| **PatientCoverages** | `GetAllCoverages`, `GetAllCoveragesWithDetail`, `GetCoverageByType`, `GetCoverages`, `GetPrimaryCoverage`, `RemoveOldCoverage` | MISSING | HIGH |
| **PatientReporting** | `GenerateReport`, `GetReportParameters`, `GetReports` | MISSING | MEDIUM |
| **PatientCorrespondence** | (empty) | MISSING | LOW |
| **ClinicRegistration** | `createPatient`, `GetAll`, `migrateToPatient` | Partially (`POST /patients`) | MEDIUM |
| **ClinicOrders** | (empty) | MISSING | LOW |
| **FinancialDemographics** | (empty) | MISSING | LOW |
| **PhotographicIdentification** | `GetDocumentPage`, `GetPhotoID`, `ImportMugshotPhoto`, `NumberOfPages`, `UploadPhotoID`, `UploadPhotoIDInline` | MISSING | LOW |
| **PhoneNumbers** | `GetRecentNumber`, `GetTypeNumber` | MISSING | LOW |

### 3. Clinical / EMR (20 modules, 2 implemented)

| PHP Module | Public Methods | Go Status | Priority |
|------------|---------------|-----------|----------|
| **ProgressNotes** | `CalculateBMI`, `NoteForDate` | `api/patient_progress_notes.go` | HIGH |
| **ProgressNotesTemplates** | `GetTemplate` | MISSING | MEDIUM |
| **EncounterNotes** | `getEncounterNoteInfo`, `getEncountersList` | Placeholder route only | HIGH |
| **EncounterNotesTemplate** | `getTemplateInfo`, `getTemplates` | MISSING | MEDIUM |
| **Vitals** | (empty) | MISSING | HIGH |
| **GrowthCharts** | `GetGrowthChartValues` | MISSING | LOW |
| **Allergies** | `GetAtoms`, `GetMostRecent`, `SetAtoms`, `SetAtoms2` | MISSING | HIGH |
| **Medications** | `GetAtoms`, `GetMostRecent`, `SetAtoms` | MISSING | HIGH |
| **CurrentProblems** | (empty) | MISSING | MEDIUM |
| **ChronicProblems** | (empty) | MISSING | MEDIUM |
| **DiagnosisFamily** | (empty) | MISSING | LOW |
| **EpisodeOfCare** | `getAllValues`, `getEOCValues`, `getHospitalizations` | MISSING | MEDIUM |
| **LabsModule** | `GetLabValues` | MISSING | MEDIUM |
| **Immunizations** | (empty) | `model/immunization.go` (table only) | MEDIUM |
| **ClinicalOrders** | (empty) | MISSING | LOW |
| **PreviousOperationsModule** | (empty) | MISSING | LOW |
| **Referrals** | `GetAllActiveByPatient` | MISSING | MEDIUM |
| **Forms** | `print` | MISSING | MEDIUM |
| **FormTemplates** | `GenerateUUID` | MISSING | LOW |
| **DicomModule** | `CheckForDuplicates`, `GetDICOM`, `LookupPatient`, `UploadDICOM`, `UploadDICOMInline` | MISSING | LOW |

### 4. Billing / Claims (18 modules, 0 implemented)

| PHP Module | Public Methods | Go Status | Priority |
|------------|---------------|-----------|----------|
| **BillingClearinghouse** | (empty) | `model/clearinghouse.go` (table only) | HIGH |
| **BillingContact** | (empty) | `model/billingcontact.go` (table only) | HIGH |
| **BillingService** | (empty) | `model/billingservice.go` (table only) | HIGH |
| **BillKey** | (empty) | `model/billkey.go` (table only) | HIGH |
| **RemittanceQueue** | (empty) | MISSING | HIGH |
| **RemittBillingTransport** | `GetClaimInformation`, `getMonthlyReportsDetails`, `getMonthsInfo`, `GetRebillList`, `GetStatus`, `MarkAsBilled`, `PatientsToBill`, `ProceduresToBill`, `ProcessClaims`, `ProcessStatement`, `rebillkeys` | MISSING | HIGH |
| **PaymentModule** | `attachProcedure`, `CoverageIdFromType`, `CoverageToInsuranceName`, `getAdvancePaymentInfo`, `getLastRecord`, `GetLedger`, `getUnAttachedCopays`, `getUnAttachedDeductables`, `getUnAttachedPayments`, `IsAuthorized`, `PayerSelection`, `RemoveProcedureAsMistake` | `model/payments.go` (table only) | HIGH |
| **ProcedureModule** | `CalculateCharge`, `GetAuthorizations`, `getCoverages`, `getLastProc`, `getNonZeroBalProcs`, `getPatientProcHistory`, `getProcByID`, `getProcedureInfo`, `getTotalArrears` | `model/procedure.go` (table only) | HIGH |
| **ClaimLogTable** | (empty) | `model/claimlog.go` (table only) | MEDIUM |
| **ClaimTypes** | `getClaimTypes` | `model/claimtype.go` (table only) | MEDIUM |
| **SuperBill** | `GetForDates`, `GetSuperbill`, `MarkAsHandled`, `printSuperbills`, `ProcessSuperbills` | MISSING | MEDIUM |
| **SuperbillTemplate** | `GetTemplate` | MISSING | LOW |
| **Authorizations** | `getActionItems`, `getActionItemsCount`, `getActionItemsQuery`, `GetAllAuthorizations`, `GetAllAuthorizationsWithDetail`, `getValidAuthorizations` | `model/authorizations.go` (table only) | HIGH |
| **CoverageTypes** | (empty) | `model/coveragetypes.go` (table only) | MEDIUM |
| **InsuranceCompanyModule** | (empty) | `model/insco.go` (table only) | MEDIUM |
| **InsuranceCompanyGroup** | (empty) | `model/inscogroup.go` (table only) | MEDIUM |
| **InsuranceModifiers** | (empty) | `model/insmod.go` (table only) | MEDIUM |
| **Xmr** / **XmrDefinition** | `SetElements` / `GetFormElementsWithDefaults` | MISSING | LOW |

### 5. Scheduling (10 modules, 3 implemented)

| PHP Module | Public Methods | Go Status | Priority |
|------------|---------------|-----------|----------|
| **SchedulerTable** | (empty) | `api/scheduler.go` (view, range, reschedule, cancel) | — |
| **SchedulerBlockSlots** | `GetAll`, `GetBlockedTimeSlots` | MISSING | LOW |
| **SchedulerPatientStatus** | `getPatientStatus` | MISSING | LOW |
| **SchedulerStatusType** | `getStatusType` | `model/schedulerstatustype.go` (table only) | LOW |
| **AppointmentTemplates** | `get` | `model/appttemplate.go` (table only) | LOW |
| **AppointmentFollowup** | (empty) | MISSING | LOW |
| **SchedulingRules** | (empty) | MISSING | LOW |
| **CalendarGroup** | `GetAll`, `GetDetailedRecord` | `model/calendargroup.go` (table only) | LOW |
| **CalendarGroupAttendance** | (empty) | `model/calendargroupattendance.go` (table) | LOW |
| **AnesthesiologyCalendar** | (empty) | MISSING | LOW |

### 6. Pharmacy / Prescriptions (8 modules, 0 implemented)

| PHP Module | Public Methods | Go Status | Priority |
|------------|---------------|-----------|----------|
| **Prescription** | `GetDistinctRx` | MISSING | HIGH |
| **RxRefillRequest** | `GetAll` | MISSING | MEDIUM |
| **DrugSampleInventory** | `Deduct` | `model/drugsampleinventory.go` (table only) | MEDIUM |
| **DrugSamples** | (empty) | MISSING | LOW |
| **MultumDrugLexicon** | `DosagesForDrug`, `DrugDosageToText` | MISSING | MEDIUM |
| **NDCLexicon** | `DosagesForDrug`, `DrugStrengthToText`, `NameLookupToText` | MISSING | MEDIUM |
| **Pharmacy** | `picklist` | Table in schema, no handler | MEDIUM |
| **RouteOfAdministration** | (empty) | `model/routeofadmin.go` (table only) | LOW |

### 7. Reporting (5 modules, 0 implemented)

| PHP Module | Public Methods | Go Status | Priority |
|------------|---------------|-----------|----------|
| **Reporting** | `GenerateReport`, `GetReport`, `GetReportParameters`, `GetReports` | MISSING | MEDIUM |
| **ReportingPrintLog** | `GetAllRecords` | MISSING | LOW |
| **PatientReporting** | `GenerateReport`, `GetReportParameters`, `GetReports` | MISSING | MEDIUM |
| **WorkListsModule** | `GenerateWorkList`, `GenerateWorklists`, `ProcessChange` | MISSING | LOW |
| **Rules** | (empty) | MISSING | LOW |

### 8. DICOM / Imaging (2 modules, 0 implemented)

| PHP Module | Public Methods | Go Status | Priority |
|------------|---------------|-----------|----------|
| **DicomModule** | `CheckForDuplicates`, `GetDICOM`, `LookupPatient`, `UploadDICOM`, `UploadDICOMInline` | MISSING | LOW |
| **ShimStation** | `dateDiff`, `GetAll`, `getStationsByType` | MISSING | LOW |

### 9. Messaging / Notifications (6 modules, 1 implemented)

| PHP Module | Public Methods | Go Status | Priority |
|------------|---------------|-----------|----------|
| **MessagesModule** | `DeleteMultiple`, `GetAllByTag`, `MessageTags`, `UnreadMessages` | `api/messages.go` (view, list users, send) | MEDIUM |
| **SystemNotifications** | `GetFromTimestamp`, `GetSystemTaskInboxCount`, `GetSystemTaskPatientInbox`, `GetSystemTaskPatientInboxCount`, `GetSystemTaskUserInbox`, `GetTimestamp` | `model/systemnotification.go` (table) | MEDIUM |
| **Notifications** | `notify` | MISSING | LOW |
| **FaxStatus** | (empty) | MISSING | LOW |
| **SmsProvider** | `SendSMSToUser` | MISSING | LOW |
| **Events** | `GetEvents` | MISSING | LOW |

### 10. Infrastructure / Utilities (12 modules, 3 implemented)

| PHP Module | Public Methods | Go Status | Priority |
|------------|---------------|-----------|----------|
| **RecordLockModule** | (empty) | MISSING | MEDIUM |
| **LogModule** | (empty) | MISSING | LOW |
| **ModuleFieldChecker** | `getUncompletedItems`, `getUncompletedItemsCount` | MISSING | LOW |
| **ModuleFieldCheckerType** | `getModuleInfo` | MISSING | LOW |
| **UnfiledDocuments** | `batchSplit`, `faxback`, `GetAll`, `GetCount`, `GetDocumentPage`, `NumberOfPages` | MISSING | LOW |
| **UnreadDocuments** | `GetAll`, `GetCount`, `GetDocumentPage`, `GetLocalCachedFile`, `MoveToAnotherProvider`, `NumberOfPages`, `ReviewIntoRecord` | MISSING | LOW |
| **ScannedDocuments** | `GetDocumentPage`, `GetDocumentPdf`, `GetPatientAllRecords`, `NumberOfPages` | MISSING | LOW |
| **Letters** / **LettersTemplates** | (empty) / `GetTemplate` | MISSING | LOW |
| **Holiday** | `CheckForHoliday`, `GetHolidayDesc` | MISSING | LOW |
| **Zipcodes** | `CalculateDistance`, `CityStateZipPicklist` | `api/zipcodes.go` (picklist only) | LOW |
| **NPI** | (empty) | MISSING | LOW |
| **Taxonomy** | (empty) | MISSING | LOW |

---

## Priority Implementation Order

### Tier 1 — Core Clinical (blocking basic EMR usage)
1. **Allergies** — `GetAtoms`, `SetAtoms` (high-frequency, patient safety)
2. **Medications** — `GetAtoms`, `SetAtoms` (high-frequency)
3. **Vitals** — create vitals table + CRUD
4. **EncounterNotes** — `getEncountersList`, `getEncounterNoteInfo`
5. **ProgressNotes** — `CalculateBMI`, `NoteForDate` (partially done)

### Tier 2 — Billing Foundation (revenue-critical)
6. **ProcedureModule** — `getPatientProcHistory`, `getProcedureInfo`, `CalculateCharge`
7. **PaymentModule** — `GetLedger`, payment CRUD
8. **Authorizations** — `GetAllAuthorizations`, `getValidAuthorizations`
9. **PatientCoverages** — `GetAllCoverages`, `GetPrimaryCoverage`
10. **BillingClaimLog** — claim submission tracking

### Tier 3 — Patient Management
11. **PatientTag** — tag CRUD + search
12. **PatientAddress** — `SetAddresses`, `DeleteAddressById` (partially in create)
13. **ACL / RBAC** — permission system (currently `return true`)

### Tier 4 — Scheduling Enhancements
14. **SchedulerBlockSlots** — block management
15. **AppointmentTemplates** — template CRUD
16. **CalendarGroup** — group appointments

### Tier 5 — Pharmacy
17. **Prescription** — Rx CRUD
18. **RxRefillRequest** — refill workflow
19. **NDCLexicon** — drug lookup

### Tier 6 — Everything Else
20. Reports, DICOM, Letters, Documents, Scanning, CDRW, etc.

---

## Quick Wins (Reference Data Already in Schema)

These PHP modules only serve picklist/display data and have **Go model tables already**:
- `model/bodysite.go`, `model/cpt.go`, `model/cptmodifier.go`, `model/documentcategory.go`
- `model/drugforms.go`, `model/drugquantityqualifier.go`, `model/enclosuretype.go`
- `model/i18nlanguages.go`, `model/icdcode.go`, `model/internalservicetype.go`
- `model/loinc.go`, `model/placeofservice.go`, `model/routeofadmin.go`
- `model/coveragetypes.go`, `model/claimtype.go`, `model/insco.go`
- `model/inscogroup.go`, `model/insmod.go`, `model/facility.go`
- `model/practice.go`, `model/provider.go`, `model/workflow_status.go`
- `model/workflow_status_type.go`

These 22 tables need simple `GET /api/{name}` list endpoints (~15 min each). The PHP picklist system already exists as `api/support.go` — just add them to `DbSupportPicklists`.

## Modules with # Public Methods = 0

69 of 133 PHP modules have zero public methods — they're abstract base classes, table definitions, or empty stubs. These can be ignored for feature parity.
