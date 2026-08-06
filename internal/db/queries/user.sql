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
