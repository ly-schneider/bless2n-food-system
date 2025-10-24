# Use bash with safer defaults
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c

# --- Binaries dir + PATH ---
BIN := $(CURDIR)/bin
export PATH := $(BIN):$(PATH)

# Tool versions (pin these)
GOLANGCI_LINT_VERSION := v2.5.0
SWAG_VERSION := v1.16.6
AIR_VERSION := v1.63.0

# Let callers tweak linter flags in CI/local runs
GOLANGCI_LINT_ARGS ?= --timeout=3m --max-issues-per-linter=0 --max-same-issues=0

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
AIR ?= $(BIN)/air

# Optional flags toggled via env
FORCE ?=
FORCE_FLAG := $(if $(filter 1 true yes on,$(FORCE)),-force,)

# Stripe
WEBHOOK_URL ?= http://localhost:$(APP_PORT)/v1/payments/webhook

.PHONY: help tools lint lint-fix fmt tidy tidy-check print-tools swag dev docker-up docker-up-backend docker-down docker-down-backend docker-down-v docker-down-v-backend docker-restart docker-restart-backend clean test db-seed db-reset stripe-listen

help:
	@echo "Available commands:"
	@echo "  tools                  - Install/update required tools (air, swag, golangci-lint)"
	@echo "  lint                   - Run linters"
	@echo "  lint-fix               - Run linters with auto-fix"
	@echo "  fmt                    - Run go fmt"
	@echo "  tidy                   - Run go mod tidy"
	@echo "  tidy-check             - Check go.mod/go.sum are tidy"
	@echo "  print-tools            - Print installed tool versions"
	@echo "  swag                   - Create swagger documentation"
	@echo "  dev                    - Start API server with live-reload (AIR)"
	@echo "  docker-up              - Start Mongo + Mongo Express"
	@echo "  docker-up-backend      - Start Mongo + Mongo Express + backend (profile)"
	@echo "  docker-down            - Stop services (keep volumes)"
	@echo "  docker-down-backend    - Stop services incl. backend (keep volumes)"
	@echo "  docker-down-v          - Stop services and REMOVE VOLUMES (DATA LOSS)"
	@echo "  docker-down-v-backend  - Stop services incl. backend and REMOVE VOLUMES (DATA LOSS)"
	@echo "  docker-restart         - Restart services (no backend)"
	@echo "  docker-restart-backend - Restart services incl. backend"
	@echo "  clean                  - Clean build artifacts"
	@echo "  test                   - Run unit tests"
	@echo "  db-seed                - Seed MongoDB with development data"
	@echo "  db-reset               - Reset and seed MongoDB with development data"
	@echo "  stripe-listen          - Start Stripe CLI and forward events to local webhook"

tools: $(BIN)/golangci-lint $(BIN)/swag $(BIN)/air
	@echo "Installing/updating tools... done."

$(BIN)/golangci-lint:
	@mkdir -p $(BIN)
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION) → $(BIN)"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
	| sh -s -- -b $(BIN) $(GOLANGCI_LINT_VERSION)

$(BIN)/swag:
	@mkdir -p $(BIN)
	@echo "Installing swag $(SWAG_VERSION) → $(BIN)"
	@GOBIN=$(BIN) go install github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION)

$(BIN)/air:
	@mkdir -p $(BIN)
	@echo "Installing air $(AIR_VERSION) → $(BIN)"
	@GOBIN=$(BIN) go install github.com/air-verse/air@$(AIR_VERSION)

lint: $(BIN)/golangci-lint
	@echo "Running linters..."
	@GOTOOLCHAIN=auto GOWORK=off $(BIN)/golangci-lint run $(GOLANGCI_LINT_ARGS) ./...

lint-fix: $(BIN)/golangci-lint
	@echo "Running linters with auto-fix..."
	@GOTOOLCHAIN=auto GOWORK=off $(BIN)/golangci-lint run --fix $(GOLANGCI_LINT_ARGS) ./...

fmt:
	@echo "Formatting..."
	@go fmt ./...

tidy:
	@echo "Tidying modules..."
	@GOWORK=off go mod tidy

tidy-check:
	@echo "Checking go.mod/go.sum are tidy..."
	@GOWORK=off go mod tidy
	@git diff --exit-code -- go.mod go.sum || (echo "Run 'make tidy' to update go.mod/go.sum"; exit 1)

print-tools:
	@echo "golangci-lint: $$($(BIN)/golangci-lint version | head -n1 || echo 'not installed')"
	@echo "swag:          $$($(BIN)/swag --version 2>/dev/null || echo 'not installed')"
	@echo "air:           $$($(BIN)/air --version 2>/dev/null || echo 'not installed')"

swag: $(BIN)/swag
	@echo "Creating swagger documentation..."
	@$(BIN)/swag init -g cmd/backend/main.go --parseInternal --parseDependency --generatedTime=false

dev: $(BIN)/air
	@echo "Starting API server with live-reload..."
	@$(AIR) -c .air.toml

docker-up:
	@echo "Starting services (Mongo + Mongo Express)"
	@$(DC) up -d

docker-up-backend:
	@echo "Starting services including backend (profile: backend)"
	@$(DC) --profile backend up -d

docker-down:
	@echo "Stopping services (keeping volumes)..."
	@$(DC) down --remove-orphans

docker-down-backend:
	@echo "Stopping services including backend (keeping volumes)..."
	@$(DC) --profile backend down --remove-orphans

docker-down-v:
	@echo "Stopping services and removing volumes (DATA LOSS)..."
	@$(DC) down --volumes --remove-orphans

docker-down-v-backend:
	@echo "Stopping services including backend and removing volumes (DATA LOSS)..."
	@$(DC) --profile backend down --volumes --remove-orphans

docker-restart:
	@echo "Restarting services (no backend)"
	@$(MAKE) docker-down
	@$(MAKE) docker-up

docker-restart-backend:
	@echo "Restarting services including backend"
	@$(MAKE) docker-down
	@$(MAKE) docker-up-backend

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf tmp/
	@rm -f build-errors.log
	@rm -rf docs/

test:
	@echo "Running unit tests..."
	@go test -v -race ./internal/...

db-seed:
	go run ./cmd/ctl $(FORCE_FLAG)

db-reset:
	go run ./cmd/ctl -reset $(FORCE_FLAG)

stripe-listen:
	@command -v stripe >/dev/null 2>&1 || { \
	  echo "Stripe CLI not found. Install: https://stripe.com/docs/stripe-cli"; exit 1; }
	@echo "Starting Stripe webhook listener → $(WEBHOOK_URL)"
	@echo "Note: Copy the signing secret into STRIPE_WEBHOOK_SECRET in .env"
	@stripe listen --forward-to $(WEBHOOK_URL)
