-- name: FindNextAvailable :many
SELECT calhour, calminute, calduration
FROM scheduler
WHERE calphysician = sqlc.arg(provider_id)
AND caldateof = sqlc.arg(req_date)
AND calstatus NOT IN ('cancelled', 'noshow')
ORDER BY calhour, calminute;
