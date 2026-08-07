-- name: ListReports :many
SELECT * FROM patient_reporting ORDER BY report_name;

-- name: GetReportByUuid :one
SELECT * FROM patient_reporting WHERE report_uuid = sqlc.arg(report_uuid);

-- name: GetReportById :one
SELECT * FROM patient_reporting WHERE id = sqlc.arg(id);
