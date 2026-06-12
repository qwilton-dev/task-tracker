package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"task-tracker/internal/domain"
	"task-tracker/internal/service"

	"github.com/go-chi/chi/v5"
)

type projHandlerProjectRepo struct {
	createFn func(ctx context.Context, p *domain.Project) error
	listFn   func(ctx context.Context, workspaceID string) ([]*domain.Project, error)
}

func (r *projHandlerProjectRepo) CreateProject(ctx context.Context, p *domain.Project) error {
	if r.createFn != nil {
		return r.createFn(ctx, p)
	}
	p.ID = "proj-1"
	p.CreatedAt = time.Now()
	return nil
}
func (r *projHandlerProjectRepo) GetProjectsByWorkspace(ctx context.Context, workspaceID string) ([]*domain.Project, error) {
	if r.listFn != nil {
		return r.listFn(ctx, workspaceID)
	}
	return nil, nil
}
func (r *projHandlerProjectRepo) ExistsByKey(ctx context.Context, workspaceID, key string) (bool, error) {
	return false, nil
}

func TestProjectHandler_Create_201(t *testing.T) {
	wsRepo := &wsHandlerRepo{
		getBySlugFn: func(ctx context.Context, slug string) (*domain.Workspace, error) {
			return &domain.Workspace{ID: "ws-1", Slug: slug}, nil
		},
	}
	projRepo := &projHandlerProjectRepo{}
	svc := service.NewProjectService(projRepo, wsRepo)
	h := NewProjectHandler(svc)

	body := bytes.NewBufferString(`{"name":"Backend","key":"BE"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/my-ws/projects", body)
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("workspaceSlug", "my-ws")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status: got %d body=%s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp["key"] != "BE" {
		t.Fatalf("expected BE, got %v", resp["key"])
	}
}

func TestProjectHandler_Create_400_BadJSON(t *testing.T) {
	wsRepo := &wsHandlerRepo{}
	projRepo := &projHandlerProjectRepo{}
	svc := service.NewProjectService(projRepo, wsRepo)
	h := NewProjectHandler(svc)

	body := bytes.NewBufferString(`not json`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/my-ws/projects", body)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("workspaceSlug", "my-ws")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want 400", rr.Code)
	}
}

func TestProjectHandler_List_200(t *testing.T) {
	wsRepo := &wsHandlerRepo{
		getBySlugFn: func(ctx context.Context, slug string) (*domain.Workspace, error) {
			return &domain.Workspace{ID: "ws-1"}, nil
		},
	}
	projRepo := &projHandlerProjectRepo{
		listFn: func(ctx context.Context, workspaceID string) ([]*domain.Project, error) {
			return []*domain.Project{
				{ID: "p1", Name: "Backend", Key: "BE"},
			}, nil
		},
	}
	svc := service.NewProjectService(projRepo, wsRepo)
	h := NewProjectHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/my-ws/projects", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("workspaceSlug", "my-ws")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}
	var resp []map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1, got %d", len(resp))
	}
}
