-- name: ListCurrentProblemsByPatient :many
SELECT * FROM current_problems
WHERE patient = sqlc.arg(patient_id)
  AND active = 'active'
  AND deleted_at IS NULL
ORDER BY date DESC;

-- name: CreateCurrentProblem :execresult
INSERT INTO current_problems (
  created_at, updated_at, patient, date, problem, user, active
) VALUES (
  NOW(), NOW(), sqlc.arg(patient_id), sqlc.arg(date),
  sqlc.arg(problem), sqlc.arg(user_id), 'active'
);
