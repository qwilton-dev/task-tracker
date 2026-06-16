package postgres

import (
	"context"

	"task-tracker/internal/domain"
	"task-tracker/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ActivityEventRepository struct {
	db *pgxpool.Pool
}

func NewActivityEventRepository(db *pgxpool.Pool) *ActivityEventRepository {
	return &ActivityEventRepository{db: db}
}

var _ repository.ActivityEventRepository = (*ActivityEventRepository)(nil)

func (r *ActivityEventRepository) Create(ctx context.Context, event *domain.ActivityEvent) error {
	query := `
		INSERT INTO activity_events (issue_id, actor_id, event_type, payload)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query,
		event.IssueID, event.ActorID, event.Type, event.Payload,
	).Scan(&event.ID, &event.CreatedAt)
}

func (r *ActivityEventRepository) ListByIssue(ctx context.Context, issueID string) ([]*domain.ActivityEvent, error) {
	query := `
		SELECT id, issue_id, COALESCE(actor_id::text, ''), event_type, payload, created_at
		FROM activity_events
		WHERE issue_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.ActivityEvent
	for rows.Next() {
		var e domain.ActivityEvent
		if err := rows.Scan(&e.ID, &e.IssueID, &e.ActorID, &e.Type, &e.Payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}
