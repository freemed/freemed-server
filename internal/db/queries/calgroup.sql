-- name: ListCalGroups :many
SELECT * FROM calgroup
ORDER BY created_at DESC;

-- name: ListCalGroupsPaginated :many
SELECT * FROM calgroup
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountCalGroups :one
SELECT COUNT(*) AS total FROM calgroup;

-- name: GetCalGroup :one
SELECT * FROM calgroup
WHERE id = sqlc.arg(id);
