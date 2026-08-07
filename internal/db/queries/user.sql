-- name: GetUserCount :one
SELECT COUNT(*) FROM user;

-- name: GetUserByUsername :one
SELECT * FROM user WHERE username = sqlc.arg(username);

-- name: GetUserById :one
SELECT * FROM user WHERE id = sqlc.arg(id);

-- name: CheckUserPassword :one
SELECT * FROM user
WHERE username = sqlc.arg(username)
  AND userpassword = sqlc.arg(password);

-- name: CheckDuplicateUsername :many
SELECT id FROM user WHERE username = sqlc.arg(username);

-- name: ListUsers :many
SELECT id, username, userfname, userlname, userdescrip, usertype
FROM user
ORDER BY userlname, userfname;

-- name: CreateUser :execresult
INSERT INTO user (username, userpassword, userfname, userlname, userdescrip, usertype)
VALUES (
  sqlc.arg(username), sqlc.arg(userpassword), sqlc.arg(userfname),
  sqlc.arg(userlname), sqlc.arg(userdescrip), sqlc.arg(usertype)
);

-- name: UpdateUser :exec
UPDATE user SET
  username = sqlc.arg(username),
  userfname = sqlc.arg(userfname),
  userlname = sqlc.arg(userlname),
  userdescrip = sqlc.arg(userdescrip),
  usertype = sqlc.arg(usertype)
WHERE id = sqlc.arg(id);

-- name: UpdateUserPassword :exec
UPDATE user SET userpassword = sqlc.arg(userpassword)
WHERE id = sqlc.arg(id);

-- name: DeleteUser :exec
DELETE FROM user WHERE id = sqlc.arg(id);
