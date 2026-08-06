-- name: GetSchedulerById :one
SELECT * FROM scheduler WHERE id = sqlc.arg(id);

-- name: UpdateScheduler :exec
UPDATE scheduler SET
  caldateof = sqlc.arg(date_of),
  calhour = sqlc.arg(hour),
  calminute = sqlc.arg(minute),
  calduration = sqlc.arg(duration),
  calmodified = sqlc.arg(modified)
WHERE id = sqlc.arg(id);
