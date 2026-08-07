-- name: ListCertificationsByPatient :many
SELECT * FROM certifications
WHERE patient = sqlc.arg(patient_id)
  AND active = 'active'
  AND deleted_at IS NULL
ORDER BY id DESC;

-- name: CreateCertification :execresult
INSERT INTO certifications (
  created_at, updated_at, patient, cert_type, cert_form_num,
  cert_desc, cert_form_data, user, active
) VALUES (
  NOW(), NOW(), sqlc.arg(patient_id), sqlc.arg(cert_type),
  sqlc.narg(cert_form_num),
  sqlc.arg(cert_desc), sqlc.narg(cert_form_data),
  sqlc.arg(user_id), 'active'
);
