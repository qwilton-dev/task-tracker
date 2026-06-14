-- +goose Up
CREATE TABLE IF NOT EXISTS label (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id  UUID NOT NULL REFERENCES workspace(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    color         TEXT NOT NULL,
    UNIQUE (workspace_id, name)
);
CREATE TABLE IF NOT EXISTS issue_label (
    issue_id   UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    label_id   UUID NOT NULL REFERENCES label(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (issue_id, label_id)
);
CREATE INDEX idx_issue_label_label_id ON issue_label(label_id);
-- +goose Down
DROP TABLE IF EXISTS issue_label;
DROP TABLE IF EXISTS label;