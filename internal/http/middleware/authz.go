package middleware

import (
	"net/http"
	"task-tracker/internal/authz"
	"task-tracker/internal/repository"

	"github.com/go-chi/chi/v5"
)

func RequireRole(repo repository.WorkspaceMemberRepository, role authz.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value("user_id").(string)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			slug := chi.URLParam(r, "workspaceSlug")

			userRole, err := repo.GetRole(r.Context(), slug, userID)
			if err != nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if authz.Role(userRole) != role && authz.Role(userRole) != authz.RoleAdmin {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
