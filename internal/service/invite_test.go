package service

import (
	"context"
	"testing"
	"time"

	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
)

type mockInviteRepo struct {
	createFn func(ctx context.Context, invite *domain.Invite) error
	getByFn  func(ctx context.Context, token string) (*domain.Invite, error)
	acceptFn func(ctx context.Context, token string) error
	listFn   func(ctx context.Context, workspaceID string) ([]*domain.Invite, error)
}

var _ repository.InviteRepository = (*mockInviteRepo)(nil)

func (m *mockInviteRepo) CreateInvite(ctx context.Context, invite *domain.Invite) error {
	if m.createFn != nil {
		return m.createFn(ctx, invite)
	}
	invite.ID = "invite-1"
	invite.CreatedAt = time.Now()
	return nil
}
func (m *mockInviteRepo) GetInviteByToken(ctx context.Context, token string) (*domain.Invite, error) {
	if m.getByFn != nil {
		return m.getByFn(ctx, token)
	}
	return nil, domain.ErrInviteNotFound
}
func (m *mockInviteRepo) AcceptInvite(ctx context.Context, token string) error {
	if m.acceptFn != nil {
		return m.acceptFn(ctx, token)
	}
	return nil
}
func (m *mockInviteRepo) ListInvitesByWorkspace(ctx context.Context, workspaceID string) ([]*domain.Invite, error) {
	if m.listFn != nil {
		return m.listFn(ctx, workspaceID)
	}
	return nil, nil
}

type mockMemberRepo struct {
	createFn  func(ctx context.Context, m *domain.WorkspaceMember) error
	getRoleFn func(ctx context.Context, workspaceID, userID string) (string, error)
}

var _ repository.WorkspaceMemberRepository = (*mockMemberRepo)(nil)

func (m *mockMemberRepo) CreateWorkspaceMember(ctx context.Context, wm *domain.WorkspaceMember) error {
	if m.createFn != nil {
		return m.createFn(ctx, wm)
	}
	return nil
}
func (m *mockMemberRepo) GetWorkspaceMembers(ctx context.Context, workspaceId string) ([]*domain.WorkspaceMember, error) {
	return nil, nil
}
func (m *mockMemberRepo) DeleteWorkspaceMember(ctx context.Context, wm *domain.WorkspaceMember) error {
	return nil
}
func (m *mockMemberRepo) UpdateWorkspaceMemberRole(ctx context.Context, wm *domain.WorkspaceMember) error {
	return nil
}
func (m *mockMemberRepo) GetRole(ctx context.Context, workspaceID, userID string) (string, error) {
	if m.getRoleFn != nil {
		return m.getRoleFn(ctx, workspaceID, userID)
	}
	return "", domain.ErrInviteNotFound
}
func (m *mockMemberRepo) GetRoleByProjectID(ctx context.Context, projectID, userID string) (string, error) {
	return "", nil
}
func (m *mockMemberRepo) GetRoleByIssueID(ctx context.Context, issueID, userID string) (string, error) {
	return "", nil
}

func TestInviteService_CreateInvite(t *testing.T) {
	inviteRepo := &mockInviteRepo{}
	memberRepo := &mockMemberRepo{}
	svc := NewInviteService(inviteRepo, memberRepo)

	invite, err := svc.CreateInvite(context.Background(), "ws-1", "a@b.com", "member", "user-1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if invite.Email != "a@b.com" {
		t.Fatalf("expected a@b.com, got %s", invite.Email)
	}
	if invite.Role != "member" {
		t.Fatalf("expected member, got %s", invite.Role)
	}
	if invite.Token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestInviteService_CreateInvite_InvalidRole(t *testing.T) {
	svc := NewInviteService(&mockInviteRepo{}, &mockMemberRepo{})
	_, err := svc.CreateInvite(context.Background(), "ws-1", "a@b.com", "admin", "user-1")
	if err != domain.ErrInviteInvalidRole {
		t.Fatalf("expected ErrInviteInvalidRole, got %v", err)
	}
}

func TestInviteService_CreateInvite_EmptyEmail(t *testing.T) {
	svc := NewInviteService(&mockInviteRepo{}, &mockMemberRepo{})
	_, err := svc.CreateInvite(context.Background(), "ws-1", "", "member", "user-1")
	if err != domain.ErrInviteEmailRequired {
		t.Fatalf("expected ErrInviteEmailRequired, got %v", err)
	}
}

func TestInviteService_AcceptInvite(t *testing.T) {
	memberCreated := false
	inviteRepo := &mockInviteRepo{
		getByFn: func(ctx context.Context, token string) (*domain.Invite, error) {
			return &domain.Invite{
				ID:          "invite-1",
				WorkspaceID: "ws-1",
				Email:       "a@b.com",
				Role:        "member",
				Token:       token,
				ExpiresAt:   time.Now().Add(24 * time.Hour),
				CreatedBy:   "user-1",
			}, nil
		},
	}
	memberRepo := &mockMemberRepo{
		createFn: func(ctx context.Context, m *domain.WorkspaceMember) error {
			memberCreated = true
			return nil
		},
	}
	svc := NewInviteService(inviteRepo, memberRepo)

	err := svc.AcceptInvite(context.Background(), "tok123", "user-2")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !memberCreated {
		t.Fatal("expected member to be created")
	}
}

func TestInviteService_AcceptInvite_AlreadyAccepted(t *testing.T) {
	inviteRepo := &mockInviteRepo{
		getByFn: func(ctx context.Context, token string) (*domain.Invite, error) {
			now := time.Now()
			return &domain.Invite{
				ID:          "invite-1",
				WorkspaceID: "ws-1",
				Email:       "a@b.com",
				Role:        "member",
				Token:       token,
				ExpiresAt:   time.Now().Add(24 * time.Hour),
				AcceptedAt:  &now,
				CreatedBy:   "user-1",
			}, nil
		},
	}
	svc := NewInviteService(inviteRepo, &mockMemberRepo{})

	err := svc.AcceptInvite(context.Background(), "tok123", "user-2")
	if err != domain.ErrInviteAlreadyAccepted {
		t.Fatalf("expected ErrInviteAlreadyAccepted, got %v", err)
	}
}

func TestInviteService_AcceptInvite_Expired(t *testing.T) {
	inviteRepo := &mockInviteRepo{
		getByFn: func(ctx context.Context, token string) (*domain.Invite, error) {
			return &domain.Invite{
				ID:          "invite-1",
				WorkspaceID: "ws-1",
				Email:       "a@b.com",
				Role:        "member",
				Token:       token,
				ExpiresAt:   time.Now().Add(-24 * time.Hour),
				CreatedBy:   "user-1",
			}, nil
		},
	}
	svc := NewInviteService(inviteRepo, &mockMemberRepo{})

	err := svc.AcceptInvite(context.Background(), "tok123", "user-2")
	if err != domain.ErrInviteExpired {
		t.Fatalf("expected ErrInviteExpired, got %v", err)
	}
}

func TestInviteService_AcceptInvite_AlreadyMember(t *testing.T) {
	inviteRepo := &mockInviteRepo{
		getByFn: func(ctx context.Context, token string) (*domain.Invite, error) {
			return &domain.Invite{
				ID:          "invite-1",
				WorkspaceID: "ws-1",
				Email:       "a@b.com",
				Role:        "member",
				Token:       token,
				ExpiresAt:   time.Now().Add(24 * time.Hour),
				CreatedBy:   "user-1",
			}, nil
		},
	}
	memberRepo := &mockMemberRepo{
		getRoleFn: func(ctx context.Context, workspaceID, userID string) (string, error) {
			return "member", nil
		},
	}
	svc := NewInviteService(inviteRepo, memberRepo)

	err := svc.AcceptInvite(context.Background(), "tok123", "user-2")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestInviteService_ListInvites(t *testing.T) {
	inviteRepo := &mockInviteRepo{
		listFn: func(ctx context.Context, workspaceID string) ([]*domain.Invite, error) {
			return []*domain.Invite{
				{ID: "i1", Email: "a@b.com"},
				{ID: "i2", Email: "c@d.com"},
			}, nil
		},
	}
	svc := NewInviteService(inviteRepo, &mockMemberRepo{})

	invites, err := svc.ListInvites(context.Background(), "ws-1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(invites) != 2 {
		t.Fatalf("expected 2, got %d", len(invites))
	}
}
