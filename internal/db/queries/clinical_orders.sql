-- List clinical orders for a patient, ordered by date ordered descending
-- name: ListClinicalOrders :many
SELECT * FROM clinical_orders
WHERE patient = sqlc.arg(patient_id)
  AND active = 'active'
ORDER BY date_ordered DESC;

-- Create a new clinical order entry
-- name: CreateClinicalOrder :execresult
INSERT INTO clinical_orders (
  patient, order_type, description, status, date_ordered,
  ordering_provider, notes, user, active, created_at, updated_at
) VALUES (
  sqlc.arg(patient_id), sqlc.arg(order_type), sqlc.arg(description),
  sqlc.arg(status), sqlc.arg(date_ordered),
  sqlc.arg(ordering_provider), sqlc.arg(notes),
  sqlc.arg(user_id), 'active', NOW(), NOW()
);
