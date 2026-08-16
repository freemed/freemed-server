-- name: ListDicomByPatient :many
SELECT id, created_at, updated_at, d_md5, d_patient,
       d_study_description, d_filename, d_study_date, d_institution_name,
       d_institution_address, d_study_uid, d_series_uid, d_sop_uid,
       d_referring_provider, d_modality, d_patient_id, storage_status, user
FROM dicom
WHERE d_patient = sqlc.arg(patient_id)
  AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetDicom :one
SELECT * FROM dicom
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: GetDicomBySop :one
SELECT * FROM dicom
WHERE d_study_uid = sqlc.arg(study_uid)
  AND d_series_uid = sqlc.arg(series_uid)
  AND d_sop_uid = sqlc.arg(sop_uid)
  AND deleted_at IS NULL
ORDER BY id DESC
LIMIT 1;

-- name: ListDicomStudies :many
SELECT d_study_uid, d_study_date, d_study_description, d_patient_id,
       d_patient, d_modality
FROM dicom
WHERE deleted_at IS NULL
  AND (sqlc.narg(patient_id) IS NULL OR d_patient_id = sqlc.narg(patient_id))
  AND (sqlc.narg(study_uid) IS NULL OR d_study_uid = sqlc.narg(study_uid))
GROUP BY d_study_uid, d_study_date, d_study_description, d_patient_id, d_patient, d_modality
ORDER BY d_study_date DESC;

-- name: CreateDicom :execresult
INSERT INTO dicom (
  created_at, updated_at, d_md5, d_patient, d_study_description, d_filename,
  d_study_date, d_institution_name, d_institution_address, d_study_uid,
  d_series_uid, d_sop_uid, d_referring_provider, d_modality, d_patient_id,
  d_xml_data, storage_status, user, d_data
) VALUES (
  NOW(), NOW(),
  sqlc.arg(md5), sqlc.arg(patient_id), sqlc.arg(study_description), sqlc.arg(filename),
  sqlc.arg(study_date), sqlc.arg(institution_name), sqlc.arg(institution_address), sqlc.arg(study_uid),
  sqlc.arg(series_uid), sqlc.arg(sop_uid), sqlc.arg(referring_provider), sqlc.arg(modality), sqlc.arg(dicom_patient_id),
  sqlc.arg(xml_data), sqlc.arg(storage_status), sqlc.arg(user_id), sqlc.arg(data)
);

-- name: DeleteDicom :exec
UPDATE dicom SET updated_at = NOW(), deleted_at = NOW()
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: CountDicomByMd5 :one
SELECT COUNT(*) FROM dicom
WHERE d_md5 = sqlc.arg(md5)
  AND deleted_at IS NULL;
