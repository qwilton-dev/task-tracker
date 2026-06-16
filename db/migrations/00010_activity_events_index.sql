-- +goose Up
CREATE INDEX idx_activity_events_issue_created ON activity_events(issue_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_activity_events_issue_created;
