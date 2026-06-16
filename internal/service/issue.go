package service

import (
	"context"
	"encoding/json"
	"fmt"
	"task-tracker/internal/domain"
	"task-tracker/internal/events"
	"task-tracker/internal/repository"
	"time"
)

type IssueService struct {
	issueRepo   repository.IssueRepository
	projectRepo repository.ProjectRepository
	activity    *ActivityEventService
	publisher   *events.Publisher
}

func NewIssueService(activityEventSvc *ActivityEventService, issueRepo repository.IssueRepository, projectRepo repository.ProjectRepository, publisher *events.Publisher) *IssueService {
	return &IssueService{activity: activityEventSvc, issueRepo: issueRepo, projectRepo: projectRepo, publisher: publisher}
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
	s.activity.CreateActivityEvent(ctx, issue.ID, createdBy, "issue.created", map[string]string{
		"title": issue.Title,
	})
	s.publish(ctx, issue.ProjectID, "issue.created", issue)
	return issue, nil
}

func (s *IssueService) GetIssue(ctx context.Context, id string) (*domain.Issue, error) {
	return s.issueRepo.GetIssueByID(ctx, id)
}

func (s *IssueService) ListIssues(ctx context.Context, projectID string, filters repository.IssueFilters) ([]*domain.Issue, error) {
	return s.issueRepo.ListIssuesByProject(ctx, projectID, filters)
}

func (s *IssueService) UpdateIssue(ctx context.Context, id, title, description, priority, assigneeID, actorID string) (*domain.Issue, error) {
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
	s.activity.CreateActivityEvent(ctx, issue.ID, actorID, "issue.updated", map[string]string{
		"title": issue.Title,
	})
	s.publish(ctx, issue.ProjectID, "issue.updated", issue)
	return issue, nil
}

func (s *IssueService) DeleteIssue(ctx context.Context, id, actorID string) error {
	issue, err := s.issueRepo.GetIssueByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.issueRepo.DeleteIssue(ctx, id); err != nil {
		return err
	}
	s.activity.CreateActivityEvent(ctx, issue.ID, actorID, "issue.deleted", map[string]string{
		"title": issue.Title,
	})
	s.publish(ctx, issue.ProjectID, "issue.deleted", issue)
	return nil
}

func (s *IssueService) MoveIssue(ctx context.Context, id, status string, position float64, actorID string) error {
	issue, err := s.issueRepo.GetIssueByID(ctx, id)
	if err != nil {
		return err
	}
	if !domain.ValidateStatus(status) {
		return fmt.Errorf("%w: %s", domain.ErrIssueInvalidStatus, status)
	}
	if err := s.issueRepo.MoveIssue(ctx, id, status, position); err != nil {
		return err
	}
	s.activity.CreateActivityEvent(ctx, issue.ID, actorID, "issue.moved", map[string]string{
		"title": issue.Title,
	})
	s.publish(ctx, issue.ProjectID, "issue.moved", issue)
	return nil
}

func (s *IssueService) publish(ctx context.Context, projectID, eventType string, payload any) {
	if s.publisher == nil {
		return
	}
	data, _ := json.Marshal(payload)
	pubCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	s.publisher.Publish(pubCtx, projectID, events.Event{
		ProjectID: projectID,
		Type:      eventType,
		Payload:   data,
	})
}
