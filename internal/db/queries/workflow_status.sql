-- name: ListWorkflowStatus :many
SELECT ws.*, wst.status_name, wst.status_order, wst.status_module
FROM workflow_status ws
LEFT JOIN workflow_status_type wst ON ws.status_type = wst.id
WHERE ws.patient = sqlc.arg(patient_id)
  AND ws.deleted_at IS NULL
ORDER BY ws.stamp DESC;

-- name: ListWorkflowStatusForDate :many
SELECT ws.*, wst.status_name, wst.status_order, wst.status_module
FROM workflow_status ws
LEFT JOIN workflow_status_type wst ON ws.status_type = wst.id
WHERE ws.patient = sqlc.arg(patient_id)
  AND DATE(ws.stamp) = sqlc.arg(status_date)
  AND ws.deleted_at IS NULL
ORDER BY ws.stamp DESC;

-- name: SetWorkflowStatus :execresult
INSERT INTO workflow_status (
  patient, user, status_type, status_completed,
  stamp, created_at, updated_at
) VALUES (
  sqlc.arg(patient_id),
  sqlc.arg(user_id),
  sqlc.arg(status_type),
  sqlc.arg(status_completed),
  sqlc.arg(stamp),
  NOW(),
  NOW()
);
