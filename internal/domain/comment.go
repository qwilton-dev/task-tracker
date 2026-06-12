package domain

import (
	"errors"
	"time"
)

type Comment struct {
	ID        string     `json:"id"`
	IssueID   string     `json:"issue_id"`
	AuthorID  string     `json:"author_id"`
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

var (
	ErrCommentBodyRequired = errors.New("comment body is required")
	ErrCommentNotFound     = errors.New("comment not found")
)

func NewComment(issueID, authorID, body string) (*Comment, error) {
	if body == "" {
		return nil, ErrCommentBodyRequired
	}
	return &Comment{
		IssueID:  issueID,
		AuthorID: authorID,
		Body:     body,
	}, nil
}
