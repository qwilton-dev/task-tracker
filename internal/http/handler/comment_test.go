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
	"task-tracker/internal/repository"
	"task-tracker/internal/service"

	"github.com/go-chi/chi/v5"
)

type commentHandlerCommentRepo struct {
	createFn func(ctx context.Context, comment *domain.Comment) error
	listFn   func(ctx context.Context, issueID string) ([]*domain.Comment, error)
}

func (r *commentHandlerCommentRepo) CreateComment(ctx context.Context, comment *domain.Comment) error {
	if r.createFn != nil {
		return r.createFn(ctx, comment)
	}
	comment.ID = "comment-1"
	comment.CreatedAt = time.Now()
	return nil
}
func (r *commentHandlerCommentRepo) ListCommentsByIssue(ctx context.Context, issueID string) ([]*domain.Comment, error) {
	if r.listFn != nil {
		return r.listFn(ctx, issueID)
	}
	return nil, nil
}

type commentHandlerIssueRepo struct {
	getByIDFn func(ctx context.Context, id string) (*domain.Issue, error)
}

func (r *commentHandlerIssueRepo) CreateIssue(ctx context.Context, issue *domain.Issue) error {
	return nil
}
func (r *commentHandlerIssueRepo) CreateIssueTx(ctx context.Context, issue *domain.Issue) error {
	return nil
}
func (r *commentHandlerIssueRepo) GetIssueByID(ctx context.Context, id string) (*domain.Issue, error) {
	if r.getByIDFn != nil {
		return r.getByIDFn(ctx, id)
	}
	return nil, domain.ErrIssueNotFound
}
func (r *commentHandlerIssueRepo) ListIssuesByProject(ctx context.Context, projectID string, filters repository.IssueFilters) ([]*domain.Issue, error) {
	return nil, nil
}
func (r *commentHandlerIssueRepo) UpdateIssue(ctx context.Context, issue *domain.Issue) error {
	return nil
}
func (r *commentHandlerIssueRepo) DeleteIssue(ctx context.Context, id string) error { return nil }
func (r *commentHandlerIssueRepo) MoveIssue(ctx context.Context, id, status string, position float64) error {
	return nil
}
func (r *commentHandlerIssueRepo) GetMaxNumber(ctx context.Context, projectID string) (int, error) {
	return 0, nil
}

func TestCommentHandler_Create_201(t *testing.T) {
	svc := service.NewCommentService(
		&commentHandlerCommentRepo{},
		&commentHandlerIssueRepo{getByIDFn: func(ctx context.Context, id string) (*domain.Issue, error) {
			return &domain.Issue{ID: id, ProjectID: "proj-1"}, nil
		}},
		service.NewActivityEventService(&handlerMockActivityEventRepo{}),
		nil,
	)
	h := NewCommentHandler(svc)

	body := bytes.NewBufferString(`{"body":"Great!"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/issues/issue-1/comments", body)
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("issueID", "issue-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
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
	if resp["body"] != "Great!" {
		t.Fatalf("expected Great!, got %v", resp["body"])
	}
}

func TestCommentHandler_Create_400_EmptyBody(t *testing.T) {
	svc := service.NewCommentService(
		&commentHandlerCommentRepo{},
		&commentHandlerIssueRepo{},
		service.NewActivityEventService(&handlerMockActivityEventRepo{}),
		nil,
	)
	h := NewCommentHandler(svc)

	body := bytes.NewBufferString(`{"body":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/issues/issue-1/comments", body)
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("issueID", "issue-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(middleware.WithUserID(req.Context(), "user-1"))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want 400", rr.Code)
	}
}

func TestCommentHandler_List_200(t *testing.T) {
	svc := service.NewCommentService(
		&commentHandlerCommentRepo{
			listFn: func(ctx context.Context, issueID string) ([]*domain.Comment, error) {
				return []*domain.Comment{
					{ID: "c1", Body: "First"},
					{ID: "c2", Body: "Second"},
				}, nil
			},
		},
		&commentHandlerIssueRepo{},
		service.NewActivityEventService(&handlerMockActivityEventRepo{}),
		nil,
	)
	h := NewCommentHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues/issue-1/comments", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("issueID", "issue-1")
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
	if len(resp) != 2 {
		t.Fatalf("expected 2, got %d", len(resp))
	}
}
