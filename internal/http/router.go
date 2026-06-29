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
	labelHandler *handler.LabelHandler,
	sseHandler *handler.SSEHandler,
	activityHandler *handler.ActivityHandler,
	inviteHandler *handler.InviteHandler,
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

	r.Get("/readyz", func(w nethttp.ResponseWriter, r *nethttp.Request) {
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

		r.With(middleware.RequireAuth(jwtService)).Post("/invites/{token}/accept", inviteHandler.Accept)

		r.Route("/workspaces", func(r chi.Router) {
			r.Use(middleware.RequireAuth(jwtService))
			r.Post("/", workspaceHandler.Create)
			r.Get("/", workspaceHandler.List)

			r.Route("/{workspaceID}", func(r chi.Router) {
				r.Get("/", workspaceHandler.Get)
				r.With(middleware.RequireRoleByWorkspaceID(workspaceMemberRepo, authz.RoleOwner)).Patch("/", workspaceHandler.Update)
				r.With(middleware.RequireRoleByWorkspaceID(workspaceMemberRepo, authz.RoleOwner)).Delete("/", workspaceHandler.Delete)

			r.Route("/projects", func(r chi.Router) {
				r.With(middleware.RequireRoleByWorkspaceID(workspaceMemberRepo, authz.RoleMember)).Post("/", projectHandler.Create)
				r.Get("/", projectHandler.List)
			})
			r.Route("/labels", func(r chi.Router) {
				r.Get("/", labelHandler.ListLabels)
				r.With(middleware.RequireRoleByWorkspaceID(workspaceMemberRepo, authz.RoleMember)).Post("/", labelHandler.CreateLabel)
			})
				r.Route("/members", func(r chi.Router) {
					r.Get("/", workspaceMemberHandler.ListMembers)
					r.With(middleware.RequireRoleByWorkspaceID(workspaceMemberRepo, authz.RoleOwner)).Post("/", workspaceMemberHandler.AddMember)
					r.With(middleware.RequireRoleByWorkspaceID(workspaceMemberRepo, authz.RoleOwner)).Patch("/{memberID}", workspaceMemberHandler.UpdateMemberRole)
					r.With(middleware.RequireRoleByWorkspaceID(workspaceMemberRepo, authz.RoleOwner)).Delete("/{memberID}", workspaceMemberHandler.DeleteMember)
				})
				r.Route("/invites", func(r chi.Router) {
					r.Get("/", inviteHandler.List)
					r.With(middleware.RequireRoleByWorkspaceID(workspaceMemberRepo, authz.RoleMember)).Post("/", inviteHandler.Create)
				})
			})
		})

		r.Route("/projects/{projectID}", func(r chi.Router) {
			r.Use(middleware.RequireAuth(jwtService))
			r.Get("/", projectHandler.Get)
			r.With(middleware.RequireRoleByProjectID(workspaceMemberRepo, authz.RoleMember)).Patch("/", projectHandler.Update)
			r.With(middleware.RequireRoleByProjectID(workspaceMemberRepo, authz.RoleOwner)).Delete("/", projectHandler.Delete)

			r.Route("/issues", func(r chi.Router) {
				r.Get("/", issueHandler.List)
				r.With(middleware.RequireRoleByProjectID(workspaceMemberRepo, authz.RoleMember)).Post("/", issueHandler.Create)
			})
			r.Get("/events", sseHandler.Stream)
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

			r.Route("/activity", func(r chi.Router) {
				r.Get("/", activityHandler.ListByIssue)
			})

			r.Route("/labels", func(r chi.Router) {
				r.Get("/", labelHandler.ListLabelsByIssue)
				r.With(middleware.RequireRoleByIssueID(workspaceMemberRepo, authz.RoleMember)).Post("/{labelID}", labelHandler.AttachLabel)
				r.With(middleware.RequireRoleByIssueID(workspaceMemberRepo, authz.RoleMember)).Delete("/{labelID}", labelHandler.DetachLabel)
			})
		})
	})

	return r
}
