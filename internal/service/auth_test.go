package service

import (
	"context"
	"testing"
	"time"

	"task-tracker/internal/auth"
	"task-tracker/internal/domain"
	"task-tracker/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	createFn     func(ctx context.Context, user *domain.User) error
	getByEmailFn func(ctx context.Context, email string) (*domain.User, error)
	getByIDFn    func(ctx context.Context, id string) (*domain.User, error)
	updateFn     func(ctx context.Context, user *domain.User) error
	deleteFn     func(ctx context.Context, id string) error
}

var _ repository.UserRepository = (*mockUserRepo)(nil)

func (m *mockUserRepo) CreateUser(ctx context.Context, user *domain.User) error {
	if m.createFn == nil {
		return nil
	}
	return m.createFn(ctx, user)
}
func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getByEmailFn == nil {
		return nil, domain.ErrUserNotFound
	}
	return m.getByEmailFn(ctx, email)
}
func (m *mockUserRepo) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	if m.getByIDFn == nil {
		return nil, domain.ErrUserNotFound
	}
	return m.getByIDFn(ctx, id)
}
func (m *mockUserRepo) UpdateUser(ctx context.Context, user *domain.User) error {
	if m.updateFn == nil {
		return nil
	}
	return m.updateFn(ctx, user)
}
func (m *mockUserRepo) DeleteUser(ctx context.Context, id string) error {
	if m.deleteFn == nil {
		return nil
	}
	return m.deleteFn(ctx, id)
}

func TestAuthService_Register_HashesPasswordAndNormalizesEmail(t *testing.T) {
	var got *domain.User
	repo := &mockUserRepo{
		createFn: func(ctx context.Context, user *domain.User) error {
			user.ID = "id-1"
			got = user
			return nil
		},
	}
	tokenRepo := &mockTokenRepo{}
	jwtSvc := auth.NewJWTService("secret", "test", "api", 1*time.Hour)
	svc := NewAuthService(repo, tokenRepo, jwtSvc, 7*24*time.Hour)

	user, err := svc.Register(context.Background(), "  A@B.COM  ", "secret", " Alice ")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if user.ID != "id-1" {
		t.Fatalf("id: got %q", user.ID)
	}
	if got == nil {
		t.Fatalf("expected repo.CreateUser called")
	}
	if got.Email != "a@b.com" {
		t.Fatalf("email: got %q", got.Email)
	}
	if got.Name != "Alice" {
		t.Fatalf("name: got %q", got.Name)
	}
	if got.PasswordHash == "secret" || got.PasswordHash == "" {
		t.Fatalf("password hash not set properly")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(got.PasswordHash), []byte("secret")); err != nil {
		t.Fatalf("hash does not match password: %v", err)
	}
}

func TestAuthService_Register_EmailAlreadyExists(t *testing.T) {
	repo := &mockUserRepo{
		createFn: func(ctx context.Context, user *domain.User) error {
			return domain.ErrEmailAlreadyExists
		},
	}
	tokenRepo := &mockTokenRepo{}
	jwtSvc := auth.NewJWTService("secret", "test", "api", 1*time.Hour)
	svc := NewAuthService(repo, tokenRepo, jwtSvc, 7*24*time.Hour)

	_, err := svc.Register(context.Background(), "a@b.com", "secret", "Alice")
	if err != domain.ErrEmailAlreadyExists {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

type mockTokenRepo struct{}

var _ repository.TokenRepository = (*mockTokenRepo)(nil)

func (m *mockTokenRepo) CreateRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	return nil
}
func (m *mockTokenRepo) GetRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	return nil, nil
}
func (m *mockTokenRepo) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	return nil
}
func (m *mockTokenRepo) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
	return nil
}
