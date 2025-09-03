# Use bash with safer defaults
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c

AIR ?= go tool air

# Load env files (don't error if missing)
-include .env
ifneq (,$(wildcard .env.$(APP_ENV)))
  -include .env.$(APP_ENV)
endif

# Export all vars from the Makefile to subcommands (AIR, docker, etc.)
.EXPORT_ALL_VARIABLES:

# Default to LOCAL for everything (e.g., `make dev`)
MONGO_URI ?= $(MONGO_URI_LOCAL)
SMTP_HOST ?= $(SMTP_HOST_LOCAL)
SMTP_PORT ?= $(SMTP_PORT_LOCAL)

# For compose-related targets with backend, override to DOCKER just for those recipes
docker-up-backend docker-restart-backend: MONGO_URI := $(MONGO_URI_DOCKER)
docker-up-backend docker-restart-backend: SMTP_HOST := $(SMTP_HOST_DOCKER)
docker-up-backend docker-restart-backend: SMTP_PORT := $(SMTP_PORT_DOCKER)

# Default target
.DEFAULT_GOAL := help

# Shortcuts
DC  := docker compose
AIR ?= ./bin/air

.PHONY: help tools swag dev docker-up docker-up-backend docker-down docker-down-v docker-restart docker-restart-backend clean logs ps mongo-shell test test-e2e test-coverage test-setup test-teardown

help:
	@echo "Available commands:"
	@echo "  tools                  - Install/update required tools (air, swag)"
	@echo "  swag                   - Create swagger documentation"
	@echo "  dev                    - Start API server with live-reload (AIR)"
	@echo "  docker-up              - Start Mongo + Mongo Express"
	@echo "  docker-up-backend      - Start Mongo + Mongo Express + backend (profile)"
	@echo "  docker-down            - Stop services (keep volumes)"
	@echo "  docker-down-v          - Stop services and REMOVE VOLUMES (DATA LOSS)"
	@echo "  docker-restart         - Restart services (no backend)"
	@echo "  docker-restart-backend - Restart services incl. backend"
	@echo "  logs                   - Tail compose logs"
	@echo "  ps                     - Show compose services"
	@echo "  mongo-shell            - Open mongosh in the mongo container"
	@echo "  clean                  - Remove build artifacts"
	@echo "  test                   - Run unit tests"
	@echo "  test-e2e               - Run end-to-end tests"
	@echo "  test-coverage          - Run tests with coverage"
	@echo "  test-setup             - Setup test infrastructure"
	@echo "  test-teardown          - Teardown test infrastructure"

tools:
	@echo "Installing/updating tools..."
	go get -tool github.com/swaggo/swag/cmd/swag@latest
	go get -tool github.com/air-verse/air@latest
	go mod tidy

swag:
	@echo "Creating swagger documentation..."
	go tool swag init -g cmd/backend/main.go --parseInternal --parseDependency --generatedTime=false

dev: 
	@echo "Starting API server with live-reload..."
	$(AIR) -c .air.toml

docker-up:
	@echo "Starting services (Mongo + Mongo Express)"
	$(DC) up -d

docker-up-backend:
	@echo "Starting services including backend (profile: backend)"
	$(DC) --profile backend up -d

docker-down:
	@echo "Stopping services (keeping volumes)..."
	$(DC) down --remove-orphans

docker-down-v:
	@echo "Stopping services and removing volumes (DATA LOSS)..."
	$(DC) down --volumes --remove-orphans

docker-restart:
	@echo "Restarting services (no backend)"
	$(MAKE) docker-down
	$(MAKE) docker-up

docker-restart-backend:
	@echo "Restarting services including backend"
	$(MAKE) docker-down
	$(MAKE) docker-up-backend

logs:
	$(DC) logs -f --tail=200

ps:
	$(DC) ps

mongo-shell:
	$(DC) exec -it mongo mongosh -u $$MONGO_ROOT_USERNAME -p $$MONGO_ROOT_PASSWORD --authenticationDatabase admin
	
clean:
	@echo "Cleaning build artifacts..."
	rm -rf tmp/
	rm -f build-errors.log

test:
	@echo "Running unit tests..."
	go test -v -race ./internal/...

test-e2e:
	@echo "Running end-to-end tests..."
	$(MAKE) -C test test-e2e

test-coverage:
	@echo "Running tests with coverage..."
	$(MAKE) -C test test-all
	$(MAKE) -C test test-coverage

test-setup:
	@echo "Setting up test infrastructure..."
	$(MAKE) -C test test-setup

test-teardown:
	@echo "Tearing down test infrastructure..."
	$(MAKE) -C test test-teardown
