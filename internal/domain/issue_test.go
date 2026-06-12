package domain

import "testing"

func TestNewIssue(t *testing.T) {
	issue, err := NewIssue("proj-1", "Fix login bug", "user-1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if issue.ProjectID != "proj-1" {
		t.Fatalf("expected proj-1, got %s", issue.ProjectID)
	}
	if issue.Title != "Fix login bug" {
		t.Fatalf("expected Fix login bug, got %s", issue.Title)
	}
	if issue.Status != "backlog" {
		t.Fatalf("expected backlog, got %s", issue.Status)
	}
	if issue.Priority != "none" {
		t.Fatalf("expected none, got %s", issue.Priority)
	}
	if issue.CreatedBy != "user-1" {
		t.Fatalf("expected user-1, got %s", issue.CreatedBy)
	}
}

func TestNewIssue_EmptyTitle(t *testing.T) {
	_, err := NewIssue("proj-1", "", "user-1")
	if err != ErrIssueTitleRequired {
		t.Fatalf("expected ErrIssueTitleRequired, got %v", err)
	}
}

func TestNewIssue_TitleTooLong(t *testing.T) {
	title := ""
	for i := 0; i < 501; i++ {
		title += "a"
	}
	_, err := NewIssue("proj-1", title, "user-1")
	if err != ErrIssueTitleTooLong {
		t.Fatalf("expected ErrIssueTitleTooLong, got %v", err)
	}
}

func TestNewIssue_500CharTitle(t *testing.T) {
	title := ""
	for i := 0; i < 500; i++ {
		title += "a"
	}
	issue, err := NewIssue("proj-1", title, "user-1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(issue.Title) != 500 {
		t.Fatalf("expected 500, got %d", len(issue.Title))
	}
}

func TestValidateStatus(t *testing.T) {
	valid := []string{"backlog", "todo", "in_progress", "review", "done"}
	for _, s := range valid {
		if !ValidateStatus(s) {
			t.Errorf("expected %s to be valid", s)
		}
	}
	if ValidateStatus("invalid") {
		t.Error("expected 'invalid' to be invalid")
	}
}

func TestValidatePriority(t *testing.T) {
	valid := []string{"none", "low", "medium", "high", "urgent"}
	for _, p := range valid {
		if !ValidatePriority(p) {
			t.Errorf("expected %s to be valid", p)
		}
	}
	if ValidatePriority("invalid") {
		t.Error("expected 'invalid' to be invalid")
	}
}
