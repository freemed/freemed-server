-- name: ListTools :many
SELECT * FROM tools
WHERE active = 'active'
  AND deleted_at IS NULL
ORDER BY tool_name;

-- name: GetTool :one
SELECT * FROM tools
WHERE id = sqlc.arg(id)
  AND active = 'active'
  AND deleted_at IS NULL;
