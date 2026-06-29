package domain

import (
	"errors"
	"time"
)

type Issue struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	AssigneeID  string    `json:"assignee_id,omitempty"`
	Position    float64   `json:"position"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

var (
	ErrIssueTitleRequired   = errors.New("issue title is required")
	ErrIssueTitleTooLong    = errors.New("issue title must be 1-500 characters")
	ErrIssueInvalidStatus   = errors.New("invalid issue status")
	ErrIssueInvalidPriority = errors.New("invalid issue priority")
	ErrIssueNotFound        = errors.New("issue not found")
)

var validStatuses = map[string]bool{
	"backlog":     true,
	"todo":        true,
	"in_progress": true,
	"review":      true,
	"done":        true,
}

var validPriorities = map[string]bool{
	"none":   true,
	"low":    true,
	"medium": true,
	"high":   true,
	"urgent": true,
}

func NewIssue(projectID, title, createdBy string) (*Issue, error) {
	if title == "" {
		return nil, ErrIssueTitleRequired
	}
	if len(title) > 500 {
		return nil, ErrIssueTitleTooLong
	}
	return &Issue{
		ProjectID: projectID,
		Title:     title,
		Status:    "backlog",
		Priority:  "none",
		CreatedBy: createdBy,
	}, nil
}

func ValidateStatus(s string) bool {
	return validStatuses[s]
}

func ValidatePriority(p string) bool {
	return validPriorities[p]
}
