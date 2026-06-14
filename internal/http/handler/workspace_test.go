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
	"task-tracker/internal/http/middleware"
	"task-tracker/internal/service"
)

type wsHandlerRepo struct {
	createFn    func(ctx context.Context, ws *domain.Workspace) error
	getByIDFn   func(ctx context.Context, id string) (*domain.Workspace, error)
	getBySlugFn func(ctx context.Context, slug string) (*domain.Workspace, error)
	listFn      func(ctx context.Context, userID string) ([]*domain.Workspace, error)
	addMemberFn func(ctx context.Context, workspaceID, userID, role string) error
}

func (r *wsHandlerRepo) CreateWorkspace(ctx context.Context, ws *domain.Workspace) error {
	if r.createFn != nil {
		return r.createFn(ctx, ws)
	}
	ws.ID = "ws-1"
	ws.CreatedAt = time.Now()
	return nil
}
func (r *wsHandlerRepo) GetWorkspaceByID(ctx context.Context, id string) (*domain.Workspace, error) {
	if r.getByIDFn != nil {
		return r.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (r *wsHandlerRepo) GetWorkspaceBySlug(ctx context.Context, slug string) (*domain.Workspace, error) {
	if r.getBySlugFn != nil {
		return r.getBySlugFn(ctx, slug)
	}
	return nil, nil
}
func (r *wsHandlerRepo) UpdateWorkspace(ctx context.Context, ws *domain.Workspace) error {
	return nil
}
func (r *wsHandlerRepo) DeleteWorkspace(ctx context.Context, id string) error { return nil }
func (r *wsHandlerRepo) ListWorkspaces(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	if r.listFn != nil {
		return r.listFn(ctx, userID)
	}
	return nil, nil
}
func (r *wsHandlerRepo) AddMember(ctx context.Context, workspaceID, userID, role string) error {
	if r.addMemberFn != nil {
		return r.addMemberFn(ctx, workspaceID, userID, role)
	}
	return nil
}

func TestWorkspaceHandler_Create_201(t *testing.T) {
	repo := &wsHandlerRepo{}
	svc := service.NewWorkspaceService(repo)
	h := NewWorkspaceHandler(svc)

	body := bytes.NewBufferString(`{"name":"My WS","slug":"my-ws"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces", body)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithUserID(req.Context(), "user-1"))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status: got %d body=%s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp["name"] != "My WS" {
		t.Fatalf("expected My WS, got %v", resp["name"])
	}
}

func TestWorkspaceHandler_Create_401_NoUserID(t *testing.T) {
	repo := &wsHandlerRepo{}
	svc := service.NewWorkspaceService(repo)
	h := NewWorkspaceHandler(svc)

	body := bytes.NewBufferString(`{"name":"My WS","slug":"my-ws"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces", body)
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status: got %d, want 401", rr.Code)
	}
}

func TestWorkspaceHandler_Create_400_BadJSON(t *testing.T) {
	repo := &wsHandlerRepo{}
	svc := service.NewWorkspaceService(repo)
	h := NewWorkspaceHandler(svc)

	body := bytes.NewBufferString(`not json`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces", body)
	req = req.WithContext(middleware.WithUserID(req.Context(), "user-1"))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want 400", rr.Code)
	}
}

func TestWorkspaceHandler_List_200(t *testing.T) {
	repo := &wsHandlerRepo{
		listFn: func(ctx context.Context, userID string) ([]*domain.Workspace, error) {
			return []*domain.Workspace{
				{ID: "ws-1", Name: "WS1", Slug: "ws1"},
				{ID: "ws-2", Name: "WS2", Slug: "ws2"},
			}, nil
		},
	}
	svc := service.NewWorkspaceService(repo)
	h := NewWorkspaceHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), "user-1"))
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}
	var resp []map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("expected 2, got %d", len(resp))
	}
}
