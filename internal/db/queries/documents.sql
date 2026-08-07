-- ============================================================================
-- Unfiled Documents
-- ============================================================================

-- name: ListUnfiledDocs :many
SELECT * FROM unfiled_docs
WHERE active = 'active'
  AND deleted_at IS NULL
ORDER BY received_date DESC;

-- name: CountUnfiledDocs :one
SELECT COUNT(*) AS total FROM unfiled_docs
WHERE active = 'active'
  AND deleted_at IS NULL;

-- name: AssignUnfiledDoc :exec
UPDATE unfiled_docs
SET assigned_to = sqlc.arg(assigned_to),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND active = 'active'
  AND deleted_at IS NULL;

-- ============================================================================
-- Unread Documents
-- ============================================================================

-- name: ListUnreadDocs :many
SELECT * FROM unread_docs
WHERE active = 'active'
  AND deleted_at IS NULL
ORDER BY sent_date DESC;

-- name: CountUnreadDocs :one
SELECT COUNT(*) AS total FROM unread_docs
WHERE active = 'active'
  AND deleted_at IS NULL;

-- name: ReviewUnreadDoc :exec
UPDATE unread_docs
SET reviewed = true,
    review_date = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND active = 'active'
  AND deleted_at IS NULL;

-- name: ReassignUnreadDoc :exec
UPDATE unread_docs
SET assigned_to = sqlc.arg(assigned_to),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND active = 'active'
  AND deleted_at IS NULL;
