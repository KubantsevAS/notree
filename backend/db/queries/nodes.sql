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
  AND user_id = $2
  AND deleted_at IS NULL
LIMIT 1;

-- name: SoftDeleteNodeCascade :many
WITH RECURSIVE subtree AS (
    SELECT id 
    FROM nodes AS n
    WHERE n.id = $1 AND n.user_id = $2 AND n.deleted_at IS NULL
    
    UNION ALL
    
    SELECT c.id 
    FROM nodes AS c
    INNER JOIN subtree AS p ON c.parent_id = p.id
    WHERE c.user_id = $2 AND c.deleted_at IS NULL
)
UPDATE nodes
SET deleted_at = NOW() 
WHERE id IN (SELECT id FROM subtree)
RETURNING id;

-- name: UpdateNode :one
UPDATE nodes
SET
    type = COALESCE(sqlc.narg('type'), type),
    title = COALESCE(sqlc.narg('title'), title),
    updated_at = NOW()
WHERE id = @id AND user_id = @user_id AND deleted_at IS NULL
RETURNING type, title, updated_at;

-- name: MoveNode :one
UPDATE nodes
SET
    parent_id = CASE
                    WHEN sqlc.narg('update_parent')::boolean THEN sqlc.narg('parent_id')
                    ELSE parent_id
                END,
    sort_order = COALESCE(sqlc.narg('sort_order'), sort_order),
    updated_at = NOW()
WHERE id = @id AND user_id = @user_id AND deleted_at IS NULL
RETURNING parent_id, sort_order, updated_at;

-- name: GetNodeAncestors :many
WITH RECURSIVE ancestors(node_id, parent_id) AS (
    SELECT n.id, n.parent_id
    FROM nodes n
    WHERE n.id = $1 AND n.deleted_at IS NULL

    UNION ALL

    SELECT c.id, c.parent_id
    FROM nodes c
    JOIN ancestors a ON c.id = a.parent_id
    WHERE c.deleted_at IS NULL
)
SELECT node_id
FROM ancestors;