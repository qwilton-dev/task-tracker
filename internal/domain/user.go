package domain

import (
	"errors"
	"strings"
	"time"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(email, passwordHash, name string) (*User, error) {
	email = normalizeEmail(email)
	name = strings.TrimSpace(name)
	if email == "" {
		return nil, ErrEmailRequired
	}
	if passwordHash == "" {
		return nil, ErrPasswordHashRequired
	}
	if name == "" {
		return nil, ErrNameRequired
	}
	return &User{
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
	}, nil
}

func NormalizeRegister(email, password, name string) (string, string, string) {
	return normalizeEmail(email), strings.TrimSpace(password), strings.TrimSpace(name)
}

func ValidateRegister(email, password, name string) error {
	email, password, name = NormalizeRegister(email, password, name)
	if email == "" {
		return ErrEmailRequired
	}
	if password == "" {
		return ErrPasswordRequired
	}
	if len(password) < 8 {
		return ErrPasswordTooShort
	}
	if name == "" {
		return ErrNameRequired
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

var (
	ErrEmailRequired        = errors.New("email is required")
	ErrEmailAlreadyExists   = errors.New("email already exists")
	ErrPasswordRequired     = errors.New("password is required")
	ErrPasswordTooShort     = errors.New("password must be at least 8 characters")
	ErrPasswordHashRequired = errors.New("password hash is required")
	ErrNameRequired         = errors.New("name is required")
	ErrUserNotFound         = errors.New("user not found")
)
