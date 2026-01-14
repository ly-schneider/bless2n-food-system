SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c

BIN := $(CURDIR)/bin
export PATH := $(BIN):$(PATH)

GOLANGCI_LINT_VERSION := v2.5.0
SWAG_VERSION := v1.16.6
AIR_VERSION := v1.63.0

GOLANGCI_LINT_ARGS ?= --timeout=3m --max-issues-per-linter=0 --max-same-issues=0

-include .env
ifneq (,$(wildcard .env.$(APP_ENV)))
  -include .env.$(APP_ENV)
endif

.EXPORT_ALL_VARIABLES:
.DEFAULT_GOAL := help

.PHONY: help tools lint lint-fix fmt tidy tidy-check print-tools swag dev docker-up docker-down docker-down-v clean test migrate migrate-info migrate-validate migrate-repair migrate-clean

help:
	@echo "Development:"
	@echo "  tools            - Install/update tools (air, swag, golangci-lint)"
	@echo "  dev              - Start API server with live-reload"
	@echo "  test             - Run unit tests"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint             - Run linters"
	@echo "  lint-fix         - Run linters with auto-fix"
	@echo "  fmt              - Format code with go fmt"
	@echo "  tidy             - Tidy go modules"
	@echo "  tidy-check       - Check if go.mod/go.sum are tidy"
	@echo "  swag             - Generate Swagger documentation"
	@echo "  print-tools      - Show installed tool versions"
	@echo ""
	@echo "Docker:"
	@echo "  docker-up        - Start services (Neon Local + Mongo)"
	@echo "  docker-down      - Stop services (keep volumes)"
	@echo "  docker-down-v    - Stop services and remove volumes (DATA LOSS)"
	@echo ""
	@echo "Database:"
	@echo "  migrate          - Run pending migrations"
	@echo "  migrate-info     - Show migration status"
	@echo "  migrate-validate - Validate applied migrations"
	@echo "  migrate-repair   - Repair migration history"
	@echo "  migrate-clean    - Drop all objects (DATA LOSS)"
	@echo ""
	@echo "Other:"
	@echo "  clean          - Clean build artifacts"

tools: $(BIN)/golangci-lint $(BIN)/swag $(BIN)/air

$(BIN)/golangci-lint:
	@mkdir -p $(BIN)
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
	| sh -s -- -b $(BIN) $(GOLANGCI_LINT_VERSION)

$(BIN)/swag:
	@mkdir -p $(BIN)
	@echo "Installing swag $(SWAG_VERSION)..."
	@GOBIN=$(BIN) go install github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION)

$(BIN)/air:
	@mkdir -p $(BIN)
	@echo "Installing air $(AIR_VERSION)..."
	@GOBIN=$(BIN) go install github.com/air-verse/air@$(AIR_VERSION)

lint: $(BIN)/golangci-lint
	@GOTOOLCHAIN=auto GOWORK=off $(BIN)/golangci-lint run $(GOLANGCI_LINT_ARGS) ./...

lint-fix: $(BIN)/golangci-lint
	@GOTOOLCHAIN=auto GOWORK=off $(BIN)/golangci-lint run --fix $(GOLANGCI_LINT_ARGS) ./...

fmt:
	@go fmt ./...

tidy:
	@GOWORK=off go mod tidy

tidy-check:
	@GOWORK=off go mod tidy
	@git diff --exit-code -- go.mod go.sum || (echo "Error: go.mod/go.sum are not tidy. Run 'make tidy'"; exit 1)

print-tools:
	@echo "golangci-lint: $$($(BIN)/golangci-lint version | head -n1 || echo 'not installed')"
	@echo "swag:          $$($(BIN)/swag --version 2>/dev/null || echo 'not installed')"
	@echo "air:           $$($(BIN)/air --version 2>/dev/null || echo 'not installed')"

swag: $(BIN)/swag
	@$(BIN)/swag init -g cmd/backend/main.go --parseInternal --parseDependency --generatedTime=false

dev: $(BIN)/air
	@$(BIN)/air -c .air.toml

docker-up:
	@docker compose up -d

docker-down:
	@docker compose down --remove-orphans

docker-down-v:
	@docker compose down --volumes --remove-orphans

clean:
	@rm -rf tmp/ build-errors.log docs/

test:
	@go test -v -race ./internal/...

migrate:
	@docker compose --profile migrate up flyway

migrate-info:
	@docker compose --profile migrate run --rm flyway info

migrate-validate:
	@docker compose --profile migrate run --rm flyway validate

migrate-repair:
	@docker compose --profile migrate run --rm flyway repair

migrate-clean:
	@echo "⚠️  WARNING: This will DROP ALL OBJECTS in the database!"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	@docker compose --profile migrate run --rm flyway clean
