-- name: GetParent :one
SELECT parent.*
FROM nodes AS node
JOIN nodes AS parent
  ON parent.id = node.parent_id
 AND parent.user_id = node.user_id
WHERE node.id = $1
  AND node.user_id = $2
  AND node.deleted_at IS NULL
  AND parent.deleted_at IS NULL;

-- name: GetChildren :many
SELECT * FROM nodes
WHERE parent_id = $1 
  AND user_id = $2 
  AND deleted_at IS NULL
ORDER BY sort_order ASC;

-- name: GetAncestors :many
WITH RECURSIVE ancestors AS (
    SELECT
        parent.*,
        1 AS depth
    FROM nodes AS node
    JOIN nodes AS parent
      ON parent.id = node.parent_id
     AND parent.user_id = node.user_id
     AND parent.deleted_at IS NULL
    WHERE node.id = $1
      AND node.user_id = $2
      AND node.deleted_at IS NULL

    UNION ALL

    SELECT
        parent.*,
        ancestors.depth + 1
    FROM ancestors
    JOIN nodes AS parent
      ON parent.id = ancestors.parent_id
     AND parent.user_id = ancestors.user_id
     AND parent.deleted_at IS NULL
)
SELECT * FROM ancestors
ORDER BY depth DESC;