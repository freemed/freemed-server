-- name: ListEvents :many
SELECT * FROM events ORDER BY event_date DESC;

-- name: GetEvent :one
SELECT * FROM events WHERE id = sqlc.arg(id) LIMIT 1;
