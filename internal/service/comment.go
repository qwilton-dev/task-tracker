package service

import (
	"context"
	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
)

type CommentService struct {
	commentRepo repository.CommentRepository
}

func NewCommentService(commentRepo repository.CommentRepository) *CommentService {
	return &CommentService{commentRepo: commentRepo}
}

func (s *CommentService) CreateComment(ctx context.Context, issueID, authorID, body string) (*domain.Comment, error) {
	if body == "" {
		return nil, domain.ErrCommentBodyRequired
	}
	comment, err := domain.NewComment(issueID, authorID, body)
	if err != nil {
		return nil, err
	}
	if err := s.commentRepo.CreateComment(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *CommentService) ListComments(ctx context.Context, issueID string) ([]*domain.Comment, error) {
	return s.commentRepo.ListCommentsByIssue(ctx, issueID)
}
