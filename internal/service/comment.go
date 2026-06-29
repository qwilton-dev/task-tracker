package service

import (
	"context"
	"encoding/json"
	"task-tracker/internal/domain"
	"task-tracker/internal/events"
	"task-tracker/internal/repository"
)

type CommentService struct {
	commentRepo repository.CommentRepository
	issueRepo   repository.IssueRepository
	activity    *ActivityEventService
	hub         *events.Hub
}

func NewCommentService(commentRepo repository.CommentRepository, issueRepo repository.IssueRepository, activity *ActivityEventService, hub *events.Hub) *CommentService {
	return &CommentService{commentRepo: commentRepo, issueRepo: issueRepo, activity: activity, hub: hub}
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
	s.activity.CreateActivityEvent(ctx, issueID, authorID, "comment.added", map[string]string{
		"comment_id": comment.ID,
	})
	issue, err := s.issueRepo.GetIssueByID(ctx, issueID)
	if err == nil && s.hub != nil {
		data, _ := json.Marshal(comment)
		s.hub.Publish(issue.ProjectID, events.Event{
			ProjectID: issue.ProjectID,
			Type:      "comment.added",
			Payload:   data,
		})
	}
	return comment, nil
}

func (s *CommentService) ListComments(ctx context.Context, issueID string) ([]*domain.Comment, error) {
	return s.commentRepo.ListCommentsByIssue(ctx, issueID)
}
