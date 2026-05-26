package postgres

import (
	"errors"
	"testing"

	"task-tracker/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapUserError_NoRows(t *testing.T) {
	if got := mapUserError(pgx.ErrNoRows); !errors.Is(got, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", got)
	}
}

func TestMapUserError_UniqueViolation(t *testing.T) {
	err := &pgconn.PgError{Code: "23505"}
	if got := mapUserError(err); !errors.Is(got, domain.ErrEmailAlreadyExists) {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", got)
	}
}

