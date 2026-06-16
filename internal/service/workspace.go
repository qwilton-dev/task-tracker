package service

import (
	"context"
	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
)

type WorkspaceService struct {
	repo repository.WorkspaceRepository
}

func NewWorkspaceService(repo repository.WorkspaceRepository) *WorkspaceService {
	return &WorkspaceService{repo: repo}
}

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, userID, name, slug string) (*domain.Workspace, error) {
	ws, err := domain.NewWorkspace(name, slug)
	if err != nil {
		return nil, err
	}
	err = s.repo.CreateWorkspace(ctx, ws)
	if err != nil {
		return nil, err
	}
	
	err = s.repo.AddMember(ctx, ws.ID, userID, "owner")
	if err != nil {
		return nil, err
	}

	return ws, nil
}

func (s *WorkspaceService) ListWorkspaces(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	return s.repo.ListWorkspaces(ctx, userID)
}

func (s *WorkspaceService) GetWorkspace(ctx context.Context, id string) (*domain.Workspace, error) {
	return s.repo.GetWorkspaceByID(ctx, id)
}

func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, id, name, slug string) (*domain.Workspace, error) {
	ws, err := s.repo.GetWorkspaceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		ws.Name = name
	}
	if slug != "" {
		ws.Slug = slug
	}
	if err := s.repo.UpdateWorkspace(ctx, ws); err != nil {
		return nil, err
	}
	return ws, nil
}

func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, id string) error {
	return s.repo.DeleteWorkspace(ctx, id)
}
