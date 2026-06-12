-- +goose Up
CREATE TABLE workspace_member (
    workspace_id UUID NOT NULL REFERENCES workspace(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role         TEXT NOT NULL,
    PRIMARY KEY (workspace_id, user_id)
);

CREATE INDEX idx_workspace_member_user_id ON workspace_member(user_id);

CREATE TABLE project (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id  UUID NOT NULL REFERENCES workspace(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    "key"         TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, "key")
);

-- +goose Down
DROP TABLE IF EXISTS project;
DROP TABLE IF EXISTS workspace_member;
DROP INDEX IF EXISTS idx_workspace_member_user_id;
