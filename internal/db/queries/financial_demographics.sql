-- name: ListFinancialDemographicsByPatient :many
SELECT * FROM financial_demographics
WHERE patient = sqlc.arg(patient_id)
  AND active = 'active'
  AND deleted_at IS NULL
ORDER BY id DESC;

-- name: CreateFinancialDemographics :execresult
INSERT INTO financial_demographics (
  created_at, updated_at, patient, income, id_type,
  id_issuer, id_number, id_expire, household_size,
  spouse, children, other_dependents, free_text,
  entry_desc, entry_ts, user, active
) VALUES (
  NOW(), NOW(), sqlc.arg(patient_id), sqlc.arg(income), sqlc.arg(id_type),
  sqlc.arg(id_issuer), sqlc.arg(id_number), sqlc.arg(id_expire),
  sqlc.arg(household_size),
  sqlc.arg(spouse), sqlc.arg(children), sqlc.arg(other_dependents),
  sqlc.narg(free_text),
  sqlc.arg(entry_desc), NOW(), sqlc.arg(user_id), 'active'
);
