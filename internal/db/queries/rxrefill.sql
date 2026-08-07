-- RxRefill: list all refill requests
-- name: ListRefillRequests :many
SELECT
  id,
  created_at,
  updated_at,
  deleted_at,
  user,
  patient,
  provider,
  rxorig,
  note,
  approved,
  locked
FROM rxrefillrequest
ORDER BY created_at DESC
LIMIT 50;
