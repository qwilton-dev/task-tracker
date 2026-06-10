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
	
	err = s.repo.AddMember(ctx, ws.ID, userID, "admin")
	if err != nil {
		return nil, err
	}

	return ws, nil
}

func (s *WorkspaceService) ListWorkspaces(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	return s.repo.ListWorkspaces(ctx, userID)
}
