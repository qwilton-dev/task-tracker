package service

import (
	"context"
	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
)

type ActivityEventService struct {
	activityEventRepo repository.ActivityEventRepository
}

func NewActivityEventService(activityEventRepo repository.ActivityEventRepository) *ActivityEventService {
	return &ActivityEventService{activityEventRepo: activityEventRepo}
}

func (s *ActivityEventService) CreateActivityEvent(ctx context.Context, issueID, actorID, eventType string, payload any) (*domain.ActivityEvent, error) {
	event, err := domain.NewActivityEvent(issueID, actorID, eventType, payload)
	if err != nil {
		return nil, err
	}
	if err := s.activityEventRepo.Create(ctx, event); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *ActivityEventService) ListActivityEventsByIssue(ctx context.Context, issueID string) ([]*domain.ActivityEvent, error) {
	return s.activityEventRepo.ListByIssue(ctx, issueID)
}
