-- List all drug samples
-- name: ListDrugSamples :many
SELECT * FROM drugsampleinv
ORDER BY created_at DESC;

-- Create drug sample entry
-- name: CreateDrugSample :execresult
INSERT INTO drugsampleinv (
  drugcode, drugndc, drugclass, packagecount, location,
  drugco, drugrep, invoice, samplecount, samplecountremain,
  lot, expiration, received, assignedto, loguser, logdate
) VALUES (
  sqlc.arg(drugcode), sqlc.arg(drugndc), sqlc.arg(drugclass),
  sqlc.arg(packagecount), sqlc.arg(location), sqlc.arg(drugco),
  sqlc.arg(drugrep), sqlc.arg(invoice), sqlc.arg(samplecount),
  sqlc.arg(samplecountremain), sqlc.arg(lot), sqlc.narg(expiration),
  sqlc.narg(received), sqlc.arg(assignedto), sqlc.arg(loguser),
  sqlc.arg(logdate)
);
