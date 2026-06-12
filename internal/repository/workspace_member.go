package repository

import (
	"context"
	"task-tracker/internal/domain"
)

type WorkspaceMemberRepository interface {
	CreateWorkspaceMember(ctx context.Context, workspaceMember *domain.WorkspaceMember) error
	GetWorkspaceMembers(ctx context.Context, workspaceId string) ([]*domain.WorkspaceMember, error)
	DeleteWorkspaceMember(ctx context.Context, workspaceMember *domain.WorkspaceMember) error
	UpdateWorkspaceMemberRole(ctx context.Context, workspaceMember *domain.WorkspaceMember) error
	GetRole(ctx context.Context, workspaceSlug, userID string) (string, error)
	GetRoleByProjectID(ctx context.Context, projectID, userID string) (string, error)
	GetRoleByIssueID(ctx context.Context, issueID, userID string) (string, error)
}
