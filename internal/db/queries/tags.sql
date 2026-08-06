-- Tags: list active tags for a patient (not expired)
-- name: ListTags :many
SELECT
  id,
  tag,
  patient,
  user,
  datecreate,
  dateexpire,
  created_at,
  updated_at,
  deleted_at
FROM patienttag
WHERE patient = sqlc.arg(patient_id)
  AND (dateexpire IS NULL OR dateexpire > NOW())
ORDER BY datecreate DESC;

-- Tags: create a new patient tag
-- name: CreateTag :execresult
INSERT INTO patienttag (
  tag, patient, user, datecreate, created_at, updated_at
) VALUES (
  sqlc.arg(tag), sqlc.arg(patient_id), sqlc.arg(user_id), NOW(), NOW(), NOW()
);

-- Tags: expire a tag (soft-delete by setting dateexpire)
-- name: ExpireTag :exec
UPDATE patienttag
SET dateexpire = NOW(), updated_at = NOW()
WHERE id = sqlc.arg(tag_id);

-- Tags: search patients by tag text
-- name: SearchByTag :many
SELECT DISTINCT
  p.id AS patient_id,
  p.ptlname AS last_name,
  p.ptfname AS first_name,
  p.ptid AS patient_id_ext,
  t.tag
FROM patienttag t
JOIN patient p ON t.patient = p.id
WHERE t.tag LIKE CONCAT('%', sqlc.arg(query), '%')
  AND (t.dateexpire IS NULL OR t.dateexpire > NOW());
