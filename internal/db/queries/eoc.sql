-- name: ListEOCs :many
SELECT *
FROM episode_of_care
WHERE patient = sqlc.arg('patient_id')
  AND active = 'active'
ORDER BY start_date DESC;

-- name: CreateEOC :execresult
INSERT INTO episode_of_care (
  patient, start_date, end_date, description,
  status, provider, notes, user,
  created_at, updated_at, active
) VALUES (
  sqlc.arg('patient_id'), sqlc.arg('start_date'), sqlc.arg('end_date'),
  sqlc.arg('description'), sqlc.arg('status'), sqlc.arg('provider_id'),
  sqlc.arg('notes'), sqlc.arg('user_id'),
  NOW(), NOW(), 'active'
);
