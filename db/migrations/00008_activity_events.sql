-- +goose Up
CREATE TABLE IF NOT EXISTS activity_events (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id     UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    actor_id      UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    event_type    TEXT NOT NULL,
    payload       JSONB NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_activity_events_issue_id ON activity_events(issue_id);

-- +goose Down
DROP TABLE IF EXISTS activity_events;