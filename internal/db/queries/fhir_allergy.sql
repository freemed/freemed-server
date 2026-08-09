-- FHIR AllergyIntolerance: list active allergies for a patient
-- name: FhirAllergiesByPatient :many
SELECT
  id,
  patient,
  active,
  created_at,
  updated_at
FROM allergies
WHERE patient = sqlc.arg(patient_id)
  AND active = 'active'
  AND deleted_at IS NULL
ORDER BY created_at DESC;

-- FHIR AllergyIntolerance: lookup a single allergy by ID
-- name: FhirAllergyById :one
SELECT
  id,
  patient,
  active,
  created_at,
  updated_at
FROM allergies
WHERE id = sqlc.arg(allergy_id)
  AND deleted_at IS NULL;
