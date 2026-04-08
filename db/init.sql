CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE node_type AS ENUM ('folder', 'note', 'task');

CREATE TABLE nodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    parent_id UUID,
    user_id UUID,
    type node_type,
    title VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE note_contents (
    node_id UUID PRIMARY KEY REFERENCES nodes(id) ON DELETE CASCADE,
    content JSONB NOT NULL
);
