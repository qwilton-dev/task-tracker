package service

import (
	"context"
	"testing"
	"time"

	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
)

type mockIssueRepo struct {
	createFn   func(ctx context.Context, issue *domain.Issue) error
	getByIDFn  func(ctx context.Context, id string) (*domain.Issue, error)
	listFn     func(ctx context.Context, projectID string, filters repository.IssueFilters) ([]*domain.Issue, error)
	updateFn   func(ctx context.Context, issue *domain.Issue) error
	deleteFn   func(ctx context.Context, id string) error
	moveFn     func(ctx context.Context, id, status string, position float64) error
	maxNumFn   func(ctx context.Context, projectID string) (int, error)
}

var _ repository.IssueRepository = (*mockIssueRepo)(nil)

func (m *mockIssueRepo) CreateIssue(ctx context.Context, issue *domain.Issue) error {
	if m.createFn != nil {
		return m.createFn(ctx, issue)
	}
	issue.ID = "issue-1"
	issue.CreatedAt = time.Now()
	issue.UpdatedAt = time.Now()
	return nil
}
func (m *mockIssueRepo) CreateIssueTx(ctx context.Context, issue *domain.Issue) error {
	if m.createFn != nil {
		return m.createFn(ctx, issue)
	}
	issue.ID = "issue-1"
	issue.CreatedAt = time.Now()
	issue.UpdatedAt = time.Now()
	return nil
}
func (m *mockIssueRepo) GetIssueByID(ctx context.Context, id string) (*domain.Issue, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrIssueNotFound
}
func (m *mockIssueRepo) ListIssuesByProject(ctx context.Context, projectID string, filters repository.IssueFilters) ([]*domain.Issue, error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, filters)
	}
	return nil, nil
}
func (m *mockIssueRepo) UpdateIssue(ctx context.Context, issue *domain.Issue) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, issue)
	}
	return nil
}
func (m *mockIssueRepo) DeleteIssue(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *mockIssueRepo) MoveIssue(ctx context.Context, id, status string, position float64) error {
	if m.moveFn != nil {
		return m.moveFn(ctx, id, status, position)
	}
	return nil
}
func (m *mockIssueRepo) GetMaxNumber(ctx context.Context, projectID string) (int, error) {
	if m.maxNumFn != nil {
		return m.maxNumFn(ctx, projectID)
	}
	return 0, nil
}

func TestIssueService_CreateIssue(t *testing.T) {
	repo := &mockIssueRepo{
		createFn: func(ctx context.Context, issue *domain.Issue) error {
			issue.ID = "issue-1"
			issue.Number = 1
			issue.CreatedAt = time.Now()
			issue.UpdatedAt = time.Now()
			return nil
		},
	}
	projRepo := &mockProjectRepo{}
	svc := NewIssueService(repo, projRepo)

	issue, err := svc.CreateIssue(context.Background(), "proj-1", "Fix bug", "description", "user-1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if issue.Title != "Fix bug" {
		t.Fatalf("expected Fix bug, got %s", issue.Title)
	}
}

func TestIssueService_CreateIssue_EmptyTitle(t *testing.T) {
	svc := NewIssueService(&mockIssueRepo{}, &mockProjectRepo{})
	_, err := svc.CreateIssue(context.Background(), "proj-1", "", "", "user-1")
	if err != domain.ErrIssueTitleRequired {
		t.Fatalf("expected ErrIssueTitleRequired, got %v", err)
	}
}

func TestIssueService_MoveIssue_InvalidStatus(t *testing.T) {
	svc := NewIssueService(&mockIssueRepo{}, &mockProjectRepo{})
	err := svc.MoveIssue(context.Background(), "issue-1", "invalid", 0)
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestIssueService_MoveIssue_ValidStatus(t *testing.T) {
	svc := NewIssueService(&mockIssueRepo{}, &mockProjectRepo{})
	err := svc.MoveIssue(context.Background(), "issue-1", "todo", 100)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestIssueService_UpdateIssue_TitleTooLong(t *testing.T) {
	repo := &mockIssueRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.Issue, error) {
			return &domain.Issue{ID: id, Title: "old"}, nil
		},
	}
	svc := NewIssueService(repo, &mockProjectRepo{})
	title := ""
	for i := 0; i < 501; i++ {
		title += "a"
	}
	_, err := svc.UpdateIssue(context.Background(), "issue-1", title, "", "", "")
	if err != domain.ErrIssueTitleTooLong {
		t.Fatalf("expected ErrIssueTitleTooLong, got %v", err)
	}
}

func TestIssueService_UpdateIssue_InvalidPriority(t *testing.T) {
	repo := &mockIssueRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.Issue, error) {
			return &domain.Issue{ID: id, Title: "old"}, nil
		},
	}
	svc := NewIssueService(repo, &mockProjectRepo{})
	_, err := svc.UpdateIssue(context.Background(), "issue-1", "", "", "invalid", "")
	if err != domain.ErrIssueInvalidPriority {
		t.Fatalf("expected ErrIssueInvalidPriority, got %v", err)
	}
}
