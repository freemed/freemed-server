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
