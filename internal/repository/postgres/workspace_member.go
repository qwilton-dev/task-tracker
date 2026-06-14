package postgres

import (
	"context"
	"task-tracker/internal/domain"
	"task-tracker/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkspaceMemberRepository struct {
	db *pgxpool.Pool
}

var _ repository.WorkspaceMemberRepository = (*WorkspaceMemberRepository)(nil)

func NewWorkspaceMemberRepository(db *pgxpool.Pool) *WorkspaceMemberRepository {
	return &WorkspaceMemberRepository{db: db}
}

func (r *WorkspaceMemberRepository) CreateWorkspaceMember(ctx context.Context, workspaceMember *domain.WorkspaceMember) error {
	query := `
		INSERT INTO workspace_member (workspace_id, user_id, role)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, query, workspaceMember.WorkspaceId, workspaceMember.UserId, workspaceMember.Role)
	return err
}
func (r *WorkspaceMemberRepository) GetWorkspaceMembers(ctx context.Context, workspaceId string) ([]*domain.WorkspaceMember, error) {
	query := `
		SELECT workspace_id, user_id, role
		FROM workspace_member
		WHERE workspace_id = $1
	`
	rows, err := r.db.Query(ctx, query, workspaceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.WorkspaceMember
	for rows.Next() {
		var m domain.WorkspaceMember
		err := rows.Scan(&m.WorkspaceId, &m.UserId, &m.Role)
		if err != nil {
			return nil, err
		}
		members = append(members, &m)
	}
	return members, nil
}

func (r *WorkspaceMemberRepository) DeleteWorkspaceMember(ctx context.Context, workspaceMember *domain.WorkspaceMember) error {
	query := `
		DELETE FROM workspace_member
		WHERE workspace_id = $1 AND user_id = $2
	`
	_, err := r.db.Exec(ctx, query, workspaceMember.WorkspaceId, workspaceMember.UserId)
	return err
}

func (r *WorkspaceMemberRepository) UpdateWorkspaceMemberRole(ctx context.Context, workspaceMember *domain.WorkspaceMember) error {
	query := `
		UPDATE workspace_member
		SET role = $3
		WHERE workspace_id = $1 AND user_id = $2
	`
	_, err := r.db.Exec(ctx, query, workspaceMember.WorkspaceId, workspaceMember.UserId, workspaceMember.Role)
	return err
}

func (r *WorkspaceMemberRepository) GetRole(ctx context.Context, workspaceID, userID string) (string, error) {
	query := `
		SELECT wm.role
		FROM workspace_member wm
		WHERE wm.workspace_id = $1 AND wm.user_id = $2
	`
	var role string
	err := r.db.QueryRow(ctx, query, workspaceID, userID).Scan(&role)
	return role, err
}

func (r *WorkspaceMemberRepository) GetRoleByProjectID(ctx context.Context, projectID, userID string) (string, error) {
	query := `
		SELECT wm.role
		FROM workspace_member wm
		JOIN project p ON wm.workspace_id = p.workspace_id
		WHERE p.id = $1 AND wm.user_id = $2
	`
	var role string
	err := r.db.QueryRow(ctx, query, projectID, userID).Scan(&role)
	return role, err
}

func (r *WorkspaceMemberRepository) GetRoleByIssueID(ctx context.Context, issueID, userID string) (string, error) {
	query := `
		SELECT wm.role
		FROM workspace_member wm
		JOIN project p ON wm.workspace_id = p.workspace_id
		JOIN issues i ON i.project_id = p.id
		WHERE i.id = $1 AND wm.user_id = $2
	`
	var role string
	err := r.db.QueryRow(ctx, query, issueID, userID).Scan(&role)
	return role, err
}
