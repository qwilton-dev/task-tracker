package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"task-tracker/internal/domain"
	"task-tracker/internal/http/middleware"
	"task-tracker/internal/service"

	"context"
	"github.com/go-chi/chi/v5"
)

type handlerMemberRepo struct {
	createFn  func(ctx context.Context, m *domain.WorkspaceMember) error
	getRoleFn func(ctx context.Context, workspaceID, userID string) (string, error)
}

func (r *handlerMemberRepo) CreateWorkspaceMember(ctx context.Context, m *domain.WorkspaceMember) error {
	if r.createFn != nil {
		return r.createFn(ctx, m)
	}
	return nil
}
func (r *handlerMemberRepo) GetWorkspaceMembers(ctx context.Context, workspaceId string) ([]*domain.WorkspaceMember, error) {
	return nil, nil
}
func (r *handlerMemberRepo) DeleteWorkspaceMember(ctx context.Context, m *domain.WorkspaceMember) error {
	return nil
}
func (r *handlerMemberRepo) UpdateWorkspaceMemberRole(ctx context.Context, m *domain.WorkspaceMember) error {
	return nil
}
func (r *handlerMemberRepo) GetRole(ctx context.Context, workspaceID, userID string) (string, error) {
	if r.getRoleFn != nil {
		return r.getRoleFn(ctx, workspaceID, userID)
	}
	return "", domain.ErrInviteNotFound
}
func (r *handlerMemberRepo) GetRoleByProjectID(ctx context.Context, projectID, userID string) (string, error) {
	return "", nil
}
func (r *handlerMemberRepo) GetRoleByIssueID(ctx context.Context, issueID, userID string) (string, error) {
	return "", nil
}

type inviteHandlerInviteRepo struct {
	createFn func(ctx context.Context, invite *domain.Invite) error
	listFn   func(ctx context.Context, workspaceID string) ([]*domain.Invite, error)
	getByFn  func(ctx context.Context, token string) (*domain.Invite, error)
	acceptFn func(ctx context.Context, token string) error
}

func (r *inviteHandlerInviteRepo) CreateInvite(ctx context.Context, invite *domain.Invite) error {
	if r.createFn != nil {
		return r.createFn(ctx, invite)
	}
	invite.ID = "invite-1"
	invite.CreatedAt = time.Now()
	return nil
}
func (r *inviteHandlerInviteRepo) GetInviteByToken(ctx context.Context, token string) (*domain.Invite, error) {
	if r.getByFn != nil {
		return r.getByFn(ctx, token)
	}
	return nil, domain.ErrInviteNotFound
}
func (r *inviteHandlerInviteRepo) AcceptInvite(ctx context.Context, token string) error {
	if r.acceptFn != nil {
		return r.acceptFn(ctx, token)
	}
	return nil
}
func (r *inviteHandlerInviteRepo) ListInvitesByWorkspace(ctx context.Context, workspaceID string) ([]*domain.Invite, error) {
	if r.listFn != nil {
		return r.listFn(ctx, workspaceID)
	}
	return nil, nil
}

func TestInviteHandler_Create_201(t *testing.T) {
	svc := service.NewInviteService(&inviteHandlerInviteRepo{}, &handlerMemberRepo{})
	h := NewInviteHandler(svc)

	body := bytes.NewBufferString(`{"email":"a@b.com","role":"member"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/invites", body)
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("workspaceID", "ws-1")
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
	if resp["email"] != "a@b.com" {
		t.Fatalf("expected a@b.com, got %v", resp["email"])
	}
}

func TestInviteHandler_Create_400_InvalidRole(t *testing.T) {
	svc := service.NewInviteService(&inviteHandlerInviteRepo{}, &handlerMemberRepo{})
	h := NewInviteHandler(svc)

	body := bytes.NewBufferString(`{"email":"a@b.com","role":"admin"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/invites", body)
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("workspaceID", "ws-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(middleware.WithUserID(req.Context(), "user-1"))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want 400", rr.Code)
	}
}

func TestInviteHandler_List_200(t *testing.T) {
	svc := service.NewInviteService(&inviteHandlerInviteRepo{
		listFn: func(ctx context.Context, workspaceID string) ([]*domain.Invite, error) {
			return []*domain.Invite{
				{ID: "i1", Email: "a@b.com", Role: "member"},
			}, nil
		},
	}, &handlerMemberRepo{})
	h := NewInviteHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/invites", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("workspaceID", "ws-1")
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

func TestInviteHandler_Accept_200(t *testing.T) {
	svc := service.NewInviteService(&inviteHandlerInviteRepo{
		getByFn: func(ctx context.Context, token string) (*domain.Invite, error) {
			return &domain.Invite{
				ID: "i1", WorkspaceID: "ws-1", Role: "member",
				Token: token, ExpiresAt: time.Now().Add(24 * time.Hour),
			}, nil
		},
	}, &handlerMemberRepo{})
	h := NewInviteHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/invites/tok123/accept", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "tok123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(middleware.WithUserID(req.Context(), "user-2"))
	rr := httptest.NewRecorder()

	h.Accept(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestInviteHandler_Accept_404_NotFound(t *testing.T) {
	svc := service.NewInviteService(&inviteHandlerInviteRepo{}, &handlerMemberRepo{})
	h := NewInviteHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/invites/bad/accept", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "bad")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(middleware.WithUserID(req.Context(), "user-2"))
	rr := httptest.NewRecorder()

	h.Accept(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want 404", rr.Code)
	}
}
