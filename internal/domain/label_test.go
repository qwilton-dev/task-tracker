package domain

import "testing"

func TestNewLabel_Valid(t *testing.T) {
	l, err := NewLabel("ws-1", "Bug", "#ff0000")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if l.WorkspaceID != "ws-1" {
		t.Fatalf("expected ws-1, got %s", l.WorkspaceID)
	}
	if l.Name != "Bug" {
		t.Fatalf("expected Bug, got %s", l.Name)
	}
	if l.Color != "#ff0000" {
		t.Fatalf("expected #ff0000, got %s", l.Color)
	}
}

func TestNewLabel_DefaultColor(t *testing.T) {
	l, err := NewLabel("ws-1", "Bug", "")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if l.Color != "#ffffff" {
		t.Fatalf("expected #ffffff, got %s", l.Color)
	}
}

func TestNewLabel_TrimSpace(t *testing.T) {
	l, err := NewLabel("ws-1", "  Bug  ", "#ff0000")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if l.Name != "Bug" {
		t.Fatalf("expected Bug, got %s", l.Name)
	}
}

func TestNewLabel_EmptyName(t *testing.T) {
	_, err := NewLabel("ws-1", "", "#ff0000")
	if err != ErrLabelNameRequired {
		t.Fatalf("expected ErrLabelNameRequired, got %v", err)
	}
}

func TestNewLabel_WhitespaceOnlyName(t *testing.T) {
	_, err := NewLabel("ws-1", "   ", "#ff0000")
	if err != ErrLabelNameRequired {
		t.Fatalf("expected ErrLabelNameRequired, got %v", err)
	}
}

func TestNewLabel_EmptyWorkspaceID(t *testing.T) {
	_, err := NewLabel("", "Bug", "#ff0000")
	if err != ErrLabelWorkspaceIDRequired {
		t.Fatalf("expected ErrLabelWorkspaceIDRequired, got %v", err)
	}
}

func TestNewLabel_InvalidColor(t *testing.T) {
	tests := []struct {
		color string
	}{
		{"red"},
		{"#fff"},
		{"#ffffff00"},
		{"#zzzzzz"},
		{"#GGGGGG"},
		{"ffffff"},
	}
	for _, tt := range tests {
		_, err := NewLabel("ws-1", "Bug", tt.color)
		if err != ErrLabelColorInvalid {
			t.Errorf("color=%q: expected ErrLabelColorInvalid, got %v", tt.color, err)
		}
	}
}

func TestNewLabel_ValidHexColors(t *testing.T) {
	tests := []struct {
		color string
	}{
		{"#000000"},
		{"#ffffff"},
		{"#FF00AA"},
		{"#abcdef"},
	}
	for _, tt := range tests {
		l, err := NewLabel("ws-1", "Bug", tt.color)
		if err != nil {
			t.Errorf("color=%q: unexpected err: %v", tt.color, err)
		}
		if l.Color != tt.color {
			t.Errorf("color=%q: expected %s, got %s", tt.color, tt.color, l.Color)
		}
	}
}
