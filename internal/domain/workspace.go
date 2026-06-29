package domain

import (
	"errors"
	"strings"
	"time"
)

type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func NewWorkspace(name string) (*Workspace, error) {
	name = normalizeName(name)
	if name == "" {
		return nil, ErrWorkspaceNameRequired
	}
	return &Workspace{
		Name: name,
	}, nil
}

func normalizeName(name string) string {
	return strings.TrimSpace(name)
}

var (
	ErrWorkspaceNameRequired = errors.New("workspace name is required")
	ErrWorkspaceNotFound     = errors.New("workspace not found")
)
