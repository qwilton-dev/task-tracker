package service

import (
	"context"
	"fmt"
	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
)

type IssueService struct {
	issueRepo  repository.IssueRepository
	projectRepo repository.ProjectRepository
}

func NewIssueService(issueRepo repository.IssueRepository, projectRepo repository.ProjectRepository) *IssueService {
	return &IssueService{issueRepo: issueRepo, projectRepo: projectRepo}
}

func (s *IssueService) CreateIssue(ctx context.Context, projectID, title, description, createdBy string) (*domain.Issue, error) {
	issue, err := domain.NewIssue(projectID, title, createdBy)
	if err != nil {
		return nil, err
	}
	issue.Description = description

	if err := s.issueRepo.CreateIssueTx(ctx, issue); err != nil {
		return nil, err
	}
	return issue, nil
}

func (s *IssueService) GetIssue(ctx context.Context, id string) (*domain.Issue, error) {
	return s.issueRepo.GetIssueByID(ctx, id)
}

func (s *IssueService) ListIssues(ctx context.Context, projectID string, filters repository.IssueFilters) ([]*domain.Issue, error) {
	return s.issueRepo.ListIssuesByProject(ctx, projectID, filters)
}

func (s *IssueService) UpdateIssue(ctx context.Context, id, title, description, priority, assigneeID string) (*domain.Issue, error) {
	issue, err := s.issueRepo.GetIssueByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if title != "" {
		if len(title) > 500 {
			return nil, domain.ErrIssueTitleTooLong
		}
		issue.Title = title
	}
	if description != "" {
		issue.Description = description
	}
	if priority != "" {
		if !domain.ValidatePriority(priority) {
			return nil, domain.ErrIssueInvalidPriority
		}
		issue.Priority = priority
	}
	issue.AssigneeID = assigneeID

	if err := s.issueRepo.UpdateIssue(ctx, issue); err != nil {
		return nil, err
	}
	return issue, nil
}

func (s *IssueService) DeleteIssue(ctx context.Context, id string) error {
	return s.issueRepo.DeleteIssue(ctx, id)
}

func (s *IssueService) MoveIssue(ctx context.Context, id, status string, position float64) error {
	if !domain.ValidateStatus(status) {
		return fmt.Errorf("%w: %s", domain.ErrIssueInvalidStatus, status)
	}
	return s.issueRepo.MoveIssue(ctx, id, status, position)
}
