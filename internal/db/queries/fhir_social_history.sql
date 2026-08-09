-- FHIR Observation: social history observations for a patient
-- name: FhirSocialHistoryObservations :many
SELECT
  id,
  patient,
  smoking_status,
  smoking_detail,
  alcohol_use,
  alcohol_detail,
  drug_use,
  drug_detail,
  exercise_frequency,
  occupation,
  living_situation,
  food_insecurity,
  transportation_access,
  notes,
  recorded_date,
  user,
  active,
  created_at,
  updated_at
FROM social_history
WHERE patient = sqlc.arg(patient_id)
  AND active = 'active'
  AND deleted_at IS NULL
ORDER BY recorded_date DESC;
