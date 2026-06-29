package service

import (
	"context"
	"testing"
	"time"

	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
)

type mockWorkspaceRepo struct {
	createFn    func(ctx context.Context, ws *domain.Workspace) error
	getByIDFn   func(ctx context.Context, id string) (*domain.Workspace, error)
	updateFn    func(ctx context.Context, ws *domain.Workspace) error
	deleteFn    func(ctx context.Context, id string) error
	listFn      func(ctx context.Context, userID string) ([]*domain.Workspace, error)
	addMemberFn func(ctx context.Context, workspaceID, userID, role string) error
}

var _ repository.WorkspaceRepository = (*mockWorkspaceRepo)(nil)

func (m *mockWorkspaceRepo) CreateWorkspace(ctx context.Context, ws *domain.Workspace) error {
	if m.createFn != nil {
		return m.createFn(ctx, ws)
	}
	ws.ID = "ws-1"
	ws.CreatedAt = time.Now()
	return nil
}
func (m *mockWorkspaceRepo) GetWorkspaceByID(ctx context.Context, id string) (*domain.Workspace, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrWorkspaceNotFound
}
func (m *mockWorkspaceRepo) UpdateWorkspace(ctx context.Context, ws *domain.Workspace) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, ws)
	}
	return nil
}
func (m *mockWorkspaceRepo) DeleteWorkspace(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *mockWorkspaceRepo) ListWorkspaces(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockWorkspaceRepo) AddMember(ctx context.Context, workspaceID, userID, role string) error {
	if m.addMemberFn != nil {
		return m.addMemberFn(ctx, workspaceID, userID, role)
	}
	return nil
}

func TestWorkspaceService_CreateWorkspace(t *testing.T) {
	var gotMember struct {
		workspaceID, userID, role string
	}
	repo := &mockWorkspaceRepo{
		addMemberFn: func(ctx context.Context, workspaceID, userID, role string) error {
			gotMember.workspaceID = workspaceID
			gotMember.userID = userID
			gotMember.role = role
			return nil
		},
	}
	svc := NewWorkspaceService(repo)

	ws, err := svc.CreateWorkspace(context.Background(), "user-1", "My Workspace")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ws.ID != "ws-1" {
		t.Fatalf("expected ws-1, got %s", ws.ID)
	}
	if ws.Name != "My Workspace" {
		t.Fatalf("expected My Workspace, got %s", ws.Name)
	}
	if gotMember.userID != "user-1" {
		t.Fatalf("expected user-1, got %s", gotMember.userID)
	}
	if gotMember.role != "owner" {
		t.Fatalf("expected owner, got %s", gotMember.role)
	}
}

func TestWorkspaceService_CreateWorkspace_EmptyName(t *testing.T) {
	svc := NewWorkspaceService(&mockWorkspaceRepo{})
	_, err := svc.CreateWorkspace(context.Background(), "user-1", "")
	if err != domain.ErrWorkspaceNameRequired {
		t.Fatalf("expected ErrWorkspaceNameRequired, got %v", err)
	}
}

func TestWorkspaceService_ListWorkspaces(t *testing.T) {
	repo := &mockWorkspaceRepo{
		listFn: func(ctx context.Context, userID string) ([]*domain.Workspace, error) {
			return []*domain.Workspace{
				{ID: "ws-1", Name: "WS1"},
				{ID: "ws-2", Name: "WS2"},
			}, nil
		},
	}
	svc := NewWorkspaceService(repo)

	wss, err := svc.ListWorkspaces(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(wss) != 2 {
		t.Fatalf("expected 2, got %d", len(wss))
	}
}
