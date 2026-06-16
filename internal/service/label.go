package service

import (
	"context"

	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
)

type LabelService struct {
	repo          repository.LabelRepository
	workspaceRepo repository.WorkspaceRepository
	issueRepo     repository.IssueRepository
	activity      *ActivityEventService
}

func NewLabelService(repo repository.LabelRepository, workspaceRepo repository.WorkspaceRepository, issueRepo repository.IssueRepository, activity *ActivityEventService) *LabelService {
	return &LabelService{repo: repo, workspaceRepo: workspaceRepo, issueRepo: issueRepo, activity: activity}
}

func (s *LabelService) resolveWorkspaceID(ctx context.Context, slugOrID string) (string, error) {
	ws, err := s.workspaceRepo.GetWorkspaceBySlug(ctx, slugOrID)
	if err != nil {
		return slugOrID, nil
	}
	return ws.ID, nil
}

func (s *LabelService) CreateLabel(ctx context.Context, workspaceID, name, color string) (*domain.Label, error) {
	resolved, err := s.resolveWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	label, err := domain.NewLabel(resolved, name, color)
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
	resolved, err := s.resolveWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListLabelsByWorkspace(ctx, resolved)
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

func (s *LabelService) AttachLabel(ctx context.Context, issueID, labelID, actorID string) error {
	if err := s.repo.AttachLabel(ctx, issueID, labelID); err != nil {
		return err
	}
	label, _ := s.repo.GetLabelByID(ctx, labelID)
	labelName := ""
	if label != nil {
		labelName = label.Name
	}
	s.activity.CreateActivityEvent(ctx, issueID, actorID, "issue.label_added", map[string]string{
		"label_id": labelID,
		"name":     labelName,
	})
	return nil
}

func (s *LabelService) DetachLabel(ctx context.Context, issueID, labelID, actorID string) error {
	if err := s.repo.DetachLabel(ctx, issueID, labelID); err != nil {
		return err
	}
	label, _ := s.repo.GetLabelByID(ctx, labelID)
	labelName := ""
	if label != nil {
		labelName = label.Name
	}
	s.activity.CreateActivityEvent(ctx, issueID, actorID, "issue.label_removed", map[string]string{
		"label_id": labelID,
		"name":     labelName,
	})
	return nil
}

func (s *LabelService) ListLabelsByIssue(ctx context.Context, issueID string) ([]*domain.Label, error) {
	return s.repo.ListLabelsByIssue(ctx, issueID)
}
