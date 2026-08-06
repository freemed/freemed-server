-- name: ListAllConfig :many
SELECT * FROM config ORDER BY c_option;

-- name: GetConfigById :one
SELECT * FROM config WHERE id = sqlc.arg(id);

-- name: UpdateConfig :exec
UPDATE config SET c_value = sqlc.arg(c_value) WHERE id = sqlc.arg(id);

-- name: GetConfigSections :many
SELECT DISTINCT c_section FROM config ORDER BY c_section;
