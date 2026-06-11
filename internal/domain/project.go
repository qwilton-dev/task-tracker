package domain

import (
	"errors"
	"time"
)

type Project struct {
	ID          string
	WorkspaceID string
	Name        string
	Slug        string
	CreatedAt   time.Time
}

var (
	ErrProjectNameRequired = errors.New("project name is required")
	ErrProjectSlugRequired = errors.New("project slug is required")
)

func NewProject(workspaceID, name, slug string) (*Project, error) {
	if name == "" {
		return nil, ErrProjectNameRequired
	}
	if slug == "" {
		return nil, ErrProjectSlugRequired
	}
	return &Project{
		WorkspaceID: workspaceID,
		Name:        name,
		Slug:        slug,
	}, nil
}
