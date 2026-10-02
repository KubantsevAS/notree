-- name: CreateNode :one
INSERT INTO nodes (user_id, parent_id, type, title, sort_order)
VALUES (
    @user_id,
    sqlc.narg('parent_id'),
    @type,
    @title,
    COALESCE((
        SELECT MAX(sibling.sort_order)
        FROM nodes AS sibling
        WHERE sibling.user_id = @user_id
          AND sibling.parent_id IS NOT DISTINCT FROM sqlc.narg('parent_id')::uuid
          AND sibling.deleted_at IS NULL
    ), 0) + sqlc.arg('gap')::bigint
)
RETURNING *;

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
