-- CallIn: list all call-in records
-- name: ListCallIns :many
SELECT
  id,
  cilname,
  cifname,
  cicomplaint,
  created_at,
  updated_at,
  deleted_at
FROM callin
ORDER BY id DESC
LIMIT 50;

-- CallIn: get a single call-in record by ID
-- name: GetCallIn :one
SELECT
  id,
  cilname,
  cifname,
  cicomplaint,
  created_at,
  updated_at,
  deleted_at
FROM callin
WHERE id = ?;
