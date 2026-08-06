-- name: GetDataStoreByPatientModule :one
SELECT * FROM pds
WHERE patient = sqlc.arg(patient_id)
  AND module = LOWER(sqlc.arg(module))
  AND id = sqlc.arg(id);
