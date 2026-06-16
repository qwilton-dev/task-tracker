package repository

import (
	"context"
	"task-tracker/internal/domain"
)

type InviteRepository interface {
	CreateInvite(ctx context.Context, invite *domain.Invite) error
	GetInviteByToken(ctx context.Context, token string) (*domain.Invite, error)
	AcceptInvite(ctx context.Context, token string) error
	ListInvitesByWorkspace(ctx context.Context, workspaceID string) ([]*domain.Invite, error)
}
