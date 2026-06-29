package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"task-tracker/internal/auth"
	"task-tracker/internal/config"
	"task-tracker/internal/events"
	internalhttp "task-tracker/internal/http"
	"task-tracker/internal/http/handler"
	"task-tracker/internal/repository/postgres"
	"task-tracker/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	hub := events.NewHub()

	userRepo := postgres.NewUserRepository(pool)
	tokenRepo := postgres.NewTokenRepository(pool)
	workspaceRepo := postgres.NewWorkspaceRepository(pool)
	workspaceMemberRepo := postgres.NewWorkspaceMemberRepository(pool)
	projectRepo := postgres.NewProjectRepository(pool)
	issueRepo := postgres.NewIssueRepository(pool)
	commentRepo := postgres.NewCommentRepository(pool)
	labelRepo := postgres.NewLabelRepository(pool)
	activityEventRepo := postgres.NewActivityEventRepository(pool)
	inviteRepo := postgres.NewInviteRepository(pool)

	jwtService := auth.NewJWTService(cfg.SecretKey, "task-tracker", "api", 15*time.Minute)
	authService := service.NewAuthService(userRepo, tokenRepo, jwtService, cfg.RefreshTokenExpiresIn)
	workspaceService := service.NewWorkspaceService(workspaceRepo)
	workspaceMemberService := service.NewWorkspaceMemberService(workspaceMemberRepo)
	projectService := service.NewProjectService(projectRepo)
	activityEventService := service.NewActivityEventService(activityEventRepo)
	issueService := service.NewIssueService(activityEventService, issueRepo, projectRepo, hub)
	commentService := service.NewCommentService(commentRepo, issueRepo, activityEventService, hub)
	labelService := service.NewLabelService(labelRepo, activityEventService)
	inviteService := service.NewInviteService(inviteRepo, workspaceMemberRepo)

	authHandler := handler.NewAuthHandler(authService)
	workspaceHandler := handler.NewWorkspaceHandler(workspaceService)
	workspaceMemberHandler := handler.NewWorkspaceMemberHandler(workspaceMemberService)
	projectHandler := handler.NewProjectHandler(projectService)
	issueHandler := handler.NewIssueHandler(issueService)
	commentHandler := handler.NewCommentHandler(commentService)
	labelHandler := handler.NewLabelHandler(labelService)
	activityHandler := handler.NewActivityHandler(activityEventService)
	inviteHandler := handler.NewInviteHandler(inviteService)
	sseHandler := handler.NewSSEHandler(hub)

	r := internalhttp.NewRouter(authHandler, workspaceHandler, jwtService, workspaceMemberHandler, projectHandler, issueHandler, commentHandler, labelHandler, sseHandler, activityHandler, inviteHandler, workspaceMemberRepo, cfg.CORSOrigins)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
