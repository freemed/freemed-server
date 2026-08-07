-- name: ListFormTemplates :many
SELECT * FROM form_templates ORDER BY name;

-- name: GetFormTemplate :one
SELECT * FROM form_templates WHERE id = sqlc.arg(id);

-- name: CreateFormTemplate :execresult
INSERT INTO form_templates (
  name, description, form_type, template_data, is_default, user,
  created_at, updated_at
) VALUES (
  sqlc.arg(name), sqlc.arg(description), sqlc.arg(form_type), sqlc.arg(template_data), sqlc.arg(is_default), sqlc.arg(user),
  NOW(), NOW()
);

-- name: UpdateFormTemplate :exec
UPDATE form_templates SET
  name = sqlc.arg(name),
  description = sqlc.arg(description),
  form_type = sqlc.arg(form_type),
  template_data = sqlc.arg(template_data),
  is_default = sqlc.arg(is_default),
  updated_at = NOW()
WHERE id = sqlc.arg(id);

-- name: DeleteFormTemplate :exec
DELETE FROM form_templates WHERE id = sqlc.arg(id);
