-- name: GetBillkeyById :one
SELECT * FROM billkey WHERE id = sqlc.arg(id);

-- name: InsertBillkey :execresult
INSERT INTO billkey (
  billkeydate,
  billkey,
  bkprocs,
  created_at,
  updated_at
) VALUES (
  sqlc.arg(billkeydate),
  sqlc.arg(billkey),
  sqlc.arg(bkprocs),
  NOW(),
  NOW()
);
