package domain

import (
	"encoding/json"
	"time"
)

type ActivityEvent struct {
	ID        string          `json:"id"`
	IssueID   string          `json:"issue_id"`
	ActorID   string          `json:"actor_id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}

func NewActivityEvent(issueID, actorID, eventType string, payload any) (*ActivityEvent, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &ActivityEvent{
		IssueID:   issueID,
		ActorID:   actorID,
		Type:      eventType,
		Payload:   payloadBytes,
		CreatedAt: time.Now(),
	}, nil
}
