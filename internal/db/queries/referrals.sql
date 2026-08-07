-- List active referrals for a patient, ordered by referral date descending
-- name: ListReferrals :many
SELECT * FROM referrals
WHERE patient = sqlc.arg(patient_id)
  AND active = 'active'
ORDER BY date_referred DESC;

-- Create a new referral entry
-- name: CreateReferral :execresult
INSERT INTO referrals (
  patient, referring_provider, referral_to, referral_type,
  reason, status, date_referred, date_completed, notes, user,
  active, created_at, updated_at
) VALUES (
  sqlc.arg(patient_id), sqlc.arg(referring_provider), sqlc.arg(referral_to),
  sqlc.arg(referral_type), sqlc.arg(reason), sqlc.arg(status),
  sqlc.arg(date_referred), sqlc.arg(date_completed), sqlc.arg(notes),
  sqlc.arg(user_id), 'active', NOW(), NOW()
);
