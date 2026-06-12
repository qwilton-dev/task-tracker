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
		INSERT INTO project (workspace_id, name, key)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query, project.WorkspaceID, project.Name, project.Key).
		Scan(&project.ID, &project.CreatedAt)
}

func (r *ProjectRepository) GetProjectsByWorkspace(ctx context.Context, workspaceID string) ([]*domain.Project, error) {
	query := `SELECT id, workspace_id, name, key, created_at FROM project WHERE workspace_id = $1`
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
