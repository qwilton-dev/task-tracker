package service

import (
	"context"
	"testing"

	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
)

type mockLabelRepo struct {
	createFn       func(ctx context.Context, label *domain.Label) error
	getByIDFn      func(ctx context.Context, id string) (*domain.Label, error)
	listByWsFn     func(ctx context.Context, workspaceID string) ([]*domain.Label, error)
	updateFn       func(ctx context.Context, label *domain.Label) error
	deleteFn       func(ctx context.Context, id string) error
	attachFn       func(ctx context.Context, issueID, labelID string) error
	detachFn       func(ctx context.Context, issueID, labelID string) error
	listByIssueFn  func(ctx context.Context, issueID string) ([]*domain.Label, error)
}

var _ repository.LabelRepository = (*mockLabelRepo)(nil)

func (m *mockLabelRepo) CreateLabel(ctx context.Context, label *domain.Label) error {
	if m.createFn != nil {
		return m.createFn(ctx, label)
	}
	label.ID = "label-1"
	return nil
}
func (m *mockLabelRepo) GetLabelByID(ctx context.Context, id string) (*domain.Label, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrLabelNotFound
}
func (m *mockLabelRepo) ListLabelsByWorkspace(ctx context.Context, workspaceID string) ([]*domain.Label, error) {
	if m.listByWsFn != nil {
		return m.listByWsFn(ctx, workspaceID)
	}
	return nil, nil
}
func (m *mockLabelRepo) UpdateLabel(ctx context.Context, label *domain.Label) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, label)
	}
	return nil
}
func (m *mockLabelRepo) DeleteLabel(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *mockLabelRepo) AttachLabel(ctx context.Context, issueID, labelID string) error {
	if m.attachFn != nil {
		return m.attachFn(ctx, issueID, labelID)
	}
	return nil
}
func (m *mockLabelRepo) DetachLabel(ctx context.Context, issueID, labelID string) error {
	if m.detachFn != nil {
		return m.detachFn(ctx, issueID, labelID)
	}
	return nil
}
func (m *mockLabelRepo) ListLabelsByIssue(ctx context.Context, issueID string) ([]*domain.Label, error) {
	if m.listByIssueFn != nil {
		return m.listByIssueFn(ctx, issueID)
	}
	return nil, nil
}

func newTestLabelService(repo repository.LabelRepository) *LabelService {
	return NewLabelService(repo, NewActivityEventService(&mockActivityEventRepo{}))
}

func TestLabelService_CreateLabel(t *testing.T) {
	svc := newTestLabelService(&mockLabelRepo{})
	label, err := svc.CreateLabel(context.Background(), "ws-1", "Bug", "#ff0000")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if label.ID != "label-1" {
		t.Fatalf("expected label-1, got %s", label.ID)
	}
	if label.Name != "Bug" {
		t.Fatalf("expected Bug, got %s", label.Name)
	}
}

func TestLabelService_CreateLabel_EmptyName(t *testing.T) {
	svc := newTestLabelService(&mockLabelRepo{})
	_, err := svc.CreateLabel(context.Background(), "ws-1", "", "#ff0000")
	if err != domain.ErrLabelNameRequired {
		t.Fatalf("expected ErrLabelNameRequired, got %v", err)
	}
}

func TestLabelService_CreateLabel_InvalidColor(t *testing.T) {
	svc := newTestLabelService(&mockLabelRepo{})
	_, err := svc.CreateLabel(context.Background(), "ws-1", "Bug", "red")
	if err != domain.ErrLabelColorInvalid {
		t.Fatalf("expected ErrLabelColorInvalid, got %v", err)
	}
}

func TestLabelService_ListLabels(t *testing.T) {
	repo := &mockLabelRepo{
		listByWsFn: func(ctx context.Context, workspaceID string) ([]*domain.Label, error) {
			return []*domain.Label{
				{ID: "l1", Name: "Bug", Color: "#ff0000"},
				{ID: "l2", Name: "Feature", Color: "#00ff00"},
			}, nil
		},
	}
	svc := newTestLabelService(repo)
	labels, err := svc.ListLabels(context.Background(), "ws-1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(labels) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(labels))
	}
}

func TestLabelService_UpdateLabel(t *testing.T) {
	repo := &mockLabelRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.Label, error) {
			return &domain.Label{ID: id, Name: "Bug", Color: "#ff0000"}, nil
		},
	}
	svc := newTestLabelService(repo)
	label, err := svc.UpdateLabel(context.Background(), "label-1", "Critical", "#ff00ff")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if label.Name != "Critical" {
		t.Fatalf("expected Critical, got %s", label.Name)
	}
	if label.Color != "#ff00ff" {
		t.Fatalf("expected #ff00ff, got %s", label.Color)
	}
}

func TestLabelService_UpdateLabel_NotFound(t *testing.T) {
	svc := newTestLabelService(&mockLabelRepo{})
	_, err := svc.UpdateLabel(context.Background(), "bad-id", "Bug", "#ff0000")
	if err != domain.ErrLabelNotFound {
		t.Fatalf("expected ErrLabelNotFound, got %v", err)
	}
}

func TestLabelService_DeleteLabel(t *testing.T) {
	svc := newTestLabelService(&mockLabelRepo{})
	if err := svc.DeleteLabel(context.Background(), "label-1"); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestLabelService_AttachLabel(t *testing.T) {
	svc := newTestLabelService(&mockLabelRepo{})
	if err := svc.AttachLabel(context.Background(), "issue-1", "label-1", "user-1"); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestLabelService_DetachLabel(t *testing.T) {
	svc := newTestLabelService(&mockLabelRepo{})
	if err := svc.DetachLabel(context.Background(), "issue-1", "label-1", "user-1"); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestLabelService_DetachLabel_NotFound(t *testing.T) {
	repo := &mockLabelRepo{
		detachFn: func(ctx context.Context, issueID, labelID string) error {
			return domain.ErrLabelNotFound
		},
	}
	svc := newTestLabelService(repo)
	err := svc.DetachLabel(context.Background(), "issue-1", "bad-label", "user-1")
	if err != domain.ErrLabelNotFound {
		t.Fatalf("expected ErrLabelNotFound, got %v", err)
	}
}

func TestLabelService_ListLabelsByIssue(t *testing.T) {
	repo := &mockLabelRepo{
		listByIssueFn: func(ctx context.Context, issueID string) ([]*domain.Label, error) {
			return []*domain.Label{
				{ID: "l1", Name: "Bug", Color: "#ff0000"},
			}, nil
		},
	}
	svc := newTestLabelService(repo)
	labels, err := svc.ListLabelsByIssue(context.Background(), "issue-1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(labels) != 1 {
		t.Fatalf("expected 1 label, got %d", len(labels))
	}
}
