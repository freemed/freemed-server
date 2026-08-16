-- name: ListSchedulingRules :many
SELECT * FROM schedulingrules
WHERE deleted_at IS NULL
ORDER BY id DESC;

-- name: GetSchedulingRule :one
SELECT * FROM schedulingrules
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: CreateSchedulingRule :execresult
INSERT INTO schedulingrules (
  created_at, updated_at, user, provider, reason,
  dowbegin, dowend, datebegin, dateend, timebegin, timeend, newpatient
) VALUES (
  NOW(), NOW(), sqlc.arg(user_id), sqlc.narg(provider), sqlc.narg(reason),
  sqlc.narg(dowbegin), sqlc.narg(dowend), sqlc.narg(datebegin), sqlc.narg(dateend),
  sqlc.narg(timebegin), sqlc.narg(timeend), sqlc.narg(newpatient)
);

-- name: UpdateSchedulingRule :exec
UPDATE schedulingrules
SET
  updated_at = NOW(),
  provider = sqlc.narg(provider),
  reason = sqlc.narg(reason),
  dowbegin = sqlc.narg(dowbegin),
  dowend = sqlc.narg(dowend),
  datebegin = sqlc.narg(datebegin),
  dateend = sqlc.narg(dateend),
  timebegin = sqlc.narg(timebegin),
  timeend = sqlc.narg(timeend),
  newpatient = sqlc.narg(newpatient)
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: DeleteSchedulingRule :exec
UPDATE schedulingrules
SET updated_at = NOW(), deleted_at = NOW()
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;
