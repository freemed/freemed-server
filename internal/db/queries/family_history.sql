-- name: ListFamilyHistory :many
SELECT * FROM family_history
WHERE patient = ? AND active = 'active'
ORDER BY created_at DESC;

-- name: CreateFamilyHistory :execresult
INSERT INTO family_history (patient, relationship, condition_name, icd10_code, onset_age, deceased, notes, user, active, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'active', NOW(), NOW());

-- name: UpdateFamilyHistory :exec
UPDATE family_history
SET relationship = ?, condition_name = ?, icd10_code = ?, onset_age = ?, deceased = ?, notes = ?, updated_at = NOW()
WHERE id = ? AND patient = ? AND active = 'active';

-- name: RemoveFamilyHistory :exec
UPDATE family_history SET active = 'inactive', deleted_at = NOW(), updated_at = NOW()
WHERE id = ? AND patient = ?;
