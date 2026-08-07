-- Patient picklist: search by first name
-- name: PatientPicklistByFirstName :many
SELECT
  CONCAT(ptlname, ', ', ptfname, ' (', ptid, ')') AS value,
  id
FROM patient
WHERE ptfname LIKE CONCAT(sqlc.arg('query'), '%')
  AND (ptarchive IS NULL OR ptarchive = 0)
LIMIT 20;

-- Patient picklist: search by last name
-- name: PatientPicklistByLastName :many
SELECT
  CONCAT(ptlname, ', ', ptfname, ' (', ptid, ')') AS value,
  id
FROM patient
WHERE ptlname LIKE CONCAT(sqlc.arg('query'), '%')
  AND (ptarchive IS NULL OR ptarchive = 0)
LIMIT 20;

-- Patient picklist: search by first AND last name
-- name: PatientPicklistByName :many
SELECT
  CONCAT(ptlname, ', ', ptfname, ' (', ptid, ')') AS value,
  id
FROM patient
WHERE ptlname LIKE CONCAT(sqlc.arg('last_name'), '%')
  AND ptfname LIKE CONCAT(sqlc.arg('first_name'), '%')
  AND (ptarchive IS NULL OR ptarchive = 0)
LIMIT 20;

-- Patient picklist: search by patient ID
-- name: PatientPicklistByPatientId :many
SELECT
  CONCAT(ptlname, ', ', ptfname, ' (', ptid, ')') AS value,
  id
FROM patient
WHERE ptid LIKE CONCAT(sqlc.arg('query'), '%')
  AND (ptarchive IS NULL OR ptarchive = 0)
LIMIT 20;

-- Patient picklist: search by either first/last name (either mode)
-- name: PatientPicklistByEither :many
SELECT
  CONCAT(ptlname, ', ', ptfname, ' (', ptid, ')') AS value,
  id
FROM patient
WHERE (
  ptfname LIKE CONCAT(sqlc.arg('query'), '%')
  OR ptlname LIKE CONCAT(sqlc.arg('query'), '%')
)
  AND (ptarchive IS NULL OR ptarchive = 0)
LIMIT 20;

-- Patient picklist: search by first name OR patient ID
-- name: PatientPicklistByFirstNameOrId :many
SELECT
  CONCAT(ptlname, ', ', ptfname, ' (', ptid, ')') AS value,
  id
FROM patient
WHERE (
  ptfname LIKE CONCAT(sqlc.arg('query'), '%')
  OR ptid LIKE CONCAT(sqlc.arg('query'), '%')
)
  AND (ptarchive IS NULL OR ptarchive = 0)
LIMIT 20;

-- Patient picklist: search by last name OR patient ID
-- name: PatientPicklistByLastNameOrId :many
SELECT
  CONCAT(ptlname, ', ', ptfname, ' (', ptid, ')') AS value,
  id
FROM patient
WHERE (
  ptlname LIKE CONCAT(sqlc.arg('query'), '%')
  OR ptid LIKE CONCAT(sqlc.arg('query'), '%')
)
  AND (ptarchive IS NULL OR ptarchive = 0)
LIMIT 20;

-- Patient search with optional filters
-- name: PatientSearch :many
SELECT
  p.ptlname AS last_name,
  p.ptfname AS first_name,
  p.ptmname AS middle_name,
  p.ptid AS patient_id,
  FLOOR((TO_DAYS(NOW()) - TO_DAYS(p.ptdob)) / 365) AS age,
  p.ptdob AS date_of_birth,
  p.id AS id
FROM patient p
LEFT OUTER JOIN patient_address pa ON p.id = pa.patient
WHERE pa.active = 1
  AND (sqlc.narg('first_name') IS NULL OR p.ptfname LIKE CONCAT('%', sqlc.narg('first_name'), '%'))
  AND (sqlc.narg('last_name') IS NULL OR p.ptlname LIKE CONCAT('%', sqlc.narg('last_name'), '%'))
  AND (sqlc.narg('patient_id') IS NULL OR p.ptid LIKE CONCAT('%', sqlc.narg('patient_id'), '%'))
  AND (sqlc.narg('city') IS NULL OR pa.city LIKE CONCAT('%', sqlc.narg('city'), '%'))
  AND (sqlc.narg('zip') IS NULL OR pa.zip LIKE CONCAT('%', sqlc.narg('zip'), '%'))
  AND (sqlc.narg('dmv') IS NULL OR p.dmv LIKE CONCAT('%', sqlc.narg('dmv'), '%'))
  AND (sqlc.narg('email') IS NULL OR p.pemail LIKE CONCAT('%', sqlc.narg('email'), '%'))
  AND (sqlc.narg('ssn') IS NULL OR p.ssn LIKE CONCAT('%', sqlc.narg('ssn'), '%'))
  AND p.ptarchive = 0
ORDER BY p.ptlname, p.ptfname, p.ptmname
LIMIT 20;

-- Patient total count (non-archived)
-- name: PatientTotalInSystem :one
SELECT COUNT(*) AS total FROM patient WHERE ptarchive = 0;

-- Patient duplicate search (with optional middle name, suffix, DOB)
-- name: PatientSearchDuplicates :many
SELECT ptid FROM patient p
WHERE ptlname = sqlc.arg(ptlname)
  AND ptfname = sqlc.arg(ptfname)
  AND (sqlc.narg('ptmname') IS NULL OR ptmname = sqlc.narg('ptmname'))
  AND (sqlc.narg('ptsuffix') IS NULL OR ptsuffix = sqlc.narg('ptsuffix'))
  AND (sqlc.narg('ptdob') IS NULL OR ptdob = sqlc.narg('ptdob'))
  AND ptarchive = 0;

-- Patient list with pagination
-- name: ListPatients :many
SELECT
  id,
  ptlname AS last_name,
  ptfname AS first_name,
  ptmname AS middle_name,
  ptsuffix AS suffix,
  ptsex AS gender,
  ptid AS patient_id,
  ptdob AS date_of_birth,
  ptarchive AS archived,
  created_at,
  updated_at
FROM patient
WHERE ptarchive = 0
ORDER BY ptlname, ptfname, ptmname
LIMIT ? OFFSET ?;

-- Patient count (non-archived)
-- name: CountPatients :one
SELECT COUNT(*) AS total FROM patient WHERE ptarchive = 0;

-- Patient create: insert a new patient record
-- name: PatientCreate :execresult
INSERT INTO patient (
  ptlname, ptfname, ptmname, ptsuffix, ptsex, ptid, ptdob,
  ptarchive, ptbilltype, user, stamp
) VALUES (
  sqlc.arg(ptlname), sqlc.arg(ptfname), sqlc.arg(ptmname), sqlc.arg(ptsuffix),
  sqlc.arg(ptsex), sqlc.arg(ptid), sqlc.arg(ptdob),
  0, '', sqlc.arg(user), NOW()
);
