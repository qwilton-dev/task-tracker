package service

import (
	"context"
	"testing"
	"time"

	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
)

type mockProjectRepo struct {
	createFn  func(ctx context.Context, p *domain.Project) error
	listFn    func(ctx context.Context, workspaceID string) ([]*domain.Project, error)
	getByIDFn func(ctx context.Context, id string) (*domain.Project, error)
	updateFn  func(ctx context.Context, p *domain.Project) error
	deleteFn  func(ctx context.Context, id string) error
}

var _ repository.ProjectRepository = (*mockProjectRepo)(nil)

func (m *mockProjectRepo) CreateProject(ctx context.Context, p *domain.Project) error {
	if m.createFn != nil {
		return m.createFn(ctx, p)
	}
	p.ID = "proj-1"
	p.CreatedAt = time.Now()
	return nil
}
func (m *mockProjectRepo) GetProjectsByWorkspace(ctx context.Context, workspaceID string) ([]*domain.Project, error) {
	if m.listFn != nil {
		return m.listFn(ctx, workspaceID)
	}
	return nil, nil
}
func (m *mockProjectRepo) ExistsByKey(ctx context.Context, workspaceID, key string) (bool, error) {
	return false, nil
}
func (m *mockProjectRepo) GetProjectByID(ctx context.Context, id string) (*domain.Project, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrProjectNotFound
}
func (m *mockProjectRepo) UpdateProject(ctx context.Context, p *domain.Project) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, p)
	}
	return nil
}
func (m *mockProjectRepo) DeleteProject(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func TestProjectService_CreateProject(t *testing.T) {
	wsRepo := &mockWorkspaceRepo{
		getBySlugFn: func(ctx context.Context, slug string) (*domain.Workspace, error) {
			return &domain.Workspace{ID: "ws-1", Slug: slug}, nil
		},
	}
	projRepo := &mockProjectRepo{}
	svc := NewProjectService(projRepo, wsRepo)

	p, err := svc.CreateProject(context.Background(), "my-ws", "Backend", "BE")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if p.ID != "proj-1" {
		t.Fatalf("expected proj-1, got %s", p.ID)
	}
	if p.WorkspaceID != "ws-1" {
		t.Fatalf("expected ws-1, got %s", p.WorkspaceID)
	}
	if p.Key != "BE" {
		t.Fatalf("expected BE, got %s", p.Key)
	}
}

func TestProjectService_CreateProject_InvalidKey(t *testing.T) {
	wsRepo := &mockWorkspaceRepo{
		getBySlugFn: func(ctx context.Context, slug string) (*domain.Workspace, error) {
			return &domain.Workspace{ID: "ws-1"}, nil
		},
	}
	svc := NewProjectService(&mockProjectRepo{}, wsRepo)

	_, err := svc.CreateProject(context.Background(), "ws", "P", "x")
	if err != domain.ErrProjectKeyInvalid {
		t.Fatalf("expected ErrProjectKeyInvalid, got %v", err)
	}

	_, err = svc.CreateProject(context.Background(), "ws", "P", "toolongkey")
	if err != domain.ErrProjectKeyInvalid {
		t.Fatalf("expected ErrProjectKeyInvalid, got %v", err)
	}

	_, err = svc.CreateProject(context.Background(), "ws", "P", "be-1")
	if err != domain.ErrProjectKeyInvalid {
		t.Fatalf("expected ErrProjectKeyInvalid for lowercase, got %v", err)
	}
}

func TestProjectService_CreateProject_EmptyName(t *testing.T) {
	wsRepo := &mockWorkspaceRepo{
		getBySlugFn: func(ctx context.Context, slug string) (*domain.Workspace, error) {
			return &domain.Workspace{ID: "ws-1"}, nil
		},
	}
	svc := NewProjectService(&mockProjectRepo{}, wsRepo)

	_, err := svc.CreateProject(context.Background(), "ws", "", "BE")
	if err != domain.ErrProjectNameRequired {
		t.Fatalf("expected ErrProjectNameRequired, got %v", err)
	}
}

func TestProjectService_ListProjects(t *testing.T) {
	wsRepo := &mockWorkspaceRepo{
		getBySlugFn: func(ctx context.Context, slug string) (*domain.Workspace, error) {
			return &domain.Workspace{ID: "ws-1"}, nil
		},
	}
	projRepo := &mockProjectRepo{
		listFn: func(ctx context.Context, workspaceID string) ([]*domain.Project, error) {
			return []*domain.Project{
				{ID: "p1", Name: "Backend", Key: "BE"},
				{ID: "p2", Name: "Frontend", Key: "FE"},
			}, nil
		},
	}
	svc := NewProjectService(projRepo, wsRepo)

	ps, err := svc.ListProjects(context.Background(), "my-ws")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(ps) != 2 {
		t.Fatalf("expected 2, got %d", len(ps))
	}
}
