package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"task-tracker/internal/auth"
	"task-tracker/internal/config"
	"task-tracker/internal/events"
	internalhttp "task-tracker/internal/http"
	"task-tracker/internal/http/handler"
	"task-tracker/internal/repository/postgres"
	"task-tracker/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
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

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
		DB:   0,
	})
	defer rdb.Close()

	// Repos
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

	publisher := events.NewPublisher(rdb)
	hub := events.NewHub()

	// Services
	jwtService := auth.NewJWTService(cfg.SecretKey, "task-tracker", "api", 15*time.Minute)
	authService := service.NewAuthService(userRepo, tokenRepo, jwtService, cfg.RefreshTokenExpiresIn)
	workspaceService := service.NewWorkspaceService(workspaceRepo)
	workspaceMemberService := service.NewWorkspaceMemberService(workspaceMemberRepo)
	projectService := service.NewProjectService(projectRepo, workspaceRepo)
	activityEventService := service.NewActivityEventService(activityEventRepo)
	issueService := service.NewIssueService(activityEventService, issueRepo, projectRepo, publisher)
	commentService := service.NewCommentService(commentRepo, issueRepo, publisher)
	labelService := service.NewLabelService(labelRepo, workspaceRepo)
	inviteService := service.NewInviteService(inviteRepo, workspaceMemberRepo)

	// Handlers
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
	go publisher.StartListening(ctx, hub)

	log.Printf("listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}
