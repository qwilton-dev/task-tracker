package postgres

import (
	"context"
	"task-tracker/internal/domain"
	"task-tracker/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WokspaceRepository struct {
	db *pgxpool.Pool
}

var _ repository.WorkspaceRepository = (*WokspaceRepository)(nil)

func NewWorkspaceRepository(db *pgxpool.Pool) *WokspaceRepository {
	return &WokspaceRepository{db: db}
}

func (r *WokspaceRepository) CreateWorkspace(ctx context.Context, workspace *domain.Workspace) error {
	query := `
		INSERT INTO workspace (name, slug)
		VALUES ($1, $2)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(ctx, query, workspace.Name, workspace.Slug).
		Scan(&workspace.ID, &workspace.CreatedAt)
	return err
}

func (r *WokspaceRepository) GetWorkspaceByID(ctx context.Context, id string) (*domain.Workspace, error) {
	query := `
		SELECT id, name, slug, created_at
		FROM workspace
		WHERE id = $1
	`
	var workspace domain.Workspace
	err := r.db.QueryRow(ctx, query, id).
		Scan(&workspace.ID, &workspace.Name, &workspace.Slug, &workspace.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}
func (r *WokspaceRepository) GetWorkspaceBySlug(ctx context.Context, slug string) (*domain.Workspace, error) {
	query := `
		SELECT id, name, slug, created_at
		FROM workspace
		WHERE slug = $1
	`
	var workspace domain.Workspace
	err := r.db.QueryRow(ctx, query, slug).
		Scan(&workspace.ID, &workspace.Name, &workspace.Slug, &workspace.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}
func (r *WokspaceRepository) UpdateWorkspace(ctx context.Context, workspace *domain.Workspace) error {
	query := `
		UPDATE workspace
		SET name = $1, slug = $2
		WHERE id = $3
	`
	_, err := r.db.Exec(ctx, query, workspace.Name, workspace.Slug, workspace.ID)
	return err
}
func (r *WokspaceRepository) DeleteWorkspace(ctx context.Context, id string) error {
	query := `
		DELETE FROM workspace
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *WokspaceRepository) ListWorkspaces(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	query := `
		SELECT w.id, w.name, w.slug, w.created_at
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
		if err := rows.Scan(&w.ID, &w.Name, &w.Slug, &w.CreatedAt); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, &w)
	}
	return workspaces, nil
}

func (r *WokspaceRepository) AddMember(ctx context.Context, workspaceID, userID, role string) error {
	query := `
		INSERT INTO workspace_member (workspace_id, user_id, role)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, query, workspaceID, userID, role)
	return err
}
