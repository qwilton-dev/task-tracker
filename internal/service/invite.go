package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
)

type InviteService struct {
	inviteRepo    repository.InviteRepository
	memberRepo    repository.WorkspaceMemberRepository
}

func NewInviteService(inviteRepo repository.InviteRepository, memberRepo repository.WorkspaceMemberRepository) *InviteService {
	return &InviteService{inviteRepo: inviteRepo, memberRepo: memberRepo}
}

func (s *InviteService) CreateInvite(ctx context.Context, workspaceID, email, role, createdBy string) (*domain.Invite, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	invite, err := domain.NewInvite(workspaceID, email, role, token, createdBy, expiresAt)
	if err != nil {
		return nil, err
	}
	if err := s.inviteRepo.CreateInvite(ctx, invite); err != nil {
		return nil, err
	}
	return invite, nil
}

func (s *InviteService) AcceptInvite(ctx context.Context, token, userID string) error {
	invite, err := s.inviteRepo.GetInviteByToken(ctx, token)
	if err != nil {
		return domain.ErrInviteNotFound
	}
	if invite.AcceptedAt != nil {
		return domain.ErrInviteAlreadyAccepted
	}
	if time.Now().After(invite.ExpiresAt) {
		return domain.ErrInviteExpired
	}
	if err := s.inviteRepo.AcceptInvite(ctx, token); err != nil {
		return err
	}
	return s.memberRepo.CreateWorkspaceMember(ctx, &domain.WorkspaceMember{
		WorkspaceId: invite.WorkspaceID,
		UserId:      userID,
		Role:        invite.Role,
	})
}

func (s *InviteService) ListInvites(ctx context.Context, workspaceID string) ([]*domain.Invite, error) {
	return s.inviteRepo.ListInvitesByWorkspace(ctx, workspaceID)
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
