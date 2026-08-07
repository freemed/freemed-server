-- name: ListFacilities :many
SELECT * FROM facility ORDER BY psrname;

-- name: GetDefaultFacility :one
SELECT * FROM facility LIMIT 1;
