-- name: CreateNode :one
INSERT INTO nodes (user_id, parent_id, type, title, sort_order)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetChildren :many
SELECT * FROM nodes
WHERE parent_id = $1 
  AND user_id = $2 
  AND deleted_at IS NULL
ORDER BY sort_order ASC;

-- name: GetNodeByID :one
SELECT * FROM nodes
WHERE id = $1 
  AND deleted_at IS NULL
LIMIT 1;

-- name: SoftDeleteNode :exec
UPDATE nodes
SET deleted_at = NOW()
WHERE id = $1 AND user_id = $2;