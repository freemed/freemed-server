-- name: ListVitals :many
SELECT * FROM vitals
WHERE patient = sqlc.arg(patient_id)
ORDER BY date_taken DESC;

-- name: GetLatestVitals :one
SELECT * FROM vitals
WHERE patient = sqlc.arg(patient_id)
ORDER BY date_taken DESC
LIMIT 1;

-- name: CreateVitals :execresult
INSERT INTO vitals (
  patient, date_taken, systolic, diastolic, heart_rate,
  respiratory_rate, temperature, oxygen_saturation,
  height_cm, weight_kg, bmi, notes, user,
  created_at, updated_at
) VALUES (
  sqlc.arg(patient_id), sqlc.arg(date_taken), sqlc.arg(systolic), sqlc.arg(diastolic),
  sqlc.arg(heart_rate), sqlc.arg(respiratory_rate), sqlc.arg(temperature),
  sqlc.arg(oxygen_saturation), sqlc.arg(height_cm), sqlc.arg(weight_kg),
  sqlc.arg(bmi), sqlc.arg(notes), sqlc.arg(user_id),
  NOW(), NOW()
);
