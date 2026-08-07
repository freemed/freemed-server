-- FHIR Observation: list vitals by patient ID
-- name: FhirVitalsByPatient :many
SELECT
  id,
  patient,
  date_taken,
  systolic,
  diastolic,
  heart_rate,
  respiratory_rate,
  temperature,
  oxygen_saturation,
  height_cm,
  weight_kg,
  bmi,
  notes,
  created_at,
  updated_at
FROM vitals
WHERE patient = sqlc.arg(patient_id)
ORDER BY date_taken DESC;

-- FHIR Observation: list all vitals (optionally filtered by patient)
-- name: FhirVitalsAll :many
SELECT
  id,
  patient,
  date_taken,
  systolic,
  diastolic,
  heart_rate,
  respiratory_rate,
  temperature,
  oxygen_saturation,
  height_cm,
  weight_kg,
  bmi,
  notes,
  created_at,
  updated_at
FROM vitals
WHERE (sqlc.narg(patient_id) IS NULL OR patient = sqlc.narg(patient_id))
ORDER BY date_taken DESC
LIMIT 100;
