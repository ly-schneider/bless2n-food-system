include .env
ifneq (,$(wildcard .env.$(APP_ENV)))
    include .env.$(APP_ENV)
endif

.PHONY: help dev docker-up docker-down docker-build flyway-migrate flyway-info flyway-validate clean

# Default target
help:
	@echo "Available commands:"
	@echo "  dev              - Start API server with live-reload (development)"
	@echo "  docker-up        - Start all services with Docker Compose"
	@echo "  docker-down      - Stop all Docker Compose services"
	@echo "  docker-build     - Build Docker images"
	@echo "  flyway-migrate   - Run Flyway migrations"
	@echo "  flyway-info      - Get Flyway migration info"
	@echo "  flyway-validate  - Validate Flyway migrations"
	@echo "  clean            - Clean build artifacts"

# Development with live-reload
dev:
	@echo "Starting API server with live-reload..."
	./bin/air -c .air.toml

# Docker commands
docker-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

docker-down:
	@echo "Stopping Docker Compose services..."
	docker-compose down --volumes --remove-orphans

docker-build:
	@echo "Building Docker images..."
	docker-compose build

# Flyway commands
flyway-migrate:
	@echo "Running Flyway migrations..."
	docker run --rm \
		--network rentro-backend_backend \
		-v $(PWD)/db/migration:/flyway/sql \
		flyway/flyway:latest \
		-url=jdbc:postgresql://postgres:5432/$(POSTGRES_DB) \
		-user=$(POSTGRES_USER) \
		-password=$(POSTGRES_PASSWORD) \
		-locations=filesystem:/flyway/sql \
		migrate

flyway-info:
	@echo "Getting Flyway migration info..."
	docker run --rm \
		--network rentro-backend_backend \
		-v $(PWD)/db/migration:/flyway/sql \
		flyway/flyway:latest \
		-url=jdbc:postgresql://postgres:5432/$(POSTGRES_DB) \
		-user=$(POSTGRES_USER) \
		-password=$(POSTGRES_PASSWORD) \
		-locations=filesystem:/flyway/sql \
		info

flyway-validate:
	@echo "Validating Flyway migrations..."
	docker run --rm \
		--network rentro-backend_backend \
		-v $(PWD)/db/migration:/flyway/sql \
		flyway/flyway:latest \
		-url=jdbc:postgresql://postgres:5432/$(POSTGRES_DB) \
		-user=$(POSTGRES_USER) \
		-password=$(POSTGRES_PASSWORD) \
		-locations=filesystem:/flyway/sql \
		validate

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf tmp/
	rm -f build-errors.log