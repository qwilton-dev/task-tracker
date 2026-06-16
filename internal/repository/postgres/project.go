package postgres

import (
	"context"
	"task-tracker/internal/domain"
	"task-tracker/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectRepository struct {
	db *pgxpool.Pool
}

func NewProjectRepository(db *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{db: db}
}

var _ repository.ProjectRepository = (*ProjectRepository)(nil)

func (r *ProjectRepository) CreateProject(ctx context.Context, project *domain.Project) error {
	query := `
		INSERT INTO project (workspace_id, name, "key")
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query, project.WorkspaceID, project.Name, project.Key).
		Scan(&project.ID, &project.CreatedAt)
}

func (r *ProjectRepository) GetProjectsByWorkspace(ctx context.Context, workspaceID string) ([]*domain.Project, error) {
	query := `SELECT id, workspace_id, name, "key", created_at FROM project WHERE workspace_id = $1`
	rows, err := r.db.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Key, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, &p)
	}
	return projects, nil
}

func (r *ProjectRepository) ExistsByKey(ctx context.Context, workspaceID, key string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM project WHERE workspace_id = $1 AND "key" = $2)`, workspaceID, key).Scan(&exists)
	return exists, err
}

func (r *ProjectRepository) GetProjectByID(ctx context.Context, id string) (*domain.Project, error) {
	query := `SELECT id, workspace_id, name, "key", created_at FROM project WHERE id = $1`
	var p domain.Project
	err := r.db.QueryRow(ctx, query, id).Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Key, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepository) UpdateProject(ctx context.Context, project *domain.Project) error {
	query := `UPDATE project SET name = $1, "key" = $2 WHERE id = $3`
	tag, err := r.db.Exec(ctx, query, project.Name, project.Key, project.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

func (r *ProjectRepository) DeleteProject(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM project WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}
