package postgres

import (
	"context"
	"task-tracker/internal/domain"
	"task-tracker/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkspaceRepository struct {
	db *pgxpool.Pool
}

var _ repository.WorkspaceRepository = (*WorkspaceRepository)(nil)

func NewWorkspaceRepository(db *pgxpool.Pool) *WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

func (r *WorkspaceRepository) CreateWorkspace(ctx context.Context, workspace *domain.Workspace) error {
	query := `
		INSERT INTO workspace (name)
		VALUES ($1)
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query, workspace.Name).
		Scan(&workspace.ID, &workspace.CreatedAt)
}

func (r *WorkspaceRepository) GetWorkspaceByID(ctx context.Context, id string) (*domain.Workspace, error) {
	query := `SELECT id, name, created_at FROM workspace WHERE id = $1`
	var ws domain.Workspace
	err := r.db.QueryRow(ctx, query, id).Scan(&ws.ID, &ws.Name, &ws.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

func (r *WorkspaceRepository) UpdateWorkspace(ctx context.Context, workspace *domain.Workspace) error {
	query := `UPDATE workspace SET name = $1 WHERE id = $2`
	tag, err := r.db.Exec(ctx, query, workspace.Name, workspace.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrWorkspaceNotFound
	}
	return nil
}

func (r *WorkspaceRepository) DeleteWorkspace(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM workspace WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrWorkspaceNotFound
	}
	return nil
}

func (r *WorkspaceRepository) ListWorkspaces(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	query := `
		SELECT w.id, w.name, w.created_at
		FROM workspace w
		JOIN workspace_member wm ON w.id = wm.workspace_id
		WHERE wm.user_id = $1
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []*domain.Workspace
	for rows.Next() {
		var w domain.Workspace
		if err := rows.Scan(&w.ID, &w.Name, &w.CreatedAt); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, &w)
	}
	return workspaces, nil
}

func (r *WorkspaceRepository) AddMember(ctx context.Context, workspaceID, userID, role string) error {
	query := `
		INSERT INTO workspace_member (workspace_id, user_id, role)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, query, workspaceID, userID, role)
	return err
}
