-- List correspondence for a patient, ordered by date descending
-- name: ListCorrespondence :many
SELECT * FROM patient_correspondence
WHERE patient = sqlc.arg(patient_id)
  AND active = 'active'
ORDER BY date DESC;

-- Create a new correspondence entry
-- name: CreateCorrespondence :execresult
INSERT INTO patient_correspondence (
  patient, correspondence_type, direction, contact_name, contact_method,
  summary, date, user, active, created_at, updated_at
) VALUES (
  sqlc.arg(patient_id), sqlc.arg(correspondence_type), sqlc.arg(direction),
  sqlc.arg(contact_name), sqlc.arg(contact_method),
  sqlc.arg(summary), sqlc.arg(date),
  sqlc.arg(user_id), 'active', NOW(), NOW()
);
