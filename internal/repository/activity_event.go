package repository

import (
	"context"
	"task-tracker/internal/domain"
)

type ActivityEventRepository interface {
	Create(ctx context.Context, event *domain.ActivityEvent) error
	ListByIssue(ctx context.Context, issueID string) ([]*domain.ActivityEvent, error)
}
