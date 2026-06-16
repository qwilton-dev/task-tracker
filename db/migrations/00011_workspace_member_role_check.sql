-- +goose Up
ALTER TABLE workspace_member ADD CONSTRAINT workspace_member_role_check CHECK (role IN ('owner', 'member', 'viewer'));

-- +goose Down
ALTER TABLE workspace_member DROP CONSTRAINT IF EXISTS workspace_member_role_check;
