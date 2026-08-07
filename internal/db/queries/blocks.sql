-- name: ListBlockedSlots :many
SELECT * FROM schedulerblockslots
WHERE sbsdate = sqlc.arg(sbsdate) AND sbsprovider = sqlc.arg(sbsprovider)
ORDER BY sbshour, sbsminute;

-- name: CreateBlockedSlot :execresult
INSERT INTO schedulerblockslots (
  sbsdate, sbshour, sbsminute, sbsduration, sbsprovider,
  sbsreason, user,
  created_at, updated_at
) VALUES (
  sqlc.arg(sbsdate), sqlc.arg(sbshour), sqlc.arg(sbsminute),
  sqlc.arg(sbsduration), sqlc.arg(sbsprovider),
  sqlc.arg(sbsreason), sqlc.arg(user),
  NOW(), NOW()
);

-- name: DeleteBlockedSlot :exec
DELETE FROM schedulerblockslots WHERE id = sqlc.arg(id);
