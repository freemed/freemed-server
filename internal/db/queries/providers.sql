-- name: ListProviders :many
SELECT id, phylname, phyfname, phynpi, physpecialties FROM physician ORDER BY phylname;

-- name: LookupNPI :many
SELECT * FROM physician WHERE phynpi = sqlc.arg(npi);
