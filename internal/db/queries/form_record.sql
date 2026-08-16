-- name: ListFormRecords :many
SELECT * FROM form_record WHERE fr_id = sqlc.arg(form_id) AND deleted_at IS NULL ORDER BY id;

-- name: CreateFormRecord :execresult
INSERT INTO form_record (created_at, updated_at, fr_id, fr_uuid, fr_name, fr_value)
VALUES (NOW(), NOW(), sqlc.arg(fr_id), sqlc.arg(fr_uuid), sqlc.arg(fr_name), sqlc.arg(fr_value));

-- name: UpdateFormRecord :exec
UPDATE form_record SET updated_at = NOW(), fr_value = sqlc.arg(fr_value) WHERE fr_id = sqlc.arg(fr_id) AND fr_uuid = sqlc.arg(fr_uuid) AND deleted_at IS NULL;

-- name: DeleteFormRecords :exec
UPDATE form_record SET updated_at = NOW(), deleted_at = NOW() WHERE fr_id = sqlc.arg(fr_id) AND deleted_at IS NULL;
