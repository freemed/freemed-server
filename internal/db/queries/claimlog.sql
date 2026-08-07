-- List claim log entries for a specific claim (clprocedure)
-- name: ListClaimLog :many
SELECT
  id,
  cltimestamp,
  cluser,
  clprocedure,
  clpayrec,
  claction,
  clcomment,
  clformat,
  cltarget,
  cltargetopt,
  clbillkey,
  created_at,
  updated_at,
  deleted_at
FROM claimlog
WHERE clprocedure = sqlc.arg(claim_id)
ORDER BY cltimestamp DESC;

-- Insert a claim log entry
-- name: InsertClaimLog :execresult
INSERT INTO claimlog (
  cltimestamp,
  cluser,
  clprocedure,
  clpayrec,
  claction,
  clcomment,
  clformat,
  cltarget,
  cltargetopt,
  clbillkey,
  created_at,
  updated_at
) VALUES (
  sqlc.arg(cltimestamp),
  sqlc.arg(cluser),
  sqlc.arg(clprocedure),
  sqlc.arg(clpayrec),
  sqlc.arg(claction),
  sqlc.arg(clcomment),
  sqlc.arg(clformat),
  sqlc.arg(cltarget),
  sqlc.arg(cltargetopt),
  sqlc.arg(clbillkey),
  NOW(),
  NOW()
);
