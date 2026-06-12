package service

import (
	"context"
	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
)

type ProjectService struct {
	projectRepo   repository.ProjectRepository
	workspaceRepo repository.WorkspaceRepository
}

func NewProjectService(projectRepo repository.ProjectRepository, workspaceRepo repository.WorkspaceRepository) *ProjectService {
	return &ProjectService{projectRepo: projectRepo, workspaceRepo: workspaceRepo}
}

func (s *ProjectService) CreateProject(ctx context.Context, workspaceSlug, name, key string) (*domain.Project, error) {
	ws, err := s.workspaceRepo.GetWorkspaceBySlug(ctx, workspaceSlug)
	if err != nil {
		return nil, err
	}

	p, err := domain.NewProject(ws.ID, name, key)
	if err != nil {
		return nil, err
	}

	err = s.projectRepo.CreateProject(ctx, p)
	return p, err
}

func (s *ProjectService) ListProjects(ctx context.Context, workspaceSlug string) ([]*domain.Project, error) {
	ws, err := s.workspaceRepo.GetWorkspaceBySlug(ctx, workspaceSlug)
	if err != nil {
		return nil, err
	}
	return s.projectRepo.GetProjectsByWorkspace(ctx, ws.ID)
}
