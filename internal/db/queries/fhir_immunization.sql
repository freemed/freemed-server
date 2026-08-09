-- FHIR Immunization: list immunizations for a patient
-- name: FhirImmunizationsByPatient :many
SELECT
  id,
  patient,
  dateof,
  provider,
  admin_provider,
  eoc,
  immunization,
  route,
  body_site,
  manufacturer,
  lot_number,
  previous_doses,
  recovered,
  notes,
  orderid,
  locked,
  user,
  active,
  created_at,
  updated_at
FROM immunization
WHERE patient = sqlc.arg(patient_id)
  AND active = 'active'
  AND deleted_at IS NULL
ORDER BY dateof DESC;

-- FHIR Immunization: lookup a single immunization by ID
-- name: FhirImmunizationById :one
SELECT
  id,
  patient,
  dateof,
  provider,
  admin_provider,
  eoc,
  immunization,
  route,
  body_site,
  manufacturer,
  lot_number,
  previous_doses,
  recovered,
  notes,
  orderid,
  locked,
  user,
  active,
  created_at,
  updated_at
FROM immunization
WHERE id = sqlc.arg(immunization_id)
  AND deleted_at IS NULL;
