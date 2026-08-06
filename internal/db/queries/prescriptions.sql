-- List active prescriptions for a patient, ordered by date
-- name: ListPrescriptions :many
SELECT * FROM prescriptions
WHERE patient = sqlc.arg(patient_id)
  AND status = 'active'
  AND deleted_at IS NULL
ORDER BY date_written DESC;

-- Create a new prescription
-- name: CreatePrescription :execresult
INSERT INTO prescriptions (
  patient, drug_name, dosage, frequency, quantity, refills,
  date_written, prescribing_provider, pharmacy, status, notes, user
) VALUES (
  sqlc.arg(patient), sqlc.arg(drug_name), sqlc.arg(dosage),
  sqlc.arg(frequency), sqlc.arg(quantity), sqlc.arg(refills),
  sqlc.arg(date_written), sqlc.arg(prescribing_provider),
  sqlc.arg(pharmacy), sqlc.arg(status), sqlc.arg(notes),
  sqlc.arg(user)
);

-- Discontinue a prescription (set status to 'discontinued')
-- name: DiscontinuePrescription :exec
UPDATE prescriptions
SET status = 'discontinued', updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;
