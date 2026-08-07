-- List letters for a patient, ordered by date sent descending
-- name: ListLetters :many
SELECT * FROM letters
WHERE patient = sqlc.arg(patient_id)
  AND active = 'active'
ORDER BY date_sent DESC;

-- Create a new letter entry
-- name: CreateLetter :execresult
INSERT INTO letters (
  patient, letter_type, recipient, subject, body, date_sent,
  user, active, created_at, updated_at
) VALUES (
  sqlc.arg(patient_id), sqlc.arg(letter_type), sqlc.arg(recipient),
  sqlc.arg(subject), sqlc.arg(body), sqlc.arg(date_sent),
  sqlc.arg(user_id), 'active', NOW(), NOW()
);
