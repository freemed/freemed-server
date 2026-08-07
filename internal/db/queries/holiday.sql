-- name: ListHolidays :many
SELECT * FROM holiday ORDER BY holiday_date;

-- name: CheckHoliday :one
SELECT * FROM holiday WHERE holiday_date = sqlc.arg(holiday_date) LIMIT 1;
