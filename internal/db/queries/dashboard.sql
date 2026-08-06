-- Dashboard: count of non-archived patients
-- name: DashboardPatientCount :one
SELECT COUNT(*) AS total FROM patient WHERE ptarchive = 0;

-- Dashboard: count of today's appointments (excluding cancelled/noshow)
-- name: DashboardTodayAppointmentsCount :one
SELECT COUNT(*) AS total FROM scheduler
WHERE DATE(caldateof) = CURDATE()
  AND calstatus NOT IN ('cancelled', 'noshow');

-- Dashboard: count of unread messages for a user
-- name: DashboardUnreadMessagesCount :one
SELECT COUNT(*) AS total FROM messages
WHERE msgfor = sqlc.arg(user_id)
  AND msgread = 0;

-- Dashboard: count of active authorizations that haven't expired
-- name: DashboardActiveAuthorizationsCount :one
SELECT COUNT(*) AS total FROM authorizations
WHERE authdtend >= CURDATE()
  AND active = 'active';
