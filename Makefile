.PHONY: up down logs run migrate-up migrate-down migrate-status

ifneq (,$(wildcard .env))
include .env
export
endif

ifndef DATABASE_URL
DATABASE_URL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable
export DATABASE_URL
endif

MIGRATE_DIR ?= db/migrations
GOOSE       := $(shell command -v goose 2>/dev/null || echo "go run github.com/pressly/goose/v3/cmd/goose@latest")

up:
	docker compose up -d
	
up-build:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

run:
	go run ./cmd/api

migrate-up:
	$(GOOSE) -dir $(MIGRATE_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	$(GOOSE) -dir $(MIGRATE_DIR) postgres "$(DATABASE_URL)" down

migrate-status:
	$(GOOSE) -dir $(MIGRATE_DIR) postgres "$(DATABASE_URL)" status
