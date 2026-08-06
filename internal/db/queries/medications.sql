-- List medications: active meds for a patient
-- name: ListMedications :many
SELECT * FROM medications
WHERE patient = sqlc.arg(patient_id)
  AND active = 'active'
  AND deleted_at IS NULL
ORDER BY start_date DESC;

-- List all medications (including inactive) for a patient
-- name: ListAllMedications :many
SELECT * FROM medications
WHERE patient = sqlc.arg(patient_id)
  AND deleted_at IS NULL
ORDER BY start_date DESC;

-- Get medication by id
-- name: GetMedicationById :one
SELECT * FROM medications
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- Create medication
-- name: CreateMedication :execresult
INSERT INTO medications (
  patient, drug_name, dosage, frequency, start_date,
  end_date, prescribing_provider, active
) VALUES (
  sqlc.arg(patient), sqlc.arg(drug_name), sqlc.arg(dosage),
  sqlc.arg(frequency), sqlc.arg(start_date), sqlc.arg(end_date),
  sqlc.arg(prescribing_provider), sqlc.arg(active)
);

-- Update medication
-- name: UpdateMedication :exec
UPDATE medications SET
  drug_name = sqlc.arg(drug_name),
  dosage = sqlc.arg(dosage),
  frequency = sqlc.arg(frequency),
  start_date = sqlc.arg(start_date),
  end_date = sqlc.arg(end_date),
  prescribing_provider = sqlc.arg(prescribing_provider)
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- Discontinue medication (set inactive and end_date)
-- name: DiscontinueMedication :exec
UPDATE medications SET
  active = 'inactive',
  end_date = sqlc.arg(end_date)
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;
