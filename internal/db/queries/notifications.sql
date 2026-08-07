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

-- NotificationsFromTimestamp returns notifications since a given timestamp
-- name: NotificationsFromTimestamp :many
SELECT * FROM systemnotification
WHERE stamp > sqlc.arg(since)
ORDER BY stamp DESC;

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
