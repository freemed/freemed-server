-- name: ListSuperbillTemplates :many
SELECT * FROM superbill_template ORDER BY name;

-- name: GetSuperbillTemplate :one
SELECT * FROM superbill_template WHERE id = sqlc.arg(id);
