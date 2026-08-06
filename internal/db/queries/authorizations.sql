-- List authorizations for a patient with insurance company name
-- name: ListAuthorizations :many
SELECT a.*, i.insconame
FROM authorizations a
LEFT JOIN insco i ON a.authinsco = i.id
WHERE a.authpatient = sqlc.arg(patient_id)
  AND a.active = 'active'
ORDER BY a.authdtbegin DESC;

-- Get single authorization with insurance company name
-- name: GetAuthorization :one
SELECT a.*, i.insconame
FROM authorizations a
LEFT JOIN insco i ON a.authinsco = i.id
WHERE a.id = sqlc.arg(id);
