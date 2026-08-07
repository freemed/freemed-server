-- name: ListPreviousOperationsByPatient :many
SELECT * FROM previous_operations
WHERE patient = sqlc.arg(patient_id)
  AND deleted_at IS NULL
ORDER BY operation_date DESC;

-- name: CreatePreviousOperation :execresult
INSERT INTO previous_operations (
  created_at, updated_at, patient, operation_date, operation, user
) VALUES (
  NOW(), NOW(), sqlc.arg(patient_id), sqlc.arg(operation_date),
  sqlc.arg(operation), sqlc.arg(user_id)
);
