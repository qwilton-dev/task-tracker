package postgres

import (
	"context"
	"task-tracker/internal/domain"
	"task-tracker/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type InviteRepository struct {
	db *pgxpool.Pool
}

func NewInviteRepository(db *pgxpool.Pool) *InviteRepository {
	return &InviteRepository{db: db}
}

var _ repository.InviteRepository = (*InviteRepository)(nil)

func (r *InviteRepository) CreateInvite(ctx context.Context, invite *domain.Invite) error {
	query := `
		INSERT INTO invites (workspace_id, email, role, token, expires_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query,
		invite.WorkspaceID, invite.Email, invite.Role, invite.Token, invite.ExpiresAt, invite.CreatedBy,
	).Scan(&invite.ID, &invite.CreatedAt)
}

func (r *InviteRepository) GetInviteByToken(ctx context.Context, token string) (*domain.Invite, error) {
	query := `
		SELECT id, workspace_id, email, role, token, expires_at, accepted_at, created_by, created_at
		FROM invites
		WHERE token = $1
	`
	var inv domain.Invite
	err := r.db.QueryRow(ctx, query, token).Scan(
		&inv.ID, &inv.WorkspaceID, &inv.Email, &inv.Role, &inv.Token,
		&inv.ExpiresAt, &inv.AcceptedAt, &inv.CreatedBy, &inv.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *InviteRepository) AcceptInvite(ctx context.Context, token string) error {
	tag, err := r.db.Exec(ctx, `UPDATE invites SET accepted_at = now() WHERE token = $1 AND accepted_at IS NULL`, token)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrInviteAlreadyAccepted
	}
	return nil
}

func (r *InviteRepository) ListInvitesByWorkspace(ctx context.Context, workspaceID string) ([]*domain.Invite, error) {
	query := `
		SELECT id, workspace_id, email, role, token, expires_at, accepted_at, created_by, created_at
		FROM invites
		WHERE workspace_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []*domain.Invite
	for rows.Next() {
		var inv domain.Invite
		if err := rows.Scan(
			&inv.ID, &inv.WorkspaceID, &inv.Email, &inv.Role, &inv.Token,
			&inv.ExpiresAt, &inv.AcceptedAt, &inv.CreatedBy, &inv.CreatedAt,
		); err != nil {
			return nil, err
		}
		invites = append(invites, &inv)
	}
	return invites, nil
}
