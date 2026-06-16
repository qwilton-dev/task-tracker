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

	if key == "" {
		key = domain.GenerateUniqueKey(name, func(k string) bool {
			exists, _ := s.projectRepo.ExistsByKey(ctx, ws.ID, k)
			return exists
		})
	}

	p, err := domain.NewProject(ws.ID, name, key)
	if err != nil {
		return nil, err
	}

	err = s.projectRepo.CreateProject(ctx, p)
	return p, err
}

func (s *ProjectService) ListProjects(ctx context.Context, workspaceID string) ([]*domain.Project, error) {
	return s.projectRepo.GetProjectsByWorkspace(ctx, workspaceID)
}

func (s *ProjectService) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	return s.projectRepo.GetProjectByID(ctx, id)
}

func (s *ProjectService) UpdateProject(ctx context.Context, id, name, key string) (*domain.Project, error) {
	p, err := s.projectRepo.GetProjectByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		p.Name = name
	}
	if key != "" {
		if err := domain.ValidateProjectKey(key); err != nil {
			return nil, err
		}
		p.Key = key
	}
	if err := s.projectRepo.UpdateProject(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, id string) error {
	return s.projectRepo.DeleteProject(ctx, id)
}
