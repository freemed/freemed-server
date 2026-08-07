-- name: ListBillkeyMonths :many
SELECT
  id,
  billkeydate,
  bkprocs
FROM billkey
ORDER BY billkeydate DESC;
