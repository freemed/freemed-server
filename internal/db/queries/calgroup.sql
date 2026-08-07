-- name: ListCalGroups :many
SELECT * FROM calgroup
ORDER BY created_at DESC;

-- name: GetCalGroup :one
SELECT * FROM calgroup
WHERE id = sqlc.arg(id);
