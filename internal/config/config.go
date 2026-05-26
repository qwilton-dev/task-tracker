package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	MigrateDir  string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		Port:        firstNonEmpty(os.Getenv("APP_PORT"), os.Getenv("PORT"), "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		MigrateDir:  firstNonEmpty(os.Getenv("MIGRATE_DIR"), "db/migrations"),
	}

	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = databaseURL("localhost")
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

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
