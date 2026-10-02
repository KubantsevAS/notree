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

-- name: GetDescendants :many
WITH RECURSIVE descendants AS (
    SELECT
        child.id,
        child.parent_id,
        child.sort_order,
        ARRAY[child.sort_order] AS so_path,
        ARRAY[child.id] AS id_path
    FROM nodes AS child
    WHERE child.parent_id = $1
      AND child.user_id = $2
      AND child.deleted_at IS NULL

    UNION ALL

    SELECT
        child.id,
        child.parent_id,
        child.sort_order,
        d.so_path || child.sort_order,
        d.id_path || child.id
    FROM descendants AS d
    JOIN nodes AS child
      ON child.parent_id = d.id
     AND child.user_id = $2
     AND child.deleted_at IS NULL
)
SELECT n.*
FROM descendants AS d
JOIN nodes AS n
  ON n.id = d.id
 AND n.user_id = $2
 AND n.deleted_at IS NULL
ORDER BY d.so_path, d.id_path;

-- name: GetSubtree :many
WITH RECURSIVE subtree AS (
    SELECT
        node.id,
        node.parent_id,
        ARRAY[node.sort_order] AS so_path,
        ARRAY[node.id] AS id_path
    FROM nodes AS node
    WHERE node.id = $1
      AND node.user_id = $2
      AND node.deleted_at IS NULL

    UNION ALL

    SELECT
        child.id,
        child.parent_id,
        s.so_path || child.sort_order,
        s.id_path || child.id
    FROM subtree AS s
    JOIN nodes AS child
      ON child.parent_id = s.id
     AND child.user_id = $2
     AND child.deleted_at IS NULL
)
SELECT n.*
FROM subtree AS s
JOIN nodes AS n
  ON n.id = s.id
 AND n.user_id = $2
 AND n.deleted_at IS NULL
ORDER BY s.so_path, s.id_path;

-- name: GetRoot :one
WITH RECURSIVE ancestors AS (
    SELECT
        node.id,
        node.parent_id
    FROM nodes AS node
    WHERE node.id = $1
      AND node.user_id = $2
      AND node.deleted_at IS NULL

    UNION ALL

    SELECT
        parent.id,
        parent.parent_id
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
WHERE a.parent_id IS NULL;


-- name: MoveNode :one
UPDATE nodes
SET
    parent_id = sqlc.narg('parent_id'),
    sort_order = @sort_order,
    updated_at = NOW()
WHERE id = @id AND user_id = @user_id AND deleted_at IS NULL
RETURNING parent_id, sort_order, updated_at;

-- name: LockUserHierarchy :exec
SELECT pg_advisory_xact_lock(hashtextextended('hierarchy:' || (sqlc.arg('user_id')::uuid)::text, 0));

-- name: GetPrevSiblingRank :one
SELECT n.sort_order
FROM nodes AS n
WHERE n.parent_id IS NOT DISTINCT FROM sqlc.narg('parent_id')::uuid
  AND n.user_id = @user_id
  AND n.deleted_at IS NULL
  AND n.id <> @exclude_id
  AND (
      sqlc.narg('before_rank')::bigint IS NULL
      OR (n.sort_order, n.id) < (sqlc.narg('before_rank')::bigint, sqlc.narg('before_id')::uuid)
  )
ORDER BY n.sort_order DESC, n.id DESC
LIMIT 1;

-- name: RebalanceChildren :exec
UPDATE nodes AS n
SET sort_order = ranked.position * sqlc.arg('gap')::bigint
FROM (
    SELECT
        c.id,
        ROW_NUMBER() OVER (ORDER BY c.sort_order, c.id) AS position
    FROM nodes AS c
    WHERE c.parent_id IS NOT DISTINCT FROM sqlc.narg('parent_id')::uuid
      AND c.user_id = @user_id
      AND c.deleted_at IS NULL
      AND c.id <> @exclude_id
) AS ranked
WHERE n.id = ranked.id;

-- name: IsInSubtree :one
WITH RECURSIVE ancestors AS (
    SELECT
        node.id,
        node.parent_id,
        ARRAY[node.id] AS id_path
    FROM nodes AS node
    WHERE node.id = @node_id
      AND node.user_id = @user_id

    UNION ALL

    SELECT
        parent.id,
        parent.parent_id,
        ancestors.id_path || parent.id
    FROM ancestors
    JOIN nodes AS parent
      ON parent.id = ancestors.parent_id
     AND parent.user_id = @user_id
    WHERE NOT parent.id = ANY(ancestors.id_path)
)
SELECT COALESCE(bool_or(a.id = sqlc.arg('root_id')::uuid), false)::boolean AS is_in_subtree
FROM ancestors AS a;
