package domain

import "testing"

func TestNewProject_ValidKey(t *testing.T) {
	p, err := NewProject("ws-1", "Backend", "BE")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if p.Key != "BE" {
		t.Fatalf("expected BE, got %s", p.Key)
	}
	if p.WorkspaceID != "ws-1" {
		t.Fatalf("expected ws-1, got %s", p.WorkspaceID)
	}
}

func TestNewProject_InvalidKey(t *testing.T) {
	tests := []struct {
		key string
		err error
	}{
		{"", ErrProjectKeyRequired},
		{"B", ErrProjectKeyInvalid},
		{"BEEEEE", ErrProjectKeyInvalid},
		{"be", ErrProjectKeyInvalid},
		{"BE-1", ErrProjectKeyInvalid},
		{"123", ErrProjectKeyInvalid},
	}
	for _, tt := range tests {
		_, err := NewProject("ws-1", "Name", tt.key)
		if err != tt.err {
			t.Errorf("key=%q: expected %v, got %v", tt.key, tt.err, err)
		}
	}
}

func TestNewProject_EmptyName(t *testing.T) {
	_, err := NewProject("ws-1", "", "BE")
	if err != ErrProjectNameRequired {
		t.Fatalf("expected ErrProjectNameRequired, got %v", err)
	}
}

func TestNewProject_5CharKey(t *testing.T) {
	p, err := NewProject("ws-1", "Name", "ABCDE")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if p.Key != "ABCDE" {
		t.Fatalf("expected ABCDE, got %s", p.Key)
	}
}

func TestGenerateKey(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Backend", "BACKE"},
		{"Front End", "FE"},
		{"My Cool Project", "MCP"},
		{"API", "API"},
		{"X", "XX"},
		{"Auth Service", "AS"},
		{"User Management Panel", "UMP"},
	}
	for _, tt := range tests {
		got := GenerateKey(tt.name)
		if got != tt.want {
			t.Errorf("GenerateKey(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
