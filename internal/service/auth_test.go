package service

import (
	"context"
	"testing"

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
	svc := NewAuthService(repo)

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
	svc := NewAuthService(repo)

	_, err := svc.Register(context.Background(), "a@b.com", "secret", "Alice")
	if err != domain.ErrEmailAlreadyExists {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

