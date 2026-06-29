package service

import (
	"context"
	"testing"
	"time"

	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
)

type mockCommentRepo struct {
	createFn func(ctx context.Context, comment *domain.Comment) error
	listFn   func(ctx context.Context, issueID string) ([]*domain.Comment, error)
}

var _ repository.CommentRepository = (*mockCommentRepo)(nil)

func (m *mockCommentRepo) CreateComment(ctx context.Context, comment *domain.Comment) error {
	if m.createFn != nil {
		return m.createFn(ctx, comment)
	}
	comment.ID = "comment-1"
	comment.CreatedAt = time.Now()
	return nil
}
func (m *mockCommentRepo) ListCommentsByIssue(ctx context.Context, issueID string) ([]*domain.Comment, error) {
	if m.listFn != nil {
		return m.listFn(ctx, issueID)
	}
	return nil, nil
}

func TestCommentService_CreateComment(t *testing.T) {
	repo := &mockCommentRepo{}
	issueRepo := &mockIssueRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.Issue, error) {
			return &domain.Issue{ID: id, ProjectID: "proj-1"}, nil
		},
	}
	activityRepo := &mockActivityEventRepo{}
	svc := NewCommentService(repo, issueRepo, NewActivityEventService(activityRepo), nil)

	comment, err := svc.CreateComment(context.Background(), "issue-1", "user-1", "Hello")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if comment.ID != "comment-1" {
		t.Fatalf("expected comment-1, got %s", comment.ID)
	}
	if comment.Body != "Hello" {
		t.Fatalf("expected Hello, got %s", comment.Body)
	}
}

func TestCommentService_CreateComment_EmptyBody(t *testing.T) {
	svc := NewCommentService(&mockCommentRepo{}, &mockIssueRepo{}, NewActivityEventService(&mockActivityEventRepo{}), nil)
	_, err := svc.CreateComment(context.Background(), "issue-1", "user-1", "")
	if err != domain.ErrCommentBodyRequired {
		t.Fatalf("expected ErrCommentBodyRequired, got %v", err)
	}
}

func TestCommentService_ListComments(t *testing.T) {
	repo := &mockCommentRepo{
		listFn: func(ctx context.Context, issueID string) ([]*domain.Comment, error) {
			return []*domain.Comment{
				{ID: "c1", Body: "First"},
				{ID: "c2", Body: "Second"},
			}, nil
		},
	}
	svc := NewCommentService(repo, &mockIssueRepo{}, NewActivityEventService(&mockActivityEventRepo{}), nil)

	comments, err := svc.ListComments(context.Background(), "issue-1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(comments) != 2 {
		t.Fatalf("expected 2, got %d", len(comments))
	}
}
