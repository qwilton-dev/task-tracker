package service

import (
	"context"

	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
)

type LabelService struct {
	repo repository.LabelRepository
}

func NewLabelService(repo repository.LabelRepository) *LabelService {
	return &LabelService{repo: repo}
}

func (s *LabelService) CreateLabel(ctx context.Context, workspaceID, name, color string) (*domain.Label, error) {
	label, err := domain.NewLabel(workspaceID, name, color)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateLabel(ctx, label); err != nil {
		return nil, err
	}
	return label, nil
}

func (s *LabelService) GetLabel(ctx context.Context, id string) (*domain.Label, error) {
	return s.repo.GetLabelByID(ctx, id)
}

func (s *LabelService) ListLabels(ctx context.Context, workspaceID string) ([]*domain.Label, error) {
	return s.repo.ListLabelsByWorkspace(ctx, workspaceID)
}

func (s *LabelService) UpdateLabel(ctx context.Context, id, name, color string) (*domain.Label, error) {
	label, err := s.repo.GetLabelByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		label.Name = name
	}
	if color != "" {
		label.Color = color
	}
	if err := s.repo.UpdateLabel(ctx, label); err != nil {
		return nil, err
	}
	return label, nil
}

func (s *LabelService) DeleteLabel(ctx context.Context, id string) error {
	return s.repo.DeleteLabel(ctx, id)
}

func (s *LabelService) AttachLabel(ctx context.Context, issueID, labelID string) error {
	return s.repo.AttachLabel(ctx, issueID, labelID)
}

func (s *LabelService) DetachLabel(ctx context.Context, issueID, labelID string) error {
	return s.repo.DetachLabel(ctx, issueID, labelID)
}

func (s *LabelService) ListLabelsByIssue(ctx context.Context, issueID string) ([]*domain.Label, error) {
	return s.repo.ListLabelsByIssue(ctx, issueID)
}
