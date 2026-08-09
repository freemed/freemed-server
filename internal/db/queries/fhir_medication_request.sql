-- FHIR MedicationRequest: list prescriptions for a patient
-- name: FhirMedicationRequestsByPatient :many
SELECT
  id,
  patient,
  drug_name,
  dosage,
  frequency,
  quantity,
  refills,
  date_written,
  prescribing_provider,
  pharmacy,
  status,
  notes,
  created_at,
  updated_at
FROM prescriptions
WHERE patient = sqlc.arg(patient_id)
  AND deleted_at IS NULL
ORDER BY date_written DESC;

-- FHIR MedicationRequest: lookup a single prescription by ID
-- name: FhirMedicationRequestById :one
SELECT
  id,
  patient,
  drug_name,
  dosage,
  frequency,
  quantity,
  refills,
  date_written,
  prescribing_provider,
  pharmacy,
  status,
  notes,
  created_at,
  updated_at
FROM prescriptions
WHERE id = sqlc.arg(prescription_id)
  AND deleted_at IS NULL;
