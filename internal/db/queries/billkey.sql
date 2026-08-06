-- name: GetBillkeyById :one
SELECT * FROM billkey WHERE id = sqlc.arg(id);
