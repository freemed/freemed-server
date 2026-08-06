-- Zipcode picklist: search by state and city
-- name: CityStateZipByStateCity :many
SELECT * FROM zipcodes
WHERE state = UPPER(sqlc.arg('state'))
  AND city LIKE CONCAT('%', sqlc.arg('city'), '%')
LIMIT 20;

-- Zipcode picklist: search by city only
-- name: CityStateZipByCity :many
SELECT * FROM zipcodes
WHERE city LIKE CONCAT('%', sqlc.arg('city'), '%')
LIMIT 20;

-- Zipcode picklist: search by zip code
-- name: CityStateZipByZip :many
SELECT * FROM zipcodes
WHERE zip LIKE CONCAT('%', sqlc.arg('zip'), '%')
LIMIT 20;
