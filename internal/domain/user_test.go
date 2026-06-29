package domain

import (
	"testing"
)

func TestNormalizeRegister(t *testing.T) {
	email, password, name := NormalizeRegister("  QWILTON@GMAIL.COM  ", "  secret123  ", "  qwilton  ")
	if email != "qwilton@gmail.com" {
		t.Fatalf("email: got %q", email)
	}
	if password != "secret123" {
		t.Fatalf("password: got %q", password)
	}
	if name != "qwilton" {
		t.Fatalf("name: got %q", name)
	}
}

func TestValidateRegister(t *testing.T) {
	if err := ValidateRegister("", "x", "n"); err != ErrEmailRequired {
		t.Fatalf("expected ErrEmailRequired, got %v", err)
	}
	if err := ValidateRegister("e@e.com", "", "n"); err != ErrPasswordRequired {
		t.Fatalf("expected ErrPasswordRequired, got %v", err)
	}
	if err := ValidateRegister("e@e.com", "short", "n"); err != ErrPasswordTooShort {
		t.Fatalf("expected ErrPasswordTooShort, got %v", err)
	}
	if err := ValidateRegister("e@e.com", "validpass", ""); err != ErrNameRequired {
		t.Fatalf("expected ErrNameRequired, got %v", err)
	}
	if err := ValidateRegister("e@e.com", "validpass", "n"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestNewUser_NormalizesEmailAndName(t *testing.T) {
	u, err := NewUser("  A@B.COM  ", "hash", "  Alice  ")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if u.Email != "a@b.com" {
		t.Fatalf("email: got %q", u.Email)
	}
	if u.Name != "Alice" {
		t.Fatalf("name: got %q", u.Name)
	}
	if u.PasswordHash != "hash" {
		t.Fatalf("password hash: got %q", u.PasswordHash)
	}
}
