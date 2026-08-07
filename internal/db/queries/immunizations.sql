-- name: ListImmunizations :many
SELECT * FROM immunization
WHERE patient = sqlc.arg(patient_id)
  AND active = 'active'
  AND deleted_at IS NULL
ORDER BY dateof DESC;

-- name: CreateImmunization :execresult
INSERT INTO immunization (
  created_at, updated_at, dateof, patient, provider,
  admin_provider, immunization, route, body_site,
  manufacturer, lot_number, previous_doses, recovered, notes, user, active
) VALUES (
  NOW(), NOW(), sqlc.arg(dateof), sqlc.arg(patient_id),
  sqlc.arg(provider), sqlc.arg(admin_provider),
  sqlc.arg(immunization), sqlc.arg(route), sqlc.arg(body_site),
  sqlc.arg(manufacturer), sqlc.arg(lot_number),
  sqlc.arg(previous_doses), sqlc.arg(recovered),
  sqlc.arg(notes), sqlc.arg(user_id), 'active'
);
