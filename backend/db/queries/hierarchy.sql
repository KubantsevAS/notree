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
        parent.id,
        parent.parent_id,
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
        parent.id,
        parent.parent_id,
        ancestors.depth + 1
    FROM ancestors
    JOIN nodes AS parent
      ON parent.id = ancestors.parent_id
     AND parent.user_id = $2
     AND parent.deleted_at IS NULL
)
SELECT n.*
FROM ancestors AS a
JOIN nodes AS n
  ON n.id = a.id
 AND n.user_id = $2
 AND n.deleted_at IS NULL
ORDER BY a.depth DESC;