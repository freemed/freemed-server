-- List all users (id and username) for message picker
-- name: MessagesListUsers :many
SELECT username, id FROM user;

-- View messages: all messages for a user (non-patient, all read states)
-- name: MessagesViewForUser :many
SELECT
  m.*,
  u.userdescrip AS sender
FROM messages m
LEFT OUTER JOIN user u ON u.id = m.msgby
WHERE (m.msgtag IS NULL OR LENGTH(m.msgtag) < 1)
  AND m.msgfor = sqlc.arg(user_id);

-- View messages: unread messages for a user (non-patient, unread only)
-- name: MessagesViewUnreadForUser :many
SELECT
  m.*,
  u.userdescrip AS sender
FROM messages m
LEFT OUTER JOIN user u ON u.id = m.msgby
WHERE (m.msgtag IS NULL OR LENGTH(m.msgtag) < 1)
  AND m.msgfor = sqlc.arg(user_id)
  AND m.msgread = 0;

-- View messages: all messages for a specific patient/user
-- name: MessagesViewForPatient :many
SELECT
  m.*,
  u.userdescrip AS sender
FROM messages m
LEFT OUTER JOIN user u ON u.id = m.msgby
WHERE (m.msgtag IS NULL OR LENGTH(m.msgtag) < 1)
  AND m.msgpatient = sqlc.arg(patient_id)
  AND m.msgfor = sqlc.arg(user_id);

-- View messages: unread messages for a specific patient/user
-- name: MessagesViewUnreadForPatient :many
SELECT
  m.*,
  u.userdescrip AS sender
FROM messages m
LEFT OUTER JOIN user u ON u.id = m.msgby
WHERE (m.msgtag IS NULL OR LENGTH(m.msgtag) < 1)
  AND m.msgpatient = sqlc.arg(patient_id)
  AND m.msgread = 0
  AND m.msgby = sqlc.arg(user_id);

-- Paginated messages: all messages for a user
-- name: MessagesViewForUserPaginated :many
SELECT
  m.*,
  u.userdescrip AS sender
FROM messages m
LEFT OUTER JOIN user u ON u.id = m.msgby
WHERE (m.msgtag IS NULL OR LENGTH(m.msgtag) < 1)
  AND m.msgfor = sqlc.arg(user_id)
LIMIT ? OFFSET ?;

-- Paginated messages: unread messages for a user
-- name: MessagesViewUnreadForUserPaginated :many
SELECT
  m.*,
  u.userdescrip AS sender
FROM messages m
LEFT OUTER JOIN user u ON u.id = m.msgby
WHERE (m.msgtag IS NULL OR LENGTH(m.msgtag) < 1)
  AND m.msgfor = sqlc.arg(user_id)
  AND m.msgread = 0
LIMIT ? OFFSET ?;

-- Paginated messages: all messages for a specific patient/user
-- name: MessagesViewForPatientPaginated :many
SELECT
  m.*,
  u.userdescrip AS sender
FROM messages m
LEFT OUTER JOIN user u ON u.id = m.msgby
WHERE (m.msgtag IS NULL OR LENGTH(m.msgtag) < 1)
  AND m.msgpatient = sqlc.arg(patient_id)
  AND m.msgfor = sqlc.arg(user_id)
LIMIT ? OFFSET ?;

-- Paginated messages: unread messages for a specific patient/user
-- name: MessagesViewUnreadForPatientPaginated :many
SELECT
  m.*,
  u.userdescrip AS sender
FROM messages m
LEFT OUTER JOIN user u ON u.id = m.msgby
WHERE (m.msgtag IS NULL OR LENGTH(m.msgtag) < 1)
  AND m.msgpatient = sqlc.arg(patient_id)
  AND m.msgread = 0
  AND m.msgby = sqlc.arg(user_id)
LIMIT ? OFFSET ?;

-- Count messages: all messages for a user
-- name: CountMessagesForUser :one
SELECT COUNT(*) AS total
FROM messages m
WHERE (m.msgtag IS NULL OR LENGTH(m.msgtag) < 1)
  AND m.msgfor = sqlc.arg(user_id);

-- Count messages: unread messages for a user
-- name: CountMessagesUnreadForUser :one
SELECT COUNT(*) AS total
FROM messages m
WHERE (m.msgtag IS NULL OR LENGTH(m.msgtag) < 1)
  AND m.msgfor = sqlc.arg(user_id)
  AND m.msgread = 0;

-- Count messages: all messages for a specific patient/user
-- name: CountMessagesForPatient :one
SELECT COUNT(*) AS total
FROM messages m
WHERE (m.msgtag IS NULL OR LENGTH(m.msgtag) < 1)
  AND m.msgpatient = sqlc.arg(patient_id)
  AND m.msgfor = sqlc.arg(user_id);

-- Count messages: unread messages for a specific patient/user
-- name: CountMessagesUnreadForPatient :one
SELECT COUNT(*) AS total
FROM messages m
WHERE (m.msgtag IS NULL OR LENGTH(m.msgtag) < 1)
  AND m.msgpatient = sqlc.arg(patient_id)
  AND m.msgread = 0
  AND m.msgby = sqlc.arg(user_id);

-- List all distinct message tags
-- name: ListMessageTags :many
SELECT DISTINCT msgtag FROM messages WHERE msgtag IS NOT NULL AND msgtag != '' ORDER BY msgtag;

-- Get messages by tag
-- name: MessagesByTag :many
SELECT * FROM messages WHERE msgtag = sqlc.arg(tag) ORDER BY msgtime DESC;

-- Delete messages by IDs, scoped to the owning session user.
--
-- The `msgfor` predicate is part of the statement text on purpose: the only
-- caller (api/messages.go messagesDelete) used to delete arbitrary ids from the
-- request body, so any authenticated user could mass-delete another user's
-- secure messages. Scoping in SQL means a non-owned id is simply not matched —
-- it cannot be reintroduced by a handler-level mistake.
-- name: DeleteMessagesForUser :execresult
DELETE FROM messages
WHERE msgfor = sqlc.arg(user_id)
  AND id IN (sqlc.slice(ids));
