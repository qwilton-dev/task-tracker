package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	DatabaseURL           string
	RedisURL              string
	MigrateDir            string
	RefreshTokenExpiresIn time.Duration
	SecretKey             string
	CORSOrigins           string
}

func Load() (Config, error) {
	_ = godotenv.Load()
	expiresInStr := firstNonEmpty(os.Getenv("REFRESH_TOKEN_EXPIRES_IN"), "604800")

	expiresIn, err := strconv.Atoi(expiresInStr)
	if err != nil {
		expiresIn = 3600
	}
	cfg := Config{
		Port:                  firstNonEmpty(os.Getenv("APP_PORT"), os.Getenv("PORT"), "8080"),
		DatabaseURL:           os.Getenv("DATABASE_URL"),
		RedisURL:              os.Getenv("REDIS_URL"),
		MigrateDir:            firstNonEmpty(os.Getenv("MIGRATE_DIR"), "db/migrations"),
		RefreshTokenExpiresIn: time.Duration(expiresIn) * time.Second,
		SecretKey:             firstNonEmpty(os.Getenv("SECRET_KEY"), "secret"),
		CORSOrigins:           firstNonEmpty(os.Getenv("CORS_ORIGINS"), "http://localhost:8080"),
	}

	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = databaseURL("localhost")
	}

	if cfg.RedisURL == "" {
		cfg.RedisURL = redisURL("localhost")
	}

	return cfg, nil
}

func databaseURL(host string) string {
	user := firstNonEmpty(os.Getenv("POSTGRES_USER"), "task_tracker")
	password := firstNonEmpty(os.Getenv("POSTGRES_PASSWORD"), "task_tracker")
	port := firstNonEmpty(os.Getenv("POSTGRES_PORT"), "5432")
	db := firstNonEmpty(os.Getenv("POSTGRES_DB"), "task_tracker")
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, db,
	)
}

func redisURL(host string) string {
	port := firstNonEmpty(os.Getenv("REDIS_PORT"), "6379")
	return fmt.Sprintf("%s:%s", host, port)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
