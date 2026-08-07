-- Addresses: list all active addresses for a patient
-- name: ListAddresses :many
SELECT
  id,
  patient,
  line1,
  line2,
  city,
  stpr,
  postal,
  zip,
  active,
  created_at,
  updated_at
FROM patient_address
WHERE patient = sqlc.arg(patient_id) AND active = 1
ORDER BY id ASC;

-- Addresses: update an address for a patient
-- name: UpdateAddress :exec
UPDATE patient_address
SET
  line1 = sqlc.arg(line1),
  line2 = sqlc.arg(line2),
  city = sqlc.arg(city),
  stpr = sqlc.arg(stpr),
  postal = sqlc.arg(postal),
  updated_at = NOW()
WHERE id = sqlc.arg(address_id) AND patient = sqlc.arg(patient_id);

-- Addresses: deactivate an address (soft delete)
-- name: DeleteAddress :exec
UPDATE patient_address
SET active = 0, updated_at = NOW()
WHERE id = sqlc.arg(address_id) AND patient = sqlc.arg(patient_id);

-- Addresses: deactivate all addresses for a patient (bulk soft delete)
-- name: DeleteAllAddresses :exec
UPDATE patient_address
SET active = 0, updated_at = NOW()
WHERE patient = sqlc.arg(patient_id);

-- Addresses: insert a new address for a patient
-- name: SetAddresses :execresult
INSERT INTO patient_address (patient, line1, line2, city, stpr, postal, active)
VALUES (
  sqlc.arg(patient_id),
  sqlc.arg(line1),
  sqlc.arg(line2),
  sqlc.arg(city),
  sqlc.arg(stpr),
  sqlc.arg(postal),
  1
);
