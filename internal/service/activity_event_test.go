package service

import (
	"context"
	"testing"
	"time"

	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
)

type mockActivityRepo struct {
	createFn func(ctx context.Context, event *domain.ActivityEvent) error
	listFn   func(ctx context.Context, issueID string) ([]*domain.ActivityEvent, error)
}

var _ repository.ActivityEventRepository = (*mockActivityRepo)(nil)

func (m *mockActivityRepo) Create(ctx context.Context, event *domain.ActivityEvent) error {
	if m.createFn != nil {
		return m.createFn(ctx, event)
	}
	event.ID = "event-1"
	event.CreatedAt = time.Now()
	return nil
}
func (m *mockActivityRepo) ListByIssue(ctx context.Context, issueID string) ([]*domain.ActivityEvent, error) {
	if m.listFn != nil {
		return m.listFn(ctx, issueID)
	}
	return nil, nil
}

func TestActivityEventService_CreateActivityEvent(t *testing.T) {
	repo := &mockActivityRepo{}
	svc := NewActivityEventService(repo)

	event, err := svc.CreateActivityEvent(context.Background(), "issue-1", "user-1", "issue.created", map[string]string{"title": "Test"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if event.IssueID != "issue-1" {
		t.Fatalf("expected issue-1, got %s", event.IssueID)
	}
	if event.ActorID != "user-1" {
		t.Fatalf("expected user-1, got %s", event.ActorID)
	}
	if event.Type != "issue.created" {
		t.Fatalf("expected issue.created, got %s", event.Type)
	}
}

func TestActivityEventService_ListByIssue(t *testing.T) {
	repo := &mockActivityRepo{
		listFn: func(ctx context.Context, issueID string) ([]*domain.ActivityEvent, error) {
			return []*domain.ActivityEvent{
				{ID: "e1", IssueID: issueID, Type: "issue.created"},
				{ID: "e2", IssueID: issueID, Type: "issue.moved"},
			}, nil
		},
	}
	svc := NewActivityEventService(repo)

	events, err := svc.ListActivityEventsByIssue(context.Background(), "issue-1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2, got %d", len(events))
	}
}
