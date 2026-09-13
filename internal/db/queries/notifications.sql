-- UserNotifications returns the most recent notifications for a user
-- name: UserNotifications :many
SELECT * FROM systemnotification
WHERE nuser = sqlc.arg(user_id)
ORDER BY stamp DESC
LIMIT 50;

-- UnreadCount returns the total number of notifications for a user
-- name: UnreadCount :one
SELECT COUNT(*) FROM systemnotification
WHERE nuser = sqlc.arg(user_id);

-- PatientNotifications returns notifications for a specific patient
-- name: PatientNotifications :many
SELECT * FROM systemnotification
WHERE npatient = sqlc.arg(patient_id)
ORDER BY stamp DESC;

-- NotificationsFromTimestamp returns a user's notifications since a timestamp.
-- The nuser filter is mandatory: systemnotification carries npatient as well as
-- nuser, so an unscoped query leaked both other users' rows and other patients'
-- rows to any authenticated caller. LIMIT bounds the poll response, matching
-- UserNotifications above.
-- name: NotificationsFromTimestamp :many
SELECT * FROM systemnotification
WHERE nuser = sqlc.arg(user_id)
  AND stamp > sqlc.arg(since)
ORDER BY stamp DESC
LIMIT 50;

-- LatestTimestamp returns the latest notification timestamp
-- name: LatestTimestamp :one
SELECT MAX(stamp) AS ts FROM systemnotification;

-- SystemTaskInboxCount returns the count of task notifications for a user
-- name: SystemTaskInboxCount :one
SELECT COUNT(*) FROM systemnotification
WHERE nuser = sqlc.arg(user_id);

-- SystemTaskPatientInbox returns task notifications for a specific patient
-- name: SystemTaskPatientInbox :many
SELECT * FROM systemnotification
WHERE npatient = sqlc.arg(patient_id)
ORDER BY stamp DESC;

-- SystemTaskPatientInboxCount returns the count of task notifications for a patient
-- name: SystemTaskPatientInboxCount :one
SELECT COUNT(*) FROM systemnotification
WHERE npatient = sqlc.arg(patient_id);

-- SystemTaskUserInbox returns task notifications for a user
-- name: SystemTaskUserInbox :many
SELECT * FROM systemnotification
WHERE nuser = sqlc.arg(user_id)
ORDER BY stamp DESC;

-- name: CreateNotification :execresult
INSERT INTO systemnotification (
  created_at, updated_at, stamp, nuser, ntext, naction, nmodule, npatient
) VALUES (
  NOW(), NOW(), NOW(), sqlc.arg(user_id), sqlc.arg(text),
  sqlc.arg(action), sqlc.arg(module), sqlc.arg(patient_id)
);
