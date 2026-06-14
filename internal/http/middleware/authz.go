package middleware

import (
	"context"
	"net/http"
	"task-tracker/internal/authz"
	"task-tracker/internal/repository"

	"github.com/go-chi/chi/v5"
)

type RoleResolver func(ctx context.Context, r *http.Request, userID string) (string, error)

func RequireRole(repo repository.WorkspaceMemberRepository, minRole authz.Role, resolve RoleResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFrom(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userRole, err := resolve(r.Context(), r, userID)
			if err != nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if !authz.Role(userRole).AtLeast(minRole) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireRoleByWorkspaceID(repo repository.WorkspaceMemberRepository, minRole authz.Role) func(http.Handler) http.Handler {
	return RequireRole(repo, minRole, func(ctx context.Context, r *http.Request, userID string) (string, error) {
		workspaceID := chi.URLParam(r, "workspaceID")
		return repo.GetRole(ctx, workspaceID, userID)
	})
}

func RequireRoleByProjectID(repo repository.WorkspaceMemberRepository, minRole authz.Role) func(http.Handler) http.Handler {
	return RequireRole(repo, minRole, func(ctx context.Context, r *http.Request, userID string) (string, error) {
		projectID := chi.URLParam(r, "projectID")
		return repo.GetRoleByProjectID(ctx, projectID, userID)
	})
}

func RequireRoleByIssueID(repo repository.WorkspaceMemberRepository, minRole authz.Role) func(http.Handler) http.Handler {
	return RequireRole(repo, minRole, func(ctx context.Context, r *http.Request, userID string) (string, error) {
		issueID := chi.URLParam(r, "issueID")
		return repo.GetRoleByIssueID(ctx, issueID, userID)
	})
}
