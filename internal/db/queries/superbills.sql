-- name: ListSuperbills :many
SELECT * FROM superbill ORDER BY date_created DESC;

-- name: CreateSuperbill :execresult
INSERT INTO superbill (
  patient, date_from, date_to, provider, status, total_charges,
  date_created, user, active, created_at, updated_at
) VALUES (
  sqlc.arg(patient), sqlc.arg(date_from), sqlc.arg(date_to),
  sqlc.arg(provider), sqlc.arg(status), sqlc.arg(total_charges),
  sqlc.arg(date_created), sqlc.arg(user), 'active', NOW(), NOW()
);

-- name: GetSuperbill :one
SELECT * FROM superbill WHERE id = sqlc.arg(id);
