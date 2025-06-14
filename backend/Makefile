include .env
ifneq (,$(wildcard .env.$(APP_ENV)))
    include .env.$(APP_ENV)
endif

.PHONY: help dev podman-up podman-down podman-build flyway-migrate flyway-info flyway-validate clean

# Default target
help:
	@echo "Available commands:"
	@echo "  dev              - Start API server with live-reload (development)"
	@echo "  podman-up        - Start all services with Podman Compose"
	@echo "  podman-down      - Stop all Podman Compose services"
	@echo "  podman-build     - Build Podman images"
	@echo "  flyway-migrate   - Run Flyway migrations"
	@echo "  flyway-info      - Get Flyway migration info"
	@echo "  flyway-validate  - Validate Flyway migrations"
	@echo "  clean            - Clean build artifacts"

# Development with live-reload
dev:
	@echo "Starting API server with live-reload..."
	./bin/air -c .air.toml

# Docker commands
podman-up:
	@echo "Starting services with Podman Compose..."
	podman compose up -d

podman-down:
	@echo "Stopping Podman Compose services..."
	podman compose down --volumes --remove-orphans

podman-build:
	@echo "Building Podman images..."
	podman compose build

# Flyway commands
flyway-migrate:
	@echo "Running Flyway migrations..."
	podman run --rm \
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
	podman run --rm \
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
	podman run --rm \
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