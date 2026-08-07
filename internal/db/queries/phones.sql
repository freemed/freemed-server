-- Phone: list all active phones for a patient
-- name: ListPhones :many
SELECT
  id,
  patient,
  type,
  number,
  active,
  created_at,
  updated_at
FROM phone
WHERE patient = sqlc.arg(patient_id) AND active = 'active'
ORDER BY id ASC;

-- Phone: insert a new phone for a patient
-- name: CreatePhone :execresult
INSERT INTO phone (patient, type, number, active, user)
VALUES (
  sqlc.arg(patient_id),
  sqlc.arg(phone_type),
  sqlc.arg(number),
  'active',
  sqlc.arg(user_id)
);
