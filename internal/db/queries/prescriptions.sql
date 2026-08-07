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

-- List prescriptions for a patient with pharmacy details and refill info
-- name: ListPrescriptionsWithPharmacy :many
SELECT
  p.id,
  p.created_at,
  p.updated_at,
  p.deleted_at,
  p.patient,
  p.drug_name,
  p.dosage,
  p.frequency,
  p.quantity,
  p.refills,
  p.date_written,
  p.prescribing_provider,
  p.status,
  p.notes,
  p.user,
  ph.id AS pharmacy_id,
  ph.phname AS pharmacy_name,
  ph.phcity AS pharmacy_city,
  ph.phstate AS pharmacy_state,
  (SELECT COUNT(*) FROM rxrefillrequest r
   WHERE r.patient = p.patient
     AND r.rxorig LIKE CONCAT('%', p.id, '%')
     AND r.deleted_at IS NULL) AS refills_used,
  (SELECT MAX(r.approved) FROM rxrefillrequest r
   WHERE r.patient = p.patient
     AND r.rxorig LIKE CONCAT('%', p.id, '%')
     AND r.approved IS NOT NULL
     AND r.deleted_at IS NULL) AS last_fill_date
FROM prescriptions p
LEFT JOIN pharmacy ph ON ph.phname = p.pharmacy AND ph.deleted_at IS NULL
WHERE p.patient = sqlc.arg(patient_id)
  AND p.status = 'active'
  AND p.deleted_at IS NULL
ORDER BY p.date_written DESC;
