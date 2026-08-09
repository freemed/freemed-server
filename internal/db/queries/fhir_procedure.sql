-- FHIR Procedure: list procedures (previous_operations) for a patient
-- name: FhirProceduresByPatient :many
SELECT
  id,
  patient,
  operation_date,
  operation,
  user,
  created_at,
  updated_at
FROM previous_operations
WHERE patient = sqlc.arg(patient_id)
  AND deleted_at IS NULL
ORDER BY operation_date DESC;

-- FHIR Procedure: lookup a single procedure by ID
-- name: FhirProcedureById :one
SELECT
  id,
  patient,
  operation_date,
  operation,
  user,
  created_at,
  updated_at
FROM previous_operations
WHERE id = sqlc.arg(procedure_id)
  AND deleted_at IS NULL;
