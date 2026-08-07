-- name: ListLabs :many
SELECT * FROM labs
WHERE patient = sqlc.arg(patient_id)
  AND active = 'active'
  AND deleted_at IS NULL
ORDER BY lab_date DESC;

-- name: CreateLab :execresult
INSERT INTO labs (
  patient, lab_name, lab_date, result, unit,
  reference_range, status, notes, user, active,
  created_at, updated_at
) VALUES (
  sqlc.arg(patient_id),
  sqlc.arg(lab_name),
  sqlc.arg(lab_date),
  sqlc.arg(result),
  sqlc.arg(unit),
  sqlc.arg(reference_range),
  sqlc.arg(status),
  sqlc.arg(notes),
  sqlc.arg(user_id),
  'active',
  NOW(),
  NOW()
);
