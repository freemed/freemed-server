-- name: ListFormResultsByPatient :many
SELECT * FROM form_results WHERE fr_patient = sqlc.arg(patient_id) AND active = 'active' AND deleted_at IS NULL ORDER BY fr_timestamp DESC;

-- name: GetFormResult :one
SELECT * FROM form_results WHERE id = sqlc.arg(id) AND deleted_at IS NULL;

-- name: CreateFormResult :execresult
INSERT INTO form_results (created_at, updated_at, fr_patient, fr_timestamp, fr_template, fr_formname, user, active)
VALUES (NOW(), NOW(), sqlc.arg(patient_id), NOW(), sqlc.arg(fr_template), sqlc.arg(fr_formname), sqlc.arg(user), sqlc.arg(active));

-- name: UpdateFormResult :exec
UPDATE form_results SET updated_at = NOW(), fr_timestamp = NOW(), fr_template = sqlc.arg(fr_template), fr_formname = sqlc.arg(fr_formname), active = sqlc.arg(active) WHERE id = sqlc.arg(id) AND deleted_at IS NULL;

-- name: DeleteFormResult :exec
UPDATE form_results SET updated_at = NOW(), deleted_at = NOW() WHERE id = sqlc.arg(id) AND deleted_at IS NULL;
