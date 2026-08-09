-- FHIR FamilyMemberHistory: list all family history for a patient
-- name: FhirFamilyHistoryByPatient :many
SELECT
  id,
  patient,
  relationship,
  condition_name,
  icd10_code,
  onset_age,
  deceased,
  notes,
  user,
  active,
  created_at,
  updated_at
FROM family_history
WHERE patient = sqlc.arg(patient_id)
  AND active = 'active'
  AND deleted_at IS NULL
ORDER BY created_at DESC;

-- FHIR FamilyMemberHistory: get a single family history record by ID
-- name: FhirFamilyHistoryById :one
SELECT
  id,
  patient,
  relationship,
  condition_name,
  icd10_code,
  onset_age,
  deceased,
  notes,
  user,
  active,
  created_at,
  updated_at
FROM family_history
WHERE id = sqlc.arg(family_history_id)
  AND active = 'active'
  AND deleted_at IS NULL;
