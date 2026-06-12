package postgres

import (
	"context"

	"task-tracker/internal/domain"
	"task-tracker/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CommentRepository struct {
	db *pgxpool.Pool
}

var _ repository.CommentRepository = (*CommentRepository)(nil)

func NewCommentRepository(db *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) CreateComment(ctx context.Context, comment *domain.Comment) error {
	query := `
		INSERT INTO comments (issue_id, author_id, body)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query, comment.IssueID, comment.AuthorID, comment.Body).
		Scan(&comment.ID, &comment.CreatedAt)
}

func (r *CommentRepository) ListCommentsByIssue(ctx context.Context, issueID string) ([]*domain.Comment, error) {
	query := `
		SELECT id, issue_id, author_id, body, created_at, updated_at
		FROM comments
		WHERE issue_id = $1
		ORDER BY created_at
	`
	rows, err := r.db.Query(ctx, query, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*domain.Comment
	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(&c.ID, &c.IssueID, &c.AuthorID, &c.Body, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, &c)
	}
	return comments, nil
}
