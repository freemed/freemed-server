-- name: ListTemplates :many
SELECT * FROM appttemplate ORDER BY atname;

-- name: GetTemplate :one
SELECT * FROM appttemplate WHERE id = sqlc.arg(id);

-- name: CreateTemplate :execresult
INSERT INTO appttemplate (
  atname, atduration, atequipment, atcolor,
  created_at, updated_at
) VALUES (
  sqlc.arg(atname), sqlc.arg(atduration), sqlc.arg(atequipment), sqlc.arg(atcolor),
  NOW(), NOW()
);

-- name: UpdateTemplate :exec
UPDATE appttemplate SET
  atname = sqlc.arg(atname),
  atduration = sqlc.arg(atduration),
  atequipment = sqlc.arg(atequipment),
  atcolor = sqlc.arg(atcolor),
  updated_at = NOW()
WHERE id = sqlc.arg(id);

-- name: DeleteTemplate :exec
DELETE FROM appttemplate WHERE id = sqlc.arg(id);
