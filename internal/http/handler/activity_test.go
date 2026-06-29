package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"task-tracker/internal/domain"
	"task-tracker/internal/service"

	"context"
	"github.com/go-chi/chi/v5"
)

type activityHandlerRepo struct {
	listFn func(ctx context.Context, issueID string) ([]*domain.ActivityEvent, error)
}

func (r *activityHandlerRepo) Create(ctx context.Context, event *domain.ActivityEvent) error {
	return nil
}
func (r *activityHandlerRepo) ListByIssue(ctx context.Context, issueID string) ([]*domain.ActivityEvent, error) {
	if r.listFn != nil {
		return r.listFn(ctx, issueID)
	}
	return nil, nil
}

func TestActivityHandler_ListByIssue_200(t *testing.T) {
	svc := service.NewActivityEventService(&activityHandlerRepo{
		listFn: func(ctx context.Context, issueID string) ([]*domain.ActivityEvent, error) {
			return []*domain.ActivityEvent{
				{ID: "e1", IssueID: issueID, Type: "issue.created", CreatedAt: time.Now()},
				{ID: "e2", IssueID: issueID, Type: "issue.moved", CreatedAt: time.Now()},
			}, nil
		},
	})
	h := NewActivityHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues/issue-1/activity", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("issueID", "issue-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.ListByIssue(rr, req)

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

func TestActivityHandler_ListByIssue_Empty(t *testing.T) {
	svc := service.NewActivityEventService(&activityHandlerRepo{})
	h := NewActivityHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues/issue-1/activity", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("issueID", "issue-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.ListByIssue(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}
}
