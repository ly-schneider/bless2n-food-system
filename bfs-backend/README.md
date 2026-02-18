# Bless2n Food System Backend

A Go HTTP backend for the Bless2n Food System -- food ordering, inventory, POS, and device management.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Prerequisites](#prerequisites)
- [Development Setup](#development-setup)
- [Available Commands](#available-commands)
- [Project Structure](#project-structure)
- [Testing](#testing)

## Architecture Overview

The backend follows **Clean Architecture** with clear layer separation:

```
+------------------------------------------+
|             HTTP Layer                    |
|   (Chi Router, Middleware, Handlers)      |
+------------------------------------------+
|            Service Layer                  |
|     (Business Logic, Email, Payments)     |
+------------------------------------------+
|           Postgres Layer                  |
|         (Data Access, pgx/GORM)          |
+------------------------------------------+
|             Model Layer                   |
|        (Entities, Value Objects)          |
+------------------------------------------+
```

### Key Architectural Decisions

- **Dependency Injection**: Uber FX manages all dependencies and lifecycle
- **Repository Pattern**: Clean abstraction over PostgreSQL via pgx and GORM
- **Model-Driven Design**: Typed models with clear separation from persistence
- **Better Auth Integration**: Authentication delegated to the Next.js app via Better Auth; the backend validates sessions through JWKS
- **Payrexx Payments**: Payment processing via Payrexx API with webhook verification
- **Atlas Migrations**: Database schema managed with versioned SQL migrations via [Atlas](https://atlasgo.io/) with Ent integration

## Prerequisites

### Required
- **Go 1.25+**
- **Docker & Docker Compose**: For running PostgreSQL and Mailpit
- **Make**: For running development commands

### Optional
- **Air**: For live reload during development (installed automatically via `make tools`)
- **psql**: For direct database access

## Development Setup

```bash
# 1. Environment setup
cp .env.example .env
# Edit .env if needed (defaults work for local Docker PostgreSQL)

# 2. Start supporting services (PostgreSQL + Mailpit)
make docker-up

# 3. Run database migrations
make migrate-local

# 4. Seed development data (optional)
make seed

# 5. Run backend with live reload
make dev
```

### Service URLs

When services are running:

- **Backend API**: http://localhost:8080
- **Mailpit UI** (email testing): http://localhost:8025
- **PostgreSQL**: localhost:5432

## Available Commands

Run `make` or `make help` to see all available commands with descriptions.

### Key Commands
```bash
# Development
make dev                # Start with live-reload (via Air)
make test               # Run unit tests
make test-integration   # Run integration tests

# Docker
make docker-up          # Start PostgreSQL + Mailpit
make docker-up-dev      # Start all dev services (+ pgAdmin)
make docker-down        # Stop services (keep volumes)
make docker-down-v      # Stop services and remove volumes (DATA LOSS)

# Database
just migrate            # Apply Atlas migrations against local PostgreSQL
just migrate-status     # Show migration status
just migrate-diff name  # Generate a new migration from Ent schema diff
make seed               # Seed dev data (idempotent)
make psql               # Open psql shell to local database

# Code quality
make lint               # Run linters
make lint-fix           # Auto-fix linting issues
make fmt                # Format code
make tidy               # Tidy go modules
```

## Project Structure

```
bfs-backend/
├── cmd/backend/           # Application entry point
├── db/
│   ├── migrations/        # Atlas versioned SQL migrations
│   ├── provisioning/      # Database role/schema provisioning SQL
│   └── seed/              # Development seed data SQL files
├── internal/
│   ├── app/               # Application wiring & DI setup (Uber FX)
│   ├── auth/              # Auth middleware (Better Auth JWKS, RBAC, device auth)
│   ├── config/            # Configuration management
│   ├── handler/           # HTTP handlers
│   ├── http/              # Chi router setup
│   ├── middleware/        # Common HTTP middleware
│   ├── model/             # Data models / entities
│   ├── observability/     # OpenTelemetry setup
│   ├── payrexx/           # Payrexx payment integration
│   ├── postgres/          # PostgreSQL repositories (pgx / GORM)
│   ├── response/          # HTTP response helpers
│   ├── service/           # Business logic
│   └── utils/             # Utility functions
├── test/                  # Integration and E2E tests
├── .air.toml              # Live reload config
├── .env.example           # Environment template
├── docker-compose.yml     # Docker services (PostgreSQL, Mailpit, etc.)
├── Dockerfile             # Container definition
├── go.mod & go.sum        # Go dependencies
├── Makefile               # Development commands
└── README.md
```

## Testing

### Test Structure

```
bfs-backend/
└── test/
    ├── integration/       # Integration tests (requires PostgreSQL)
    └── fixtures/          # Test data and fixtures
```

### Running Tests

```bash
make test               # Unit tests: go test -v -race ./internal/...
make test-integration   # Integration tests (requires POSTGRES_TEST_DSN)
```
