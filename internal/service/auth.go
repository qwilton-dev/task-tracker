package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"task-tracker/internal/auth"
	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	email, _, _ = domain.NormalizeRegister(email, "", "")
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", err
	}

	accessToken, err := s.jwt.GenerateToken(user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := generateRandomToken()
	if err != nil {
		return "", "", err
	}

	refreshTokenHash := hashToken(refreshToken)
	if err := s.tokenRepo.CreateRefreshToken(ctx, user.ID, refreshTokenHash, time.Now().Add(s.refreshTokenTTL)); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	refreshTokenHash := hashToken(refreshToken)
	token, err := s.tokenRepo.GetRefreshToken(ctx, refreshTokenHash)
	if err != nil {
		return "", "", err
	}
	if token.RevokedAt != nil || time.Now().After(token.ExpiresAt) {
		return "", "", fmt.Errorf("invalid token")
	}

	user, err := s.userRepo.GetUserByID(ctx, token.UserID)
	if err != nil {
		return "", "", err
	}

	if err := s.tokenRepo.RevokeRefreshToken(ctx, refreshTokenHash); err != nil {
		return "", "", err
	}

	accessToken, err := s.jwt.GenerateToken(user)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := generateRandomToken()
	if err != nil {
		return "", "", err
	}

	newRefreshTokenHash := hashToken(newRefreshToken)
	if err := s.tokenRepo.CreateRefreshToken(ctx, user.ID, newRefreshTokenHash, time.Now().Add(s.refreshTokenTTL)); err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func generateRandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.tokenRepo.RevokeRefreshToken(ctx, hashToken(refreshToken))
}

type AuthService struct {
	userRepo         repository.UserRepository
	tokenRepo        repository.TokenRepository
	jwt              *auth.JWTService
	refreshTokenTTL  time.Duration
}

func NewAuthService(userRepo repository.UserRepository, tokenRepo repository.TokenRepository, jwt *auth.JWTService, refreshTokenTTL time.Duration) *AuthService {
	return &AuthService{userRepo: userRepo, tokenRepo: tokenRepo, jwt: jwt, refreshTokenTTL: refreshTokenTTL}
}

func (s *AuthService) Register(ctx context.Context, email, password, name string) (*domain.User, error) {
	email, password, name = domain.NormalizeRegister(email, password, name)
	if err := domain.ValidateRegister(email, password, name); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := domain.NewUser(email, string(hash), name)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}
