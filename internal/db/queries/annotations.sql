-- List annotations for a patient, ordered by created_at descending
-- name: ListAnnotations :many
SELECT * FROM annotations
WHERE apatient = sqlc.arg(patient_id)
ORDER BY created_at DESC;

-- Create a new annotation entry
-- name: CreateAnnotation :execresult
INSERT INTO annotations (
  atimestamp, apatient, amodule, atable, aid, auser, annotation,
  created_at, updated_at
) VALUES (
  sqlc.arg(atimestamp), sqlc.arg(patient_id), sqlc.arg(amodule),
  sqlc.arg(atable), sqlc.arg(aid), sqlc.arg(user_id),
  sqlc.arg(annotation), NOW(), NOW()
);
