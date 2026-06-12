package handler

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"task-tracker/internal/auth"
	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
	"task-tracker/internal/service"
)

type handlerRepo struct {
	createFn             func(ctx context.Context, user *domain.User) error
	getUserByEmailFn     func(ctx context.Context, email string) (*domain.User, error)
	getUserByIDFn        func(ctx context.Context, id string) (*domain.User, error)
	getRefreshTokenFn    func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	revokeRefreshTokenFn func(ctx context.Context, tokenHash string) error
}

var _ repository.UserRepository = (*handlerRepo)(nil)
var _ repository.TokenRepository = (*handlerRepo)(nil)

func (h *handlerRepo) CreateUser(ctx context.Context, user *domain.User) error {
	if h.createFn == nil {
		return nil
	}
	return h.createFn(ctx, user)
}
func (h *handlerRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	if h.getUserByEmailFn != nil {
		return h.getUserByEmailFn(ctx, email)
	}
	return nil, domain.ErrUserNotFound
}
func (h *handlerRepo) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	if h.getUserByIDFn != nil {
		return h.getUserByIDFn(ctx, id)
	}
	return nil, domain.ErrUserNotFound
}
func (h *handlerRepo) UpdateUser(ctx context.Context, user *domain.User) error {
	return nil
}
func (h *handlerRepo) DeleteUser(ctx context.Context, id string) error {
	return nil
}

func (h *handlerRepo) CreateRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	return nil
}
func (h *handlerRepo) GetRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	if h.getRefreshTokenFn != nil {
		return h.getRefreshTokenFn(ctx, tokenHash)
	}
	return nil, domain.ErrUserNotFound
}
func (h *handlerRepo) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	if h.revokeRefreshTokenFn != nil {
		return h.revokeRefreshTokenFn(ctx, tokenHash)
	}
	return nil
}
func (h *handlerRepo) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
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
	jwtSvc := auth.NewJWTService("secret", "test", "api", 1*time.Hour)
	svc := service.NewAuthService(repo, repo, jwtSvc, 7*24*time.Hour)
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
	jwtSvc := auth.NewJWTService("secret", "test", "api", 1*time.Hour)
	svc := service.NewAuthService(repo, repo, jwtSvc, 7*24*time.Hour)
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
	jwtSvc := auth.NewJWTService("secret", "test", "api", 1*time.Hour)
	svc := service.NewAuthService(repo, repo, jwtSvc, 7*24*time.Hour)
	h := NewAuthHandler(svc)

	body := bytes.NewBufferString(`{"email":"","password":"secret","name":"Alice"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", body)
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestAuthHandler_Refresh_200(t *testing.T) {
	rawRefreshToken := "refresh123"
	repo := &handlerRepo{
		getRefreshTokenFn: func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
			if tokenHash != hashToken(rawRefreshToken) {
				return nil, domain.ErrUserNotFound
			}
			return &domain.RefreshToken{
				UserID:    "u1",
				TokenHash: tokenHash,
				ExpiresAt: time.Now().Add(1 * time.Hour),
				CreatedAt: time.Now(),
			}, nil
		},
		getUserByIDFn: func(ctx context.Context, id string) (*domain.User, error) {
			return &domain.User{ID: "u1", Email: "a@b.com", Name: "Alice"}, nil
		},
		revokeRefreshTokenFn: func(ctx context.Context, tokenHash string) error {
			return nil
		},
	}
	jwtSvc := auth.NewJWTService("secret", "test", "api", 1*time.Hour)
	svc := service.NewAuthService(repo, repo, jwtSvc, 7*24*time.Hour)
	h := NewAuthHandler(svc)

	body := bytes.NewBufferString(`{"refresh_token":"` + rawRefreshToken + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", body)
	rr := httptest.NewRecorder()

	h.Refresh(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp["access_token"] == "" || resp["refresh_token"] == "" {
		t.Fatalf("expected tokens, got %v", resp)
	}
}

func TestAuthHandler_Logout_200(t *testing.T) {
	rawRefreshToken := "refresh123"
	repo := &handlerRepo{
		revokeRefreshTokenFn: func(ctx context.Context, tokenHash string) error {
			if tokenHash != hashToken(rawRefreshToken) {
				return domain.ErrUserNotFound
			}
			return nil
		},
	}
	jwtSvc := auth.NewJWTService("secret", "test", "api", 1*time.Hour)
	svc := service.NewAuthService(repo, repo, jwtSvc, 7*24*time.Hour)
	h := NewAuthHandler(svc)

	body := bytes.NewBufferString(`{"refresh_token":"` + rawRefreshToken + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", body)
	rr := httptest.NewRecorder()

	h.Logout(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%s", rr.Code, rr.Body.String())
	}
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
