DROP INDEX IF EXISTS idx_nodes_parent_id;
CREATE INDEX idx_nodes_parent_id ON nodes(parent_id) WHERE deleted_at IS NULL;
