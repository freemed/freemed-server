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
