-- name: ListTemplates :many
SELECT * FROM appttemplate ORDER BY atname;

-- name: GetTemplate :one
SELECT * FROM appttemplate WHERE id = sqlc.arg(id);
