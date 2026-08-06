-- Allergies: list active allergies for a patient
-- name: ListAllergies :many
SELECT
  id,
  created_at,
  updated_at,
  deleted_at,
  patient,
  active
FROM allergies
WHERE patient = sqlc.arg(patient_id) AND active = 'active'
ORDER BY created_at DESC;

-- Allergies: create a new allergy record
-- name: CreateAllergy :execresult
INSERT INTO allergies (
  created_at, updated_at, patient, active
) VALUES (
  NOW(), NOW(), sqlc.arg(patient_id), 'active'
);

-- Allergies: deactivate an allergy (soft delete)
-- name: DeactivateAllergy :exec
UPDATE allergies
SET active = 'inactive', updated_at = NOW()
WHERE id = sqlc.arg(allergy_id);
