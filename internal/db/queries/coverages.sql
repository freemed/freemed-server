-- ListCoverages: list active coverages for a patient with insurance company and coverage type names
-- name: ListCoverages :many
SELECT
  pc.id,
  pc.patient,
  pc.insurance_company,
  i.insconame AS insurance_company_name,
  pc.coverage_type,
  ct.covtpname AS coverage_type_name,
  pc.policy_number,
  pc.group_number,
  pc.effective_date,
  pc.termination_date,
  pc.primary_coverage,
  pc.active,
  pc.created_at,
  pc.updated_at
FROM patient_coverage pc
LEFT JOIN insco i ON pc.insurance_company = i.id
LEFT JOIN covtypes ct ON pc.coverage_type = ct.id
WHERE pc.patient = sqlc.arg(patient_id)
  AND pc.active = 'active'
ORDER BY pc.primary_coverage DESC, pc.effective_date DESC;

-- CreateCoverage: create a new patient coverage record
-- name: CreateCoverage :execresult
INSERT INTO patient_coverage (
  patient, insurance_company, coverage_type,
  policy_number, group_number, effective_date,
  termination_date, primary_coverage, active,
  created_at, updated_at
) VALUES (
  sqlc.arg(patient), sqlc.arg(insurance_company), sqlc.arg(coverage_type),
  sqlc.arg(policy_number), sqlc.arg(group_number), sqlc.arg(effective_date),
  sqlc.arg(termination_date), sqlc.arg(primary_coverage), 'active',
  NOW(), NOW()
);

-- RemoveCoverage: soft-delete a patient coverage record
-- name: RemoveCoverage :exec
UPDATE patient_coverage
SET active = 'inactive', updated_at = NOW()
WHERE id = sqlc.arg(coverage_id)
  AND patient = sqlc.arg(patient_id);

-- GetPrimaryCoverage: get the primary active coverage for a patient
-- name: GetPrimaryCoverage :one
SELECT
  pc.id,
  pc.patient,
  pc.insurance_company,
  i.insconame AS insurance_company_name,
  pc.coverage_type,
  ct.covtpname AS coverage_type_name,
  pc.policy_number,
  pc.group_number,
  pc.effective_date,
  pc.termination_date,
  pc.primary_coverage,
  pc.active,
  pc.created_at,
  pc.updated_at
FROM patient_coverage pc
LEFT JOIN insco i ON pc.insurance_company = i.id
LEFT JOIN covtypes ct ON pc.coverage_type = ct.id
WHERE pc.patient = sqlc.arg(patient_id)
  AND pc.primary_coverage = 1
  AND pc.active = 'active'
LIMIT 1;
