package repository

import (
	"context"

	"task-tracker/internal/domain"
)

type CommentRepository interface {
	CreateComment(ctx context.Context, comment *domain.Comment) error
	ListCommentsByIssue(ctx context.Context, issueID string) ([]*domain.Comment, error)
}
