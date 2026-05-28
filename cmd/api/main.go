package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"task-tracker/internal/auth"
	"task-tracker/internal/config"
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

	userRepo := postgres.NewUserRepository(pool)
	tokenRepo := postgres.NewTokenRepository(pool)
	jwtService := auth.NewJWTService(cfg.SecretKey, "task-tracker", "api", 15*time.Minute)
	authService := service.NewAuthService(userRepo, tokenRepo, jwtService)
	authHandler := handler.NewAuthHandler(authService)
	r := internalhttp.NewRouter(authHandler, jwtService)

	log.Printf("listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}
