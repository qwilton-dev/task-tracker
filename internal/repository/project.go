package repository

import (
	"context"
	"task-tracker/internal/domain"
)

type ProjectRepository interface {
	CreateProject(ctx context.Context, project *domain.Project) error
	GetProjectsByWorkspace(ctx context.Context, workspaceID string) ([]*domain.Project, error)
	GetProjectByID(ctx context.Context, id string) (*domain.Project, error)
	ExistsByKey(ctx context.Context, workspaceID, key string) (bool, error)
	UpdateProject(ctx context.Context, project *domain.Project) error
	DeleteProject(ctx context.Context, id string) error
}
