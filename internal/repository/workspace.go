package repository

import (
	"context"
	"task-tracker/internal/domain"
)

type WorkspaceRepository interface {
	CreateWorkspace(ctx context.Context, workspace *domain.Workspace) error
	GetWorkspaceByID(ctx context.Context, id string) (*domain.Workspace, error)
	GetWorkspaceBySlug(ctx context.Context, slug string) (*domain.Workspace, error)
	UpdateWorkspace(ctx context.Context, workspace *domain.Workspace) error
	DeleteWorkspace(ctx context.Context, id string) error
	ListWorkspaces(ctx context.Context, userID string) ([]*domain.Workspace, error)
	AddMember(ctx context.Context, workspaceID, userID, role string) error
}
