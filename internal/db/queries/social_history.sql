-- name: ListSocialHistory :many
SELECT * FROM social_history
WHERE patient = ? AND active = 'active'
ORDER BY recorded_date DESC;

-- name: GetLatestSocialHistory :one
SELECT * FROM social_history
WHERE patient = ? AND active = 'active'
ORDER BY recorded_date DESC
LIMIT 1;

-- name: CreateSocialHistory :execresult
INSERT INTO social_history (patient, smoking_status, smoking_detail, alcohol_use, alcohol_detail, drug_use, drug_detail, exercise_frequency, occupation, living_situation, food_insecurity, transportation_access, notes, recorded_date, user, active, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'active', NOW(), NOW());

-- name: RemoveSocialHistory :exec
UPDATE social_history SET active = 'inactive', deleted_at = NOW(), updated_at = NOW()
WHERE id = ? AND patient = ?;
