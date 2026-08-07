-- name: ListRemindersByUser :many
SELECT * FROM reminders
WHERE user = sqlc.arg(user_id)
  AND deleted_at IS NULL
ORDER BY due_date ASC;

-- name: ListRemindersByPatient :many
SELECT * FROM reminders
WHERE patient = sqlc.arg(patient_id)
  AND deleted_at IS NULL
ORDER BY due_date ASC;

-- name: CreateReminder :execresult
INSERT INTO reminders (
  created_at, updated_at, user, patient,
  title, description, due_date, priority
) VALUES (
  NOW(), NOW(), sqlc.arg(user_id), sqlc.narg(patient_id),
  sqlc.arg(title), sqlc.narg(description), sqlc.narg(due_date),
  sqlc.arg(priority)
);

-- name: UpdateReminderStatus :exec
UPDATE reminders
SET
  updated_at = NOW(),
  status = sqlc.arg(status),
  completed_at = CASE
    WHEN sqlc.arg(status) = 'completed' THEN NOW()
    ELSE NULL
  END
WHERE id = sqlc.arg(reminder_id)
  AND deleted_at IS NULL;

-- name: DeleteReminder :exec
UPDATE reminders
SET updated_at = NOW(), deleted_at = NOW()
WHERE id = sqlc.arg(reminder_id)
  AND deleted_at IS NULL;
