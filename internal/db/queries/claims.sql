-- List claim log entries for a specific patient
-- name: PatientClaims :many
SELECT
  cl.id,
  cl.cltimestamp,
  cl.cluser,
  cl.clprocedure,
  cl.clpayrec,
  cl.claction,
  cl.clcomment,
  cl.clformat,
  cl.cltarget,
  cl.cltargetopt,
  cl.clbillkey,
  cl.created_at,
  cl.updated_at,
  cl.deleted_at,
  pr.proccpt AS cpt_code,
  pr.procdiag1 AS diagnosis_code,
  pr.proccharges AS charges,
  u.username
FROM claimlog cl
LEFT JOIN procrec pr ON cl.clprocedure = pr.id
LEFT JOIN user u ON cl.cluser = u.id
WHERE pr.procpatient = sqlc.arg(patient_id)
ORDER BY cl.cltimestamp DESC;

-- List most recent claim log entries system-wide
-- name: RecentClaims :many
SELECT
  cl.id,
  cl.cltimestamp,
  cl.cluser,
  cl.clprocedure,
  cl.clpayrec,
  cl.claction,
  cl.clcomment,
  cl.clformat,
  cl.cltarget,
  cl.cltargetopt,
  cl.clbillkey,
  cl.created_at,
  cl.updated_at,
  cl.deleted_at,
  pr.proccpt AS cpt_code,
  u.username
FROM claimlog cl
LEFT JOIN procrec pr ON cl.clprocedure = pr.id
LEFT JOIN user u ON cl.cluser = u.id
ORDER BY cl.cltimestamp DESC
LIMIT 50;

-- List pending claim log entries (claction = 'pending')
-- name: PendingClaims :many
SELECT
  cl.id,
  cl.cltimestamp,
  cl.cluser,
  cl.clprocedure,
  cl.clpayrec,
  cl.claction,
  cl.clcomment,
  cl.clformat,
  cl.cltarget,
  cl.cltargetopt,
  cl.clbillkey,
  cl.created_at,
  cl.updated_at,
  cl.deleted_at,
  pr.proccpt AS cpt_code,
  u.username
FROM claimlog cl
LEFT JOIN procrec pr ON cl.clprocedure = pr.id
LEFT JOIN user u ON cl.cluser = u.id
WHERE cl.claction = 'pending'
ORDER BY cl.cltimestamp DESC;

-- Update claim status (claction field)
-- name: UpdateClaimStatus :exec
UPDATE claimlog
SET claction = sqlc.arg(status),
    updated_at = NOW()
WHERE id = sqlc.arg(id);
