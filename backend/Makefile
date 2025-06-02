include .env
ifneq (,$(wildcard .env.$(APP_ENV)))
    include .env.$(APP_ENV)
endif

DB_URL = postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

.PHONY: help dev api worker migrate-up migrate-down migrate-create docker-up docker-down docker-build test clean install-tools

# Default target
help:
	@echo "Available commands:"
	@echo "  dev          - Start API server with live-reload (development)"
	@echo "  migrate-up   - Run database migrations up"
	@echo "  migrate-down - Run database migrations down"
	@echo "  migrate-create - Create new migration (usage: make migrate-create name=migration_name)"
	@echo "  docker-up    - Start all services with Docker Compose"
	@echo "  docker-down  - Stop all Docker Compose services"
	@echo "  docker-build - Build Docker images"
	@echo "  clean        - Clean build artifacts"
	@echo "  install-tools - Install development tools"

# Development with live-reload
dev:
	@echo "Starting API server with live-reload..."
	./bin/air -c .air.toml

# Database migrations
migrate-up:
	@echo "Running migrations up..."
	migrate -path migrations -database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" up

migrate-down:
	@echo "Running migrations down..."
	migrate -path migrations -database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" down

migrate-create:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=migration_name"; exit 1; fi
	@echo "Creating migration: $(name)"
	migrate create -ext sql -dir migrations $(name)

# Docker commands
docker-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

docker-down:
	@echo "Stopping Docker Compose services..."
	docker-compose down

docker-build:
	@echo "Building Docker images..."
	docker-compose build

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf tmp/
	rm -f build-errors.log