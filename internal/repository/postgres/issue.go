package postgres

import (
	"context"
	"errors"
	"strconv"

	"task-tracker/internal/domain"
	"task-tracker/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IssueRepository struct {
	db *pgxpool.Pool
}

var _ repository.IssueRepository = (*IssueRepository)(nil)

func NewIssueRepository(db *pgxpool.Pool) *IssueRepository {
	return &IssueRepository{db: db}
}

func (r *IssueRepository) CreateIssue(ctx context.Context, issue *domain.Issue) error {
	query := `
		INSERT INTO issues (project_id, number, title, description, status, priority, assignee_id, position, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		issue.ProjectID, issue.Number, issue.Title, issue.Description,
		issue.Status, issue.Priority, strToPtr(issue.AssigneeID), issue.Position, issue.CreatedBy,
	).Scan(&issue.ID, &issue.CreatedAt, &issue.UpdatedAt)
}

func (r *IssueRepository) CreateIssueTx(ctx context.Context, issue *domain.Issue) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var maxNum *int
	err = tx.QueryRow(ctx, `SELECT MAX(number) FROM issues WHERE project_id = $1 FOR UPDATE`, issue.ProjectID).Scan(&maxNum)
	if err != nil {
		return err
	}
	if maxNum != nil {
		issue.Number = *maxNum + 1
	} else {
		issue.Number = 1
	}

	query := `
		INSERT INTO issues (project_id, number, title, description, status, priority, assignee_id, position, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRow(ctx, query,
		issue.ProjectID, issue.Number, issue.Title, issue.Description,
		issue.Status, issue.Priority, strToPtr(issue.AssigneeID), issue.Position, issue.CreatedBy,
	).Scan(&issue.ID, &issue.CreatedAt, &issue.UpdatedAt)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *IssueRepository) GetIssueByID(ctx context.Context, id string) (*domain.Issue, error) {
	query := `
		SELECT id, project_id, number, title, COALESCE(description, ''),
		       status, priority, assignee_id, position, created_by, created_at, updated_at
		FROM issues
		WHERE id = $1
	`
	var issue domain.Issue
	var assigneeID *string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&issue.ID, &issue.ProjectID, &issue.Number, &issue.Title, &issue.Description,
		&issue.Status, &issue.Priority, &assigneeID, &issue.Position, &issue.CreatedBy,
		&issue.CreatedAt, &issue.UpdatedAt,
	)
	if err != nil {
		return nil, mapIssueError(err)
	}
	if assigneeID != nil {
		issue.AssigneeID = *assigneeID
	}
	return &issue, nil
}

func (r *IssueRepository) ListIssuesByProject(ctx context.Context, projectID string, filters repository.IssueFilters) ([]*domain.Issue, error) {
	query := `
		SELECT id, project_id, number, title, COALESCE(description, ''),
		       status, priority, assignee_id, position, created_by, created_at, updated_at
		FROM issues
		WHERE project_id = $1
	`
	args := []any{projectID}
	argIdx := 2

	if filters.Status != "" {
		query += ` AND status = $` + strconv.Itoa(argIdx)
		args = append(args, filters.Status)
		argIdx++
	}
	if filters.Assignee != "" {
		query += ` AND assignee_id = $` + strconv.Itoa(argIdx)
		args = append(args, filters.Assignee)
		argIdx++
	}
	if filters.Q != "" {
		query += ` AND title ILIKE $` + strconv.Itoa(argIdx)
		args = append(args, "%"+filters.Q+"%")
		argIdx++
	}

	query += ` ORDER BY status, position, created_at`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []*domain.Issue
	for rows.Next() {
		var issue domain.Issue
		var assigneeID *string
		if err := rows.Scan(
			&issue.ID, &issue.ProjectID, &issue.Number, &issue.Title, &issue.Description,
			&issue.Status, &issue.Priority, &assigneeID, &issue.Position, &issue.CreatedBy,
			&issue.CreatedAt, &issue.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if assigneeID != nil {
			issue.AssigneeID = *assigneeID
		}
		issues = append(issues, &issue)
	}
	return issues, nil
}

func (r *IssueRepository) UpdateIssue(ctx context.Context, issue *domain.Issue) error {
	query := `
		UPDATE issues
		SET title = $1, description = $2, priority = $3, assignee_id = $4, updated_at = now()
		WHERE id = $5
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query,
		issue.Title, issue.Description, issue.Priority, strToPtr(issue.AssigneeID), issue.ID,
	).Scan(&issue.UpdatedAt)
}

func (r *IssueRepository) DeleteIssue(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM issues WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrIssueNotFound
	}
	return nil
}

func (r *IssueRepository) MoveIssue(ctx context.Context, id, status string, position float64) error {
	query := `
		UPDATE issues
		SET status = $1, position = $2, updated_at = now()
		WHERE id = $3
	`
	tag, err := r.db.Exec(ctx, query, status, position, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrIssueNotFound
	}
	return nil
}

func (r *IssueRepository) GetMaxNumber(ctx context.Context, projectID string) (int, error) {
	var maxNum *int
	err := r.db.QueryRow(ctx, `SELECT MAX(number) FROM issues WHERE project_id = $1`, projectID).Scan(&maxNum)
	if err != nil {
		return 0, err
	}
	if maxNum == nil {
		return 0, nil
	}
	return *maxNum, nil
}

func mapIssueError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrIssueNotFound
	}
	return err
}

func strToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
