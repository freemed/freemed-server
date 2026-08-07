-- name: ListChronicProblemsByPatient :many
SELECT * FROM chronic_problems
WHERE patient = sqlc.arg(patient_id)
  AND deleted_at IS NULL
ORDER BY date DESC;

-- name: CreateChronicProblem :execresult
INSERT INTO chronic_problems (
  created_at, updated_at, patient, date, problem, user
) VALUES (
  NOW(), NOW(), sqlc.arg(patient_id), sqlc.arg(date),
  sqlc.arg(problem), sqlc.arg(user_id)
);
