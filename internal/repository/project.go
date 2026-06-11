package repository

import (
	"context"
	"task-tracker/internal/domain"
)

type ProjectRepository interface {
	CreateProject(ctx context.Context, project *domain.Project) error
	GetProjectsByWorkspace(ctx context.Context, workspaceID string) ([]*domain.Project, error)
}
