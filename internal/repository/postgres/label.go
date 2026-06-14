package postgres

import (
	"context"

	"task-tracker/internal/domain"
	"task-tracker/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LabelRepository struct {
	db *pgxpool.Pool
}

func NewLabelRepository(db *pgxpool.Pool) *LabelRepository {
	return &LabelRepository{db: db}
}

var _ repository.LabelRepository = (*LabelRepository)(nil)

func (r *LabelRepository) CreateLabel(ctx context.Context, label *domain.Label) error {
	query := `
		INSERT INTO label (workspace_id, name, color)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	return r.db.QueryRow(ctx, query, label.WorkspaceID, label.Name, label.Color).Scan(&label.ID)
}

func (r *LabelRepository) GetLabelByID(ctx context.Context, id string) (*domain.Label, error) {
	query := `
		SELECT id, workspace_id, name, color
		FROM label
		WHERE id = $1
	`
	var l domain.Label
	err := r.db.QueryRow(ctx, query, id).Scan(&l.ID, &l.WorkspaceID, &l.Name, &l.Color)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *LabelRepository) ListLabelsByWorkspace(ctx context.Context, workspaceID string) ([]*domain.Label, error) {
	query := `
		SELECT id, workspace_id, name, color
		FROM label
		WHERE workspace_id = $1
		ORDER BY name
	`
	rows, err := r.db.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []*domain.Label
	for rows.Next() {
		var l domain.Label
		if err := rows.Scan(&l.ID, &l.WorkspaceID, &l.Name, &l.Color); err != nil {
			return nil, err
		}
		labels = append(labels, &l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return labels, nil
}

func (r *LabelRepository) UpdateLabel(ctx context.Context, label *domain.Label) error {
	query := `
		UPDATE label
		SET name = $1, color = $2
		WHERE id = $3
	`
	tag, err := r.db.Exec(ctx, query, label.Name, label.Color, label.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrLabelNotFound
	}
	return nil
}

func (r *LabelRepository) DeleteLabel(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM label WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrLabelNotFound
	}
	return nil
}

func (r *LabelRepository) AttachLabel(ctx context.Context, issueID, labelID string) error {
	query := `
		INSERT INTO issue_label (issue_id, label_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, issueID, labelID)
	return err
}

func (r *LabelRepository) DetachLabel(ctx context.Context, issueID, labelID string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM issue_label WHERE issue_id = $1 AND label_id = $2`, issueID, labelID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrLabelNotFound
	}
	return nil
}

func (r *LabelRepository) ListLabelsByIssue(ctx context.Context, issueID string) ([]*domain.Label, error) {
	query := `
		SELECT l.id, l.workspace_id, l.name, l.color
		FROM label l
		JOIN issue_label il ON l.id = il.label_id
		WHERE il.issue_id = $1
		ORDER BY l.name
	`
	rows, err := r.db.Query(ctx, query, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []*domain.Label
	for rows.Next() {
		var l domain.Label
		if err := rows.Scan(&l.ID, &l.WorkspaceID, &l.Name, &l.Color); err != nil {
			return nil, err
		}
		labels = append(labels, &l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return labels, nil
}
