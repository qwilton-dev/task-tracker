package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
	"task-tracker/internal/service"
)

type handlerRepo struct {
	createFn func(ctx context.Context, user *domain.User) error
}

var _ repository.UserRepository = (*handlerRepo)(nil)

func (h *handlerRepo) CreateUser(ctx context.Context, user *domain.User) error {
	if h.createFn == nil {
		return nil
	}
	return h.createFn(ctx, user)
}
func (h *handlerRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}
func (h *handlerRepo) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}
func (h *handlerRepo) UpdateUser(ctx context.Context, user *domain.User) error {
	return nil
}
func (h *handlerRepo) DeleteUser(ctx context.Context, id string) error {
	return nil
}

func TestAuthHandler_Register_201(t *testing.T) {
	repo := &handlerRepo{
		createFn: func(ctx context.Context, user *domain.User) error {
			user.ID = "u1"
			user.CreatedAt = time.Date(2026, 5, 26, 7, 20, 58, 0, time.UTC)
			user.UpdatedAt = user.CreatedAt
			return nil
		},
	}
	svc := service.NewAuthService(repo)
	h := NewAuthHandler(svc)

	body := bytes.NewBufferString(`{"email":"a@b.com","password":"secret","name":"Alice"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status: got %d body=%s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp["id"] != "u1" {
		t.Fatalf("id: got %v", resp["id"])
	}
	if resp["email"] != "a@b.com" {
		t.Fatalf("email: got %v", resp["email"])
	}
	if resp["name"] != "Alice" {
		t.Fatalf("name: got %v", resp["name"])
	}
}

func TestAuthHandler_Register_409(t *testing.T) {
	repo := &handlerRepo{
		createFn: func(ctx context.Context, user *domain.User) error {
			return domain.ErrEmailAlreadyExists
		},
	}
	svc := service.NewAuthService(repo)
	h := NewAuthHandler(svc)

	body := bytes.NewBufferString(`{"email":"a@b.com","password":"secret","name":"Alice"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", body)
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("status: got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestAuthHandler_Register_400_Validation(t *testing.T) {
	repo := &handlerRepo{}
	svc := service.NewAuthService(repo)
	h := NewAuthHandler(svc)

	body := bytes.NewBufferString(`{"email":"","password":"secret","name":"Alice"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", body)
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d body=%s", rr.Code, rr.Body.String())
	}
}

