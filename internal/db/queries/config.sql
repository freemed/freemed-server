-- name: ListAllConfig :many
SELECT * FROM config ORDER BY c_option;

-- name: GetConfigById :one
SELECT * FROM config WHERE id = sqlc.arg(id);
