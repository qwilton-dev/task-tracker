package repository

import (
	"context"
	"task-tracker/internal/domain"
)

type LabelRepository interface {
	CreateLabel(ctx context.Context, label *domain.Label) error
	GetLabelByID(ctx context.Context, id string) (*domain.Label, error)
	ListLabelsByWorkspace(ctx context.Context, workspaceID string) ([]*domain.Label, error)
	UpdateLabel(ctx context.Context, label *domain.Label) error
	DeleteLabel(ctx context.Context, id string) error

	AttachLabel(ctx context.Context, issueID, labelID string) error
	DetachLabel(ctx context.Context, issueID, labelID string) error
	ListLabelsByIssue(ctx context.Context, issueID string) ([]*domain.Label, error)
}
