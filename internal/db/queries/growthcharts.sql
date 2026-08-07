-- name: ListGrowthCharts :many
SELECT * FROM growthcharts WHERE patient = sqlc.arg(patient) ORDER BY record_date DESC;
