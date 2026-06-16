package domain

import (
	"errors"
	"time"
)

type Invite struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspace_id"`
	Email       string     `json:"email"`
	Role        string     `json:"role"`
	Token       string     `json:"-"`
	ExpiresAt   time.Time  `json:"expires_at"`
	AcceptedAt  *time.Time `json:"accepted_at,omitempty"`
	CreatedBy   string     `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
}

var (
	ErrInviteEmailRequired    = errors.New("invite email is required")
	ErrInviteRoleRequired     = errors.New("invite role is required")
	ErrInviteInvalidRole      = errors.New("invite role must be 'member' or 'viewer'")
	ErrInviteNotFound         = errors.New("invite not found")
	ErrInviteExpired          = errors.New("invite has expired")
	ErrInviteAlreadyAccepted  = errors.New("invite already accepted")
	ErrInviteTokenRequired    = errors.New("invite token is required")
)

func NewInvite(workspaceID, email, role, token, createdBy string, expiresAt time.Time) (*Invite, error) {
	if email == "" {
		return nil, ErrInviteEmailRequired
	}
	if role == "" {
		return nil, ErrInviteRoleRequired
	}
	if role != "member" && role != "viewer" {
		return nil, ErrInviteInvalidRole
	}
	return &Invite{
		WorkspaceID: workspaceID,
		Email:       email,
		Role:        role,
		Token:       token,
		ExpiresAt:   expiresAt,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
	}, nil
}
