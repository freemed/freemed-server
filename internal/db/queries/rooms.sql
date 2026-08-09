-- name: ListRooms :many
SELECT * FROM room
WHERE deleted_at IS NULL
ORDER BY name;

-- name: ListActiveRooms :many
SELECT * FROM room
WHERE deleted_at IS NULL
  AND active = 'active'
ORDER BY name;

-- name: GetRoom :one
SELECT * FROM room
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL
LIMIT 1;

-- name: CreateRoom :execresult
INSERT INTO room (
  created_at, updated_at, name, facility_id, room_type
) VALUES (
  NOW(), NOW(), sqlc.arg(name), sqlc.arg(facility_id), sqlc.arg(room_type)
);

-- name: UpdateRoom :exec
UPDATE room
SET updated_at = NOW(),
    name = sqlc.arg(name),
    facility_id = sqlc.arg(facility_id),
    room_type = sqlc.arg(room_type)
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: DeactivateRoom :exec
UPDATE room
SET updated_at = NOW(),
    deleted_at = NOW(),
    active = 'inactive'
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;
