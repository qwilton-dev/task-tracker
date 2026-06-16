-- +goose Up
CREATE TABLE IF NOT EXISTS invites (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id  UUID NOT NULL REFERENCES workspace(id) ON DELETE CASCADE,
    email         TEXT NOT NULL,
    role          TEXT NOT NULL CHECK (role IN ('member', 'viewer')),
    token         TEXT NOT NULL UNIQUE,
    expires_at    TIMESTAMPTZ NOT NULL,
    accepted_at   TIMESTAMPTZ,
    created_by    UUID NOT NULL REFERENCES users(id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_invites_workspace_id ON invites(workspace_id);
CREATE INDEX idx_invites_token ON invites(token);

-- +goose Down
DROP TABLE IF EXISTS invites;
