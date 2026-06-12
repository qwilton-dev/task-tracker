package domain

import (
	"errors"
	"regexp"
	"time"
)

var projectKeyRegex = regexp.MustCompile(`^[A-Z]{2,5}$`)

type Project struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Name        string    `json:"name"`
	Key         string    `json:"key"`
	CreatedAt   time.Time `json:"created_at"`
}

var (
	ErrProjectNameRequired = errors.New("project name is required")
	ErrProjectKeyRequired  = errors.New("project key is required")
	ErrProjectKeyInvalid   = errors.New("project key must be 2-5 uppercase letters")
)

func NewProject(workspaceID, name, key string) (*Project, error) {
	if name == "" {
		return nil, ErrProjectNameRequired
	}
	if key == "" {
		return nil, ErrProjectKeyRequired
	}
	if !projectKeyRegex.MatchString(key) {
		return nil, ErrProjectKeyInvalid
	}
	return &Project{
		WorkspaceID: workspaceID,
		Name:        name,
		Key:         key,
	}, nil
}
