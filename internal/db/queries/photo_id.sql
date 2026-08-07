-- name: GetPhotoID :many
SELECT * FROM photoid
WHERE patient = sqlc.arg(patient_id)
  AND active = 'active'
  AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CreatePhotoID :execresult
INSERT INTO photoid (
  patient, photo, photo_mime, page_count, description, user
) VALUES (
  sqlc.arg(patient), sqlc.arg(photo), sqlc.arg(photo_mime),
  sqlc.arg(page_count), sqlc.arg(description), sqlc.arg(user)
);
