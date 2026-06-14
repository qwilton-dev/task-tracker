package domain

import (
	"errors"
	"regexp"
	"strings"
)

var hexColorRegex = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type Label struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
}

func NewLabel(workspaceID, name, color string) (*Label, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrLabelNameRequired
	}
	if workspaceID == "" {
		return nil, ErrLabelWorkspaceIDRequired
	}
	if color == "" {
		color = "#ffffff"
	}
	if !hexColorRegex.MatchString(color) {
		return nil, ErrLabelColorInvalid
	}
	return &Label{
		WorkspaceID: workspaceID,
		Name:        name,
		Color:       color,
	}, nil
}

var (
	ErrLabelNameRequired        = errors.New("label name is required")
	ErrLabelWorkspaceIDRequired = errors.New("label workspace ID is required")
	ErrLabelColorInvalid        = errors.New("label color must be a valid hex color (e.g. #ff0000)")
	ErrLabelNotFound            = errors.New("label not found")
)
