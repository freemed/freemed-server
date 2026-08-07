-- name: InsertPortalAuditLog :execresult
INSERT INTO portal_audit_log (
  created_at, updated_at, patient_id,
  action, ip_address, user_agent, success
) VALUES (
  NOW(), NOW(), sqlc.arg(patient_id),
  sqlc.arg(action), sqlc.arg(ip_address), sqlc.arg(user_agent),
  sqlc.arg(success)
);

-- name: ListPortalAuditByPatient :many
SELECT * FROM portal_audit_log
WHERE patient_id = sqlc.arg(patient_id)
  AND deleted_at IS NULL
ORDER BY created_at DESC;
