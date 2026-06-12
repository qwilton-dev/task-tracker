package http

import (
	"encoding/json"
	nethttp "net/http"

	"task-tracker/internal/auth"
	"task-tracker/internal/authz"
	"task-tracker/internal/http/handler"
	"task-tracker/internal/http/middleware"
	"task-tracker/internal/repository"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(authHandler *handler.AuthHandler,
	workspaceHandler *handler.WorkspaceHandler,
	jwtService *auth.JWTService,
	workspaceMemberHandler *handler.WorkspaceMemberHandler,
	projectHandler *handler.ProjectHandler,
	issueHandler *handler.IssueHandler,
	commentHandler *handler.CommentHandler,
	workspaceMemberRepo repository.WorkspaceMemberRepository,
	corsOrigins string) nethttp.Handler {
	r := chi.NewRouter()

	r.Use(middleware.CORS(corsOrigins))
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.StripSlashes)

	r.Get("/healthz", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	r.Handle("/*", nethttp.FileServer(nethttp.Dir("web")))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/logout", authHandler.Logout)
		})
		r.With(middleware.RequireAuth(jwtService)).Get("/me", authHandler.Me)

		r.Route("/workspaces", func(r chi.Router) {
			r.Use(middleware.RequireAuth(jwtService))
			r.Post("/", workspaceHandler.Create)
			r.Get("/", workspaceHandler.List)

			r.Route("/{workspaceSlug}", func(r chi.Router) {
				r.Route("/projects", func(r chi.Router) {
					r.With(middleware.RequireRoleBySlug(workspaceMemberRepo, authz.RoleMember)).Post("/", projectHandler.Create)
					r.Get("/", projectHandler.List)
				})
			})
		})

		r.Route("/workspace-members", func(r chi.Router) {
			r.Use(middleware.RequireAuth(jwtService))
			r.Post("/", workspaceMemberHandler.AddMember)
			r.Get("/", workspaceMemberHandler.ListMembers)
			r.Delete("/", workspaceMemberHandler.DeleteMember)
			r.Patch("/", workspaceMemberHandler.UpdateMemberRole)
		})

		r.Route("/projects/{projectID}/issues", func(r chi.Router) {
			r.Use(middleware.RequireAuth(jwtService))
			r.Get("/", issueHandler.List)
			r.With(middleware.RequireRoleByProjectID(workspaceMemberRepo, authz.RoleMember)).Post("/", issueHandler.Create)
		})

		r.Route("/issues/{issueID}", func(r chi.Router) {
			r.Use(middleware.RequireAuth(jwtService))
			r.Get("/", issueHandler.Get)
			r.With(middleware.RequireRoleByIssueID(workspaceMemberRepo, authz.RoleMember)).Patch("/", issueHandler.Update)
			r.With(middleware.RequireRoleByIssueID(workspaceMemberRepo, authz.RoleMember)).Delete("/", issueHandler.Delete)
			r.With(middleware.RequireRoleByIssueID(workspaceMemberRepo, authz.RoleMember)).Patch("/move", issueHandler.Move)

			r.Route("/comments", func(r chi.Router) {
				r.Get("/", commentHandler.List)
				r.With(middleware.RequireRoleByIssueID(workspaceMemberRepo, authz.RoleMember)).Post("/", commentHandler.Create)
			})
		})
	})

	return r
}
