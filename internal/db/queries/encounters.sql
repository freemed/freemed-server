-- Encounter list: all encounters for a patient
-- name: ListEncounters :many
SELECT *
FROM patient_emr
WHERE patient = sqlc.arg('patient_id')
  AND module = 'encounters'
ORDER BY stamp DESC;

-- Encounter detail: single encounter by patient and encounter ID
-- name: GetEncounter :one
SELECT *
FROM patient_emr
WHERE patient = sqlc.arg('patient_id')
  AND module = 'encounters'
  AND id = sqlc.arg('encounter_id');
