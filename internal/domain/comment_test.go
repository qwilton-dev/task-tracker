package domain

import "testing"

func TestNewComment(t *testing.T) {
	c, err := NewComment("issue-1", "user-1", "This is a comment")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if c.IssueID != "issue-1" {
		t.Fatalf("expected issue-1, got %s", c.IssueID)
	}
	if c.AuthorID != "user-1" {
		t.Fatalf("expected user-1, got %s", c.AuthorID)
	}
	if c.Body != "This is a comment" {
		t.Fatalf("expected This is a comment, got %s", c.Body)
	}
}

func TestNewComment_EmptyBody(t *testing.T) {
	_, err := NewComment("issue-1", "user-1", "")
	if err != ErrCommentBodyRequired {
		t.Fatalf("expected ErrCommentBodyRequired, got %v", err)
	}
}
