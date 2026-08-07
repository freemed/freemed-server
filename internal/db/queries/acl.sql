-- Groups

-- name: ListGroups :many
SELECT id, groupname, groupdescrip, created_at, updated_at
FROM usergroup
ORDER BY groupname;

-- name: GetGroup :one
SELECT id, groupname, groupdescrip, created_at, updated_at
FROM usergroup
WHERE id = sqlc.arg(id);

-- name: CreateGroup :execresult
INSERT INTO usergroup (groupname, groupdescrip, created_at, updated_at)
VALUES (sqlc.arg(groupname), sqlc.arg(groupdescrip), NOW(), NOW());

-- name: UpdateGroup :exec
UPDATE usergroup
SET groupname = sqlc.arg(groupname), groupdescrip = sqlc.arg(groupdescrip), updated_at = NOW()
WHERE id = sqlc.arg(id);

-- name: DeleteGroup :exec
DELETE FROM usergroup WHERE id = sqlc.arg(id);

-- User-Group Assignments

-- name: ListUserGroups :many
SELECT ug.group_id, g.groupname, g.groupdescrip
FROM user_groups ug
JOIN usergroup g ON ug.group_id = g.id
WHERE ug.user_id = sqlc.arg(user_id)
ORDER BY g.groupname;

-- name: AddUserToGroup :exec
INSERT INTO user_groups (user_id, group_id, created_at, updated_at)
VALUES (sqlc.arg(user_id), sqlc.arg(group_id), NOW(), NOW());

-- name: RemoveUserFromGroup :exec
DELETE FROM user_groups
WHERE user_id = sqlc.arg(user_id) AND group_id = sqlc.arg(group_id);

-- Permissions

-- name: ListPermissions :many
SELECT id, permission_name, permission_desc, created_at, updated_at
FROM acl_permissions
ORDER BY permission_name;

-- name: GetPermission :one
SELECT id, permission_name, permission_desc, created_at, updated_at
FROM acl_permissions
WHERE id = sqlc.arg(id);

-- name: CreatePermission :execresult
INSERT INTO acl_permissions (permission_name, permission_desc, created_at, updated_at)
VALUES (sqlc.arg(permission_name), sqlc.arg(permission_desc), NOW(), NOW());

-- name: DeletePermission :exec
DELETE FROM acl_permissions WHERE id = sqlc.arg(id);

-- Group-Permission Assignments

-- name: GetGroupPermissions :many
SELECT p.id, p.permission_name, p.permission_desc
FROM group_permissions gp
JOIN acl_permissions p ON gp.permission_id = p.id
WHERE gp.group_id = sqlc.arg(group_id)
ORDER BY p.permission_name;

-- name: AddGroupPermission :exec
INSERT INTO group_permissions (group_id, permission_id, created_at, updated_at)
VALUES (sqlc.arg(group_id), sqlc.arg(permission_id), NOW(), NOW());

-- name: RemoveGroupPermission :exec
DELETE FROM group_permissions
WHERE group_id = sqlc.arg(group_id) AND permission_id = sqlc.arg(permission_id);

-- User Permission Check (for authorization)

-- name: GetUserPermissions :many
SELECT DISTINCT p.permission_name
FROM user_groups ug
JOIN group_permissions gp ON ug.group_id = gp.group_id
JOIN acl_permissions p ON gp.permission_id = p.id
WHERE ug.user_id = sqlc.arg(user_id)
ORDER BY p.permission_name;

-- name: UserHasPermission :one
SELECT COUNT(*) > 0 AS has_permission
FROM user_groups ug
JOIN group_permissions gp ON ug.group_id = gp.group_id
JOIN acl_permissions p ON gp.permission_id = p.id
WHERE ug.user_id = sqlc.arg(user_id) AND p.permission_name = sqlc.arg(permission_name);
