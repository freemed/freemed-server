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
