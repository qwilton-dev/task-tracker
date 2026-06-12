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
	"task-tracker/internal/repository"
	"task-tracker/internal/service"

	"github.com/go-chi/chi/v5"
)

type issueHandlerRepo struct {
	createFn  func(ctx context.Context, issue *domain.Issue) error
	getByIDFn func(ctx context.Context, id string) (*domain.Issue, error)
	listFn    func(ctx context.Context, projectID string, filters repository.IssueFilters) ([]*domain.Issue, error)
	deleteFn  func(ctx context.Context, id string) error
	moveFn    func(ctx context.Context, id, status string, position float64) error
	maxNumFn  func(ctx context.Context, projectID string) (int, error)
}

var _ repository.IssueRepository = (*issueHandlerRepo)(nil)

func (r *issueHandlerRepo) CreateIssue(ctx context.Context, issue *domain.Issue) error {
	if r.createFn != nil {
		return r.createFn(ctx, issue)
	}
	issue.ID = "issue-1"
	issue.CreatedAt = time.Now()
	issue.UpdatedAt = time.Now()
	return nil
}
func (r *issueHandlerRepo) CreateIssueTx(ctx context.Context, issue *domain.Issue) error {
	return r.CreateIssue(ctx, issue)
}
func (r *issueHandlerRepo) GetIssueByID(ctx context.Context, id string) (*domain.Issue, error) {
	if r.getByIDFn != nil {
		return r.getByIDFn(ctx, id)
	}
	return nil, domain.ErrIssueNotFound
}
func (r *issueHandlerRepo) ListIssuesByProject(ctx context.Context, projectID string, filters repository.IssueFilters) ([]*domain.Issue, error) {
	if r.listFn != nil {
		return r.listFn(ctx, projectID, filters)
	}
	return nil, nil
}
func (r *issueHandlerRepo) UpdateIssue(ctx context.Context, issue *domain.Issue) error {
	return nil
}
func (r *issueHandlerRepo) DeleteIssue(ctx context.Context, id string) error {
	if r.deleteFn != nil {
		return r.deleteFn(ctx, id)
	}
	return nil
}
func (r *issueHandlerRepo) MoveIssue(ctx context.Context, id, status string, position float64) error {
	if r.moveFn != nil {
		return r.moveFn(ctx, id, status, position)
	}
	return nil
}
func (r *issueHandlerRepo) GetMaxNumber(ctx context.Context, projectID string) (int, error) {
	if r.maxNumFn != nil {
		return r.maxNumFn(ctx, projectID)
	}
	return 0, nil
}

func issueCtx(r *http.Request, projectID, issueID string) context.Context {
	rctx := chi.NewRouteContext()
	if projectID != "" {
		rctx.URLParams.Add("projectID", projectID)
	}
	if issueID != "" {
		rctx.URLParams.Add("issueID", issueID)
	}
	return context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
}

func TestIssueHandler_Create_201(t *testing.T) {
	issueRepo := &issueHandlerRepo{
		maxNumFn: func(ctx context.Context, projectID string) (int, error) { return 0, nil },
	}
	projRepo := &projHandlerProjectRepo{}
	svc := service.NewIssueService(issueRepo, projRepo)
	h := NewIssueHandler(svc)

	body := bytes.NewBufferString(`{"title":"Fix bug","description":"desc"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/proj-1/issues", body)
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(issueCtx(req, "proj-1", ""), "user_id", "user-1")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status: got %d body=%s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp["title"] != "Fix bug" {
		t.Fatalf("expected Fix bug, got %v", resp["title"])
	}
}

func TestIssueHandler_Create_400_EmptyTitle(t *testing.T) {
	issueRepo := &issueHandlerRepo{}
	projRepo := &projHandlerProjectRepo{}
	svc := service.NewIssueService(issueRepo, projRepo)
	h := NewIssueHandler(svc)

	body := bytes.NewBufferString(`{"title":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/proj-1/issues", body)
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(issueCtx(req, "proj-1", ""), "user_id", "user-1")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want 400", rr.Code)
	}
}

func TestIssueHandler_List_200(t *testing.T) {
	issueRepo := &issueHandlerRepo{
		listFn: func(ctx context.Context, projectID string, filters repository.IssueFilters) ([]*domain.Issue, error) {
			return []*domain.Issue{
				{ID: "i1", Title: "Bug 1", Status: "todo"},
				{ID: "i2", Title: "Bug 2", Status: "done"},
			}, nil
		},
	}
	projRepo := &projHandlerProjectRepo{}
	svc := service.NewIssueService(issueRepo, projRepo)
	h := NewIssueHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/proj-1/issues", nil)
	req = req.WithContext(issueCtx(req, "proj-1", ""))
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

func TestIssueHandler_Move_200(t *testing.T) {
	issueRepo := &issueHandlerRepo{}
	projRepo := &projHandlerProjectRepo{}
	svc := service.NewIssueService(issueRepo, projRepo)
	h := NewIssueHandler(svc)

	body := bytes.NewBufferString(`{"status":"todo","position":100}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/issues/issue-1/move", body)
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(issueCtx(req, "", "issue-1"), "user_id", "user-1")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.Move(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestIssueHandler_Delete_204(t *testing.T) {
	issueRepo := &issueHandlerRepo{
		deleteFn: func(ctx context.Context, id string) error { return nil },
	}
	projRepo := &projHandlerProjectRepo{}
	svc := service.NewIssueService(issueRepo, projRepo)
	h := NewIssueHandler(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/issues/issue-1", nil)
	ctx := context.WithValue(issueCtx(req, "", "issue-1"), "user_id", "user-1")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.Delete(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status: got %d, want 204", rr.Code)
	}
}
