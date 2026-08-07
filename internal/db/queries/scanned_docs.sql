-- name: ListScannedDocs :many
SELECT * FROM scanned_docs
WHERE patient = sqlc.arg(patient_id)
  AND active = 'active'
  AND deleted_at IS NULL
ORDER BY document_date DESC;

-- name: GetScannedDoc :one
SELECT * FROM scanned_docs
WHERE id = sqlc.arg(id)
  AND active = 'active'
  AND deleted_at IS NULL;
