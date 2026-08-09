-- name: AcquireRecordLock :execresult
INSERT INTO record_lock (
  created_at, updated_at, table_name, record_id, user_id, locked_at, expires_at
) VALUES (
  NOW(), NOW(), sqlc.arg(table_name), sqlc.arg(record_id),
  sqlc.arg(user_id), NOW(), DATE_ADD(NOW(), INTERVAL 5 MINUTE)
);

-- name: CheckRecordLock :one
SELECT id, table_name, record_id, user_id, locked_at, expires_at
FROM record_lock
WHERE table_name = sqlc.arg(table_name)
  AND record_id = sqlc.arg(record_id)
  AND expires_at > NOW()
  AND deleted_at IS NULL
LIMIT 1;

-- name: ReleaseRecordLock :exec
DELETE FROM record_lock
WHERE table_name = sqlc.arg(table_name)
  AND record_id = sqlc.arg(record_id)
  AND user_id = sqlc.arg(user_id);

-- name: ExpireRecordLocks :exec
UPDATE record_lock
SET deleted_at = NOW(), updated_at = NOW()
WHERE expires_at <= NOW()
  AND deleted_at IS NULL;
