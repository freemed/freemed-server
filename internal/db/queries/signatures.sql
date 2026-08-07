-- name: ListSignaturesByPatient :many
SELECT * FROM signatures
WHERE patient = sqlc.arg(patient_id)
  AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetSignature :one
SELECT * FROM signatures
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: CreateSignature :execresult
INSERT INTO signatures (
  patient, module, module_field, oid,
  signature_data, format,
  collector_location, collector_model, collector_jobid,
  collector_finished, user
) VALUES (
  sqlc.arg(patient), sqlc.arg(module), sqlc.arg(module_field), sqlc.arg(oid),
  sqlc.narg(signature_data), sqlc.arg(format),
  sqlc.narg(collector_location), sqlc.narg(collector_model), sqlc.narg(collector_jobid),
  sqlc.arg(collector_finished), sqlc.arg(user)
);
