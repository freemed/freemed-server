-- FHIR Condition: list all problems (current + chronic) for a patient
-- name: FhirConditionsByPatient :many
SELECT
  cp.id AS condition_id,
  cp.patient AS condition_patient,
  cp.date AS condition_date,
  cp.problem AS condition_text,
  'current' AS condition_type,
  cp.active AS condition_active,
  cp.updated_at AS condition_updated
FROM current_problems cp
WHERE cp.patient = sqlc.arg(patient_id)
  AND cp.deleted_at IS NULL
UNION ALL
SELECT
  ch.id AS condition_id,
  ch.patient AS condition_patient,
  ch.date AS condition_date,
  ch.problem AS condition_text,
  'chronic' AS condition_type,
  NULL AS condition_active,
  ch.updated_at AS condition_updated
FROM chronic_problems ch
WHERE ch.patient = sqlc.arg(patient_id)
  AND ch.deleted_at IS NULL
ORDER BY condition_date DESC;

-- FHIR Condition: lookup a single problem by ID (check current first, then chronic)
-- name: FhirConditionById :one
SELECT
  cp.id AS condition_id,
  cp.patient AS condition_patient,
  cp.date AS condition_date,
  cp.problem AS condition_text,
  'current' AS condition_type,
  cp.active AS condition_active,
  cp.updated_at AS condition_updated
FROM current_problems cp
WHERE cp.id = sqlc.arg(condition_id)
  AND cp.deleted_at IS NULL
UNION ALL
SELECT
  ch.id AS condition_id,
  ch.patient AS condition_patient,
  ch.date AS condition_date,
  ch.problem AS condition_text,
  'chronic' AS condition_type,
  NULL AS condition_active,
  ch.updated_at AS condition_updated
FROM chronic_problems ch
WHERE ch.id = sqlc.arg(condition_id)
  AND ch.deleted_at IS NULL
LIMIT 1;
