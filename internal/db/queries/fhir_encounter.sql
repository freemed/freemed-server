-- FHIR Encounter: list encounters (procrec) for a patient
-- name: FhirEncountersByPatient :many
SELECT
  id,
  procpatient AS patient,
  proceoc AS eoc,
  proccpt AS cpt,
  proccptmod AS cpt_mod,
  procphysician AS physician,
  procdt AS encounter_date,
  procdtend AS encounter_end_date,
  procpos AS place_of_service,
  proccomment AS comment,
  created_at,
  updated_at
FROM procrec
WHERE procpatient = sqlc.arg(patient_id)
  AND deleted_at IS NULL
ORDER BY procdt DESC;

-- FHIR Encounter: lookup a single encounter by ID
-- name: FhirEncounterById :one
SELECT
  id,
  procpatient AS patient,
  proceoc AS eoc,
  proccpt AS cpt,
  proccptmod AS cpt_mod,
  procphysician AS physician,
  procdt AS encounter_date,
  procdtend AS encounter_end_date,
  procpos AS place_of_service,
  proccomment AS comment,
  created_at,
  updated_at
FROM procrec
WHERE id = sqlc.arg(encounter_id)
  AND deleted_at IS NULL;
