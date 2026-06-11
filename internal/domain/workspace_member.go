package domain

import (
	"errors"
	"strings"
)

type WorkspaceMember struct {
	WorkspaceId string
	UserId      string
	Role        string
}

func NewWorkspaceMember(workspaceId, userId, role string) (*WorkspaceMember, error) {

	role = normalizeWorkspaceMemberName(role)
	if workspaceId == "" {
		return nil, ErrWorkspaceIdRequired
	}
	if userId == "" {
		return nil, ErrWorkspaceMemberIdRequired
	}
	if role == "" {
		return nil, ErrWorkspaceMemberRoleRequired
	}
	return &WorkspaceMember{
		WorkspaceId: workspaceId,
		UserId:      userId,
		Role:        role,
	}, nil
}

func normalizeWorkspaceMemberName(name string) string {
	return strings.TrimSpace(name)
}

var (
	ErrWorkspaceMemberIdRequired   = errors.New("member id is required")
	ErrWorkspaceIdRequired         = errors.New("workspace id is required")
	ErrWorkspaceMemberRoleRequired = errors.New("member role is required")
)
