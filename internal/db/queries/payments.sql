-- Patient payments: list all payments for a patient
-- name: PatientPayments :many
SELECT
  pr.id,
  pr.payrecdtadd AS date,
  pr.payrecamt AS amount,
  pr.payrectype AS type,
  pr.payrecdescrip AS description,
  pr.payrecproc AS procedure_id
FROM payrec pr
WHERE pr.payrecpatient = sqlc.arg(patient_id)
  AND pr.active = 'active'
ORDER BY pr.payrecdtadd DESC;

-- Patient ledger: combined procedure charges + payments view (simplified)
-- name: PatientLedger :many
SELECT
  'charge' AS entry_type,
  p.id,
  p.procdt AS date,
  p.proccharges AS amount,
  p.procbalcurrent AS balance,
  cpt.abbrev AS cpt_code,
  p.proccomment AS description,
  p.procstatus AS status
FROM procrec p
LEFT OUTER JOIN cpt cpt ON cpt.id = p.proccpt
WHERE p.procpatient = sqlc.arg(patient_id)
UNION ALL
SELECT
  'payment' AS entry_type,
  pr.id,
  pr.payrecdtadd AS date,
  pr.payrecamt AS amount,
  0.0 AS balance,
  '' AS cpt_code,
  pr.payrecdescrip AS description,
  '' AS status
FROM payrec pr
WHERE pr.payrecpatient = sqlc.arg(patient_id)
  AND pr.active = 'active'
ORDER BY date DESC;
