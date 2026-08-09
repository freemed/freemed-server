-- name: SearchRxNorm :many
SELECT * FROM rxnorm
WHERE deleted_at IS NULL
  AND drug_name LIKE CONCAT('%', sqlc.arg(query), '%')
ORDER BY drug_name
LIMIT 50;
