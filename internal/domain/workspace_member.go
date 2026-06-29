package domain

import (
	"errors"
	"strings"
)

type WorkspaceMember struct {
	WorkspaceId string `json:"workspace_id"`
	UserId      string `json:"user_id"`
	Role        string `json:"role"`
	UserName    string `json:"user_name,omitempty"`
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
	if !isValidRole(role) {
		return nil, ErrWorkspaceMemberInvalidRole
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

func isValidRole(role string) bool {
	return role == "owner" || role == "member" || role == "viewer"
}

var (
	ErrWorkspaceMemberIdRequired   = errors.New("member id is required")
	ErrWorkspaceIdRequired         = errors.New("workspace id is required")
	ErrWorkspaceMemberRoleRequired = errors.New("member role is required")
	ErrWorkspaceMemberInvalidRole  = errors.New("role must be 'owner', 'member', or 'viewer'")
)
