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

-- AttachPaymentToProcedure: link a payment to a procedure
-- name: AttachPaymentToProcedure :exec
UPDATE payrec
SET payrecproc = sqlc.arg(procedure_id)
WHERE id = sqlc.arg(payment_id);

-- UnattachedCopays: copay payments not yet attached to a procedure
-- name: UnattachedCopays :many
SELECT
  pr.id,
  pr.payrecdtadd AS date,
  pr.payrecamt AS amount,
  pr.payrectype AS type,
  pr.payrecdescrip AS description,
  pr.payrecproc AS procedure_id
FROM payrec pr
WHERE pr.payrecproc = 0
  AND pr.payrectype = 'copay'
  AND pr.active = 'active'
ORDER BY pr.payrecdtadd DESC;

-- UnattachedDeductibles: deductible payments not yet attached to a procedure
-- name: UnattachedDeductibles :many
SELECT
  pr.id,
  pr.payrecdtadd AS date,
  pr.payrecamt AS amount,
  pr.payrectype AS type,
  pr.payrecdescrip AS description,
  pr.payrecproc AS procedure_id
FROM payrec pr
WHERE pr.payrecproc = 0
  AND pr.payrectype = 'deductible'
  AND pr.active = 'active'
ORDER BY pr.payrecdtadd DESC;

-- UnattachedPayments: non-copay/non-deductible payments not yet attached to a procedure
-- name: UnattachedPayments :many
SELECT
  pr.id,
  pr.payrecdtadd AS date,
  pr.payrecamt AS amount,
  pr.payrectype AS type,
  pr.payrecdescrip AS description,
  pr.payrecproc AS procedure_id
FROM payrec pr
WHERE pr.payrecproc = 0
  AND pr.payrectype NOT IN ('copay', 'deductible')
  AND pr.active = 'active'
ORDER BY pr.payrecdtadd DESC;

-- RemovePaymentAsMistake: mark a payment as a mistake (soft delete)
-- name: RemovePaymentAsMistake :exec
UPDATE payrec
SET active = 'inactive'
WHERE id = sqlc.arg(payment_id);

-- GetCoverageCopayInfo: get copay-related coverage info for a patient
-- name: GetCoverageCopayInfo :one
SELECT
  pc.id,
  pc.patient,
  pc.insurance_company,
  pc.coverage_type,
  pc.policy_number,
  pc.group_number,
  pc.effective_date,
  pc.termination_date,
  pc.primary_coverage
FROM patient_coverage pc
WHERE pc.patient = sqlc.arg(patient_id)
  AND pc.active = 'active'
LIMIT 1;

-- GetCoverageDeductibleInfo: get deductible-related coverage info for a patient
-- name: GetCoverageDeductibleInfo :one
SELECT
  pc.id,
  pc.patient,
  pc.insurance_company,
  pc.coverage_type,
  pc.policy_number,
  pc.group_number,
  pc.effective_date,
  pc.termination_date,
  pc.primary_coverage
FROM patient_coverage pc
WHERE pc.patient = sqlc.arg(patient_id)
  AND pc.active = 'active'
LIMIT 1;
