package domain

import (
	"errors"
	"strings"
	"time"
)

type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

func NewWorkspace(name, slug string) (*Workspace, error) {
	name = normalizeName(name)
	slug = normalizeSlug(slug)
	if name == "" {
		return nil, ErrWorkspaceNameRequired
	}
	if slug == "" {
		return nil, ErrWorkspaceSlugRequired
	}
	return &Workspace{
		Name: name,
		Slug: slug,
	}, nil
}

func normalizeName(name string) string {
	return strings.TrimSpace(name)
}

func normalizeSlug(slug string) string {
	return strings.TrimSpace(slug)
}

var (
	ErrWorkspaceNameRequired = errors.New("workspace name is required")
	ErrWorkspaceSlugRequired = errors.New("workspace slug is required")
)
