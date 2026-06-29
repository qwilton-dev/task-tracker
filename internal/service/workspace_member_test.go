package service

import (
	"context"
	"testing"

	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
)

type mockWorkspaceMemberRepo struct {
	createFn    func(ctx context.Context, m *domain.WorkspaceMember) error
	listFn      func(ctx context.Context, workspaceId string) ([]*domain.WorkspaceMember, error)
	deleteFn    func(ctx context.Context, m *domain.WorkspaceMember) error
	updateRoleFn func(ctx context.Context, m *domain.WorkspaceMember) error
}

var _ repository.WorkspaceMemberRepository = (*mockWorkspaceMemberRepo)(nil)

func (m *mockWorkspaceMemberRepo) CreateWorkspaceMember(ctx context.Context, wm *domain.WorkspaceMember) error {
	if m.createFn != nil {
		return m.createFn(ctx, wm)
	}
	return nil
}
func (m *mockWorkspaceMemberRepo) GetWorkspaceMembers(ctx context.Context, workspaceId string) ([]*domain.WorkspaceMember, error) {
	if m.listFn != nil {
		return m.listFn(ctx, workspaceId)
	}
	return nil, nil
}
func (m *mockWorkspaceMemberRepo) DeleteWorkspaceMember(ctx context.Context, wm *domain.WorkspaceMember) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, wm)
	}
	return nil
}
func (m *mockWorkspaceMemberRepo) UpdateWorkspaceMemberRole(ctx context.Context, wm *domain.WorkspaceMember) error {
	if m.updateRoleFn != nil {
		return m.updateRoleFn(ctx, wm)
	}
	return nil
}
func (m *mockWorkspaceMemberRepo) GetRole(ctx context.Context, workspaceID, userID string) (string, error) {
	return "", nil
}
func (m *mockWorkspaceMemberRepo) GetRoleByProjectID(ctx context.Context, projectID, userID string) (string, error) {
	return "", nil
}
func (m *mockWorkspaceMemberRepo) GetRoleByIssueID(ctx context.Context, issueID, userID string) (string, error) {
	return "", nil
}

func TestWorkspaceMemberService_AddMember(t *testing.T) {
	var got *domain.WorkspaceMember
	repo := &mockWorkspaceMemberRepo{
		createFn: func(ctx context.Context, m *domain.WorkspaceMember) error {
			got = m
			return nil
		},
	}
	svc := NewWorkspaceMemberService(repo)

	err := svc.AddMember(context.Background(), "ws-1", "user-1", "member")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.WorkspaceId != "ws-1" {
		t.Fatalf("expected ws-1, got %s", got.WorkspaceId)
	}
	if got.UserId != "user-1" {
		t.Fatalf("expected user-1, got %s", got.UserId)
	}
	if got.Role != "member" {
		t.Fatalf("expected member, got %s", got.Role)
	}
}

func TestWorkspaceMemberService_ListMembers(t *testing.T) {
	repo := &mockWorkspaceMemberRepo{
		listFn: func(ctx context.Context, workspaceId string) ([]*domain.WorkspaceMember, error) {
			return []*domain.WorkspaceMember{
				{WorkspaceId: "ws-1", UserId: "u1", Role: "owner"},
				{WorkspaceId: "ws-1", UserId: "u2", Role: "member"},
			}, nil
		},
	}
	svc := NewWorkspaceMemberService(repo)

	members, err := svc.ListMembers(context.Background(), "ws-1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("expected 2, got %d", len(members))
	}
}

func TestWorkspaceMemberService_RemoveMember(t *testing.T) {
	var deleted *domain.WorkspaceMember
	repo := &mockWorkspaceMemberRepo{
		deleteFn: func(ctx context.Context, m *domain.WorkspaceMember) error {
			deleted = m
			return nil
		},
	}
	svc := NewWorkspaceMemberService(repo)

	err := svc.RemoveMember(context.Background(), "ws-1", "user-1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if deleted.UserId != "user-1" {
		t.Fatalf("expected user-1, got %s", deleted.UserId)
	}
}

func TestWorkspaceMemberService_UpdateMemberRole(t *testing.T) {
	var updated *domain.WorkspaceMember
	repo := &mockWorkspaceMemberRepo{
		updateRoleFn: func(ctx context.Context, m *domain.WorkspaceMember) error {
			updated = m
			return nil
		},
	}
	svc := NewWorkspaceMemberService(repo)

	err := svc.UpdateMemberRole(context.Background(), "ws-1", "user-1", "owner")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if updated.Role != "owner" {
		t.Fatalf("expected owner, got %s", updated.Role)
	}
}
