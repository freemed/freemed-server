-- name: InsertAuditLog :execresult
INSERT INTO audit_log (
  created_at, updated_at, action, user_id, patient_id,
  resource_type, resource_id, ip_address, user_agent, details, success
) VALUES (
  NOW(), NOW(), sqlc.arg(action), sqlc.arg(user_id), sqlc.arg(patient_id),
  sqlc.arg(resource_type), sqlc.arg(resource_id), sqlc.arg(ip_address),
  sqlc.arg(user_agent), sqlc.arg(details), sqlc.arg(success)
);

-- name: ListAuditLog :many
SELECT * FROM audit_log
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListAuditLogByUser :many
SELECT * FROM audit_log
WHERE user_id = sqlc.arg(user_id)
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListAuditLogByPatient :many
SELECT * FROM audit_log
WHERE patient_id = sqlc.arg(patient_id)
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT ? OFFSET ?;
