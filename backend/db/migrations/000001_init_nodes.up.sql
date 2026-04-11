CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE node_type AS ENUM ('folder', 'note', 'task');

CREATE TABLE nodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    parent_id UUID REFERENCES nodes(id) ON DELETE CASCADE,
    type node_type NOT NULL,
    title VARCHAR(255) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE note_contents (
    node_id UUID PRIMARY KEY REFERENCES nodes(id) ON DELETE CASCADE,
    content JSONB NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_nodes_parent_id ON nodes(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_nodes_user_id ON nodes(user_id) WHERE deleted_at IS NULL;