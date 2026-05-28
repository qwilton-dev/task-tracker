package auth

import (
	"task-tracker/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secretKey string
	issuer    string
	audience  string
	expiresAt time.Duration
}

func NewJWTService(secretKey string, issuer string, audience string, expiresAt time.Duration) *JWTService {
	return &JWTService{secretKey: secretKey, issuer: issuer, audience: audience, expiresAt: expiresAt}
}

func (s *JWTService) GenerateToken(user *domain.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"iss": s.issuer,
		"aud": s.audience,
		"exp": time.Now().Add(s.expiresAt).Unix(),
	})
	return token.SignedString([]byte(s.secretKey))
}

func (s *JWTService) VerifyToken(tokenString string) (*domain.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	return &domain.User{ID: token.Claims.(jwt.MapClaims)["sub"].(string)}, nil
}
