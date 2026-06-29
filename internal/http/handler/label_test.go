package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"task-tracker/internal/domain"
	"task-tracker/internal/service"

	"github.com/go-chi/chi/v5"
	"context"
)

type labelHandlerLabelRepo struct {
	createFn      func(ctx context.Context, label *domain.Label) error
	listByWsFn    func(ctx context.Context, workspaceID string) ([]*domain.Label, error)
	listByIssueFn func(ctx context.Context, issueID string) ([]*domain.Label, error)
	attachFn      func(ctx context.Context, issueID, labelID string) error
	detachFn      func(ctx context.Context, issueID, labelID string) error
}

func (r *labelHandlerLabelRepo) CreateLabel(ctx context.Context, label *domain.Label) error {
	if r.createFn != nil {
		return r.createFn(ctx, label)
	}
	label.ID = "label-1"
	return nil
}
func (r *labelHandlerLabelRepo) GetLabelByID(ctx context.Context, id string) (*domain.Label, error) {
	return nil, nil
}
func (r *labelHandlerLabelRepo) ListLabelsByWorkspace(ctx context.Context, workspaceID string) ([]*domain.Label, error) {
	if r.listByWsFn != nil {
		return r.listByWsFn(ctx, workspaceID)
	}
	return nil, nil
}
func (r *labelHandlerLabelRepo) UpdateLabel(ctx context.Context, label *domain.Label) error { return nil }
func (r *labelHandlerLabelRepo) DeleteLabel(ctx context.Context, id string) error          { return nil }
func (r *labelHandlerLabelRepo) AttachLabel(ctx context.Context, issueID, labelID string) error {
	if r.attachFn != nil {
		return r.attachFn(ctx, issueID, labelID)
	}
	return nil
}
func (r *labelHandlerLabelRepo) DetachLabel(ctx context.Context, issueID, labelID string) error {
	if r.detachFn != nil {
		return r.detachFn(ctx, issueID, labelID)
	}
	return nil
}
func (r *labelHandlerLabelRepo) ListLabelsByIssue(ctx context.Context, issueID string) ([]*domain.Label, error) {
	if r.listByIssueFn != nil {
		return r.listByIssueFn(ctx, issueID)
	}
	return nil, nil
}

func TestLabelHandler_CreateLabel_201(t *testing.T) {
	svc := service.NewLabelService(&labelHandlerLabelRepo{}, service.NewActivityEventService(&handlerMockActivityEventRepo{}))
	h := NewLabelHandler(svc)

	body := bytes.NewBufferString(`{"name":"Bug","color":"#ff0000"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/labels", body)
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("workspaceID", "ws-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.CreateLabel(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status: got %d body=%s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp["name"] != "Bug" {
		t.Fatalf("expected Bug, got %v", resp["name"])
	}
}

func TestLabelHandler_CreateLabel_400_EmptyName(t *testing.T) {
	svc := service.NewLabelService(&labelHandlerLabelRepo{}, service.NewActivityEventService(&handlerMockActivityEventRepo{}))
	h := NewLabelHandler(svc)

	body := bytes.NewBufferString(`{"name":"","color":"#ff0000"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/labels", body)
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("workspaceID", "ws-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.CreateLabel(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want 400", rr.Code)
	}
}

func TestLabelHandler_ListLabels_200(t *testing.T) {
	svc := service.NewLabelService(&labelHandlerLabelRepo{
		listByWsFn: func(ctx context.Context, workspaceID string) ([]*domain.Label, error) {
			return []*domain.Label{
				{ID: "l1", Name: "Bug", Color: "#ff0000"},
				{ID: "l2", Name: "Feature", Color: "#00ff00"},
			}, nil
		},
	}, service.NewActivityEventService(&handlerMockActivityEventRepo{}))
	h := NewLabelHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/labels", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("workspaceID", "ws-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.ListLabels(rr, req)

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

func TestLabelHandler_AttachLabel_200(t *testing.T) {
	svc := service.NewLabelService(&labelHandlerLabelRepo{}, service.NewActivityEventService(&handlerMockActivityEventRepo{}))
	h := NewLabelHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/issues/issue-1/labels/label-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("issueID", "issue-1")
	rctx.URLParams.Add("labelID", "label-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.AttachLabel(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestLabelHandler_DetachLabel_200(t *testing.T) {
	svc := service.NewLabelService(&labelHandlerLabelRepo{}, service.NewActivityEventService(&handlerMockActivityEventRepo{}))
	h := NewLabelHandler(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/issues/issue-1/labels/label-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("issueID", "issue-1")
	rctx.URLParams.Add("labelID", "label-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.DetachLabel(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%s", rr.Code, rr.Body.String())
	}
}
