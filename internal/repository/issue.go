package repository

import (
	"context"

	"task-tracker/internal/domain"
)

type IssueRepository interface {
	CreateIssue(ctx context.Context, issue *domain.Issue) error
	CreateIssueTx(ctx context.Context, issue *domain.Issue) error
	GetIssueByID(ctx context.Context, id string) (*domain.Issue, error)
	ListIssuesByProject(ctx context.Context, projectID string, filters IssueFilters) ([]*domain.Issue, error)
	UpdateIssue(ctx context.Context, issue *domain.Issue) error
	DeleteIssue(ctx context.Context, id string) error
	MoveIssue(ctx context.Context, id, status string, position float64) error
	GetMaxNumber(ctx context.Context, projectID string) (int, error)
}

type IssueFilters struct {
	Status   string
	Assignee string
	Q        string
}
