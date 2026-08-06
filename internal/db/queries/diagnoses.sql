-- PatientDiagnoses: list all ICD-9 diagnoses for a patient
-- name: PatientDiagnoses :many
SELECT
  'primary' AS dx_type,
  i.icd9code,
  i.icd9descrip
FROM patient p
JOIN icd9 i ON i.id = p.ptdiag1
WHERE p.id = sqlc.arg('patient_id')
  AND p.ptdiag1 IS NOT NULL
UNION ALL
SELECT
  'secondary' AS dx_type,
  i.icd9code,
  i.icd9descrip
FROM patient p
JOIN icd9 i ON i.id = p.ptdiag2
WHERE p.id = sqlc.arg('patient_id')
  AND p.ptdiag2 IS NOT NULL
UNION ALL
SELECT
  'secondary' AS dx_type,
  i.icd9code,
  i.icd9descrip
FROM patient p
JOIN icd9 i ON i.id = p.ptdiag3
WHERE p.id = sqlc.arg('patient_id')
  AND p.ptdiag3 IS NOT NULL
UNION ALL
SELECT
  'secondary' AS dx_type,
  i.icd9code,
  i.icd9descrip
FROM patient p
JOIN icd9 i ON i.id = p.ptdiag4
WHERE p.id = sqlc.arg('patient_id')
  AND p.ptdiag4 IS NOT NULL;
