-- name: ListSmsProviders :many
SELECT * FROM smsprovider ORDER BY name;

-- name: GetSmsProvider :one
SELECT * FROM smsprovider WHERE id = sqlc.arg(id);
