package service

import (
	"context"
	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
)

type WorkspaceMemberService struct {
	repo repository.WorkspaceMemberRepository
}

func NewWorkspaceMemberService(repo repository.WorkspaceMemberRepository) *WorkspaceMemberService {
	return &WorkspaceMemberService{repo: repo}
}

func (s *WorkspaceMemberService) AddMember(ctx context.Context, workspaceId, userId, role string) error {
	member := &domain.WorkspaceMember{
		WorkspaceId: workspaceId,
		UserId:      userId,
		Role:        role,
	}
	return s.repo.CreateWorkspaceMember(ctx, member)
}

func (s *WorkspaceMemberService) ListMembers(ctx context.Context, workspaceId string) ([]*domain.WorkspaceMember, error) {
	return s.repo.GetWorkspaceMembers(ctx, workspaceId) // TODO: List group by role
}

func (s *WorkspaceMemberService) RemoveMember(ctx context.Context, workspaceId, userId string) error {
	member := &domain.WorkspaceMember{
		WorkspaceId: workspaceId,
		UserId:      userId,
	}
	return s.repo.DeleteWorkspaceMember(ctx, member)
}

func (s *WorkspaceMemberService) UpdateMemberRole(ctx context.Context, workspaceId, userId, role string) error {
	member := &domain.WorkspaceMember{
		WorkspaceId: workspaceId,
		UserId:      userId,
		Role:        role,
	}
	return s.repo.UpdateWorkspaceMemberRole(ctx, member)
}
