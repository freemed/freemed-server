-- Global search: patients by name or ID
-- name: SearchPatients :many
SELECT
  id,
  ptlname,
  ptfname,
  ptid,
  'patient' AS result_type
FROM patient
WHERE (
  ptlname LIKE CONCAT('%', sqlc.arg('query'), '%')
  OR ptfname LIKE CONCAT('%', sqlc.arg('query'), '%')
  OR ptid LIKE CONCAT('%', sqlc.arg('query'), '%')
)
  AND ptarchive = 0
LIMIT 10;

-- Global search: messages by subject
-- name: SearchMessages :many
SELECT
  id,
  msgsubject AS title,
  'message' AS result_type
FROM messages
WHERE msgsubject LIKE CONCAT('%', sqlc.arg('query'), '%')
LIMIT 5;

-- Global search: scheduler appointments by patient name (future only)
-- name: SearchAppointments :many
SELECT
  s.id,
  CONCAT('Appt: ', p.ptlname) AS title,
  'appointment' AS result_type
FROM scheduler s
LEFT JOIN patient p ON s.calpatient = p.id
WHERE s.caldateof >= CURDATE()
  AND p.ptlname LIKE CONCAT('%', sqlc.arg('query'), '%')
LIMIT 5;
