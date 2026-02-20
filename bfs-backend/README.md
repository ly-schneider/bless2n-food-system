# BFS Backend

Go HTTP API powering the Bless2n Food System — food ordering, inventory management, payment processing, and device authentication for live event food stations.

## Architecture

The backend follows **Clean Architecture** with strict layer separation and dependency injection via Uber FX:

```
┌──────────────────────────────────────────────┐
│              HTTP Layer                      │
│     Chi Router · Middleware · Handlers       │
├──────────────────────────────────────────────┤
│              Auth Layer                      │
│    Better Auth JWKS · RBAC · Device Auth     │
├──────────────────────────────────────────────┤
│            Service Layer                     │
│   Business Logic · Email · Payments · Orders │
├──────────────────────────────────────────────┤
│           Repository Layer                   │
│        PostgreSQL via pgx · Ent ORM          │
├──────────────────────────────────────────────┤
│             Model Layer                      │
│        Entities · Value Objects              │
└──────────────────────────────────────────────┘
```

### Key Design Decisions

- **Uber FX** for dependency injection — manages all wiring, lifecycle hooks, and graceful shutdown
- **Ent ORM** with Atlas migrations — declarative Go schema generates versioned SQL migrations with CI linting
- **Better Auth JWKS validation** — authentication lives in the Next.js app; the backend only validates tokens via JWKS discovery
- **Payrexx integration** — TWINT payments with HMAC-signed webhook verification
- **Device authentication** — dedicated middleware for IoT/station devices with separate auth flow
- **OpenAPI-first** — API spec is the contract; types are generated via `oapi-codegen`

## Tech Stack

| Category | Technology |
|----------|-----------|
| Language | Go 1.25+ |
| Router | Chi v5 |
| Database | PostgreSQL 18 via pgx v5 |
| ORM / Schema | Ent v0.14 |
| Migrations | Atlas (versioned SQL) |
| DI | Uber FX |
| Auth | Better Auth JWKS, RBAC |
| Payments | Payrexx API (TWINT) |
| Email | Plunk |
| Blob Storage | Azure Blob Storage |
| Logging | Uber Zap (structured JSON) |
| Error Tracking | Sentry |
| Observability | OpenTelemetry |
| API Docs | Swagger / OpenAPI |
| Live Reload | Air |
| Linting | golangci-lint v2 |
| Container | Distroless (non-root) |

## Project Structure

```
bfs-backend/
├── cmd/backend/              Entry point
├── db/
│   ├── migrations/           Atlas versioned SQL migrations
│   ├── provisioning/         Database role/schema setup
│   └── seed/                 Development seed data
├── internal/
│   ├── app/                  Uber FX wiring & DI modules
│   ├── api/                  Generated OpenAPI types
│   ├── auth/                 JWKS validation, RBAC, device auth
│   ├── blobstore/            Azure Blob Storage integration
│   ├── config/               Configuration management
│   ├── database/             Connection pooling & lifecycle
│   ├── generated/            Generated code (Ent, OpenAPI)
│   ├── handler/              HTTP handlers
│   ├── http/                 Chi router setup
│   ├── inventory/            Inventory management
│   ├── middleware/           Common HTTP middleware
│   ├── model/                Data models & entities
│   ├── payrexx/              Payment gateway integration
│   ├── repository/           Data access layer
│   ├── response/             HTTP response helpers
│   ├── schema/               Ent database schema definitions
│   ├── service/              Business logic (orders, payments, email)
│   ├── utils/                Utilities
│   └── version/              Build version info
├── openapi/                  OpenAPI specifications
├── docs/                     Generated Swagger docs
├── test/
│   ├── integration/          Integration tests (requires PostgreSQL)
│   └── fixtures/             Test data
├── docker-compose.yml        PostgreSQL, Mailpit, Azurite
├── Dockerfile                Multi-stage distroless build
├── justfile                  Development commands (Just)
├── Makefile                  Development commands (Make)
├── atlas.hcl                 Atlas migration config
└── .air.toml                 Live reload config
```

## Prerequisites

- **Go 1.25+**
- **Docker & Docker Compose** — for PostgreSQL, Mailpit, and Azurite
- **Make** or **Just** — for running development commands

## Development Setup

```bash
cp .env.example .env

make docker-up          # Start PostgreSQL + Mailpit + Azurite
make migrate-local      # Apply database migrations
make seed               # Seed development data (idempotent)
make dev                # Start with live reload via Air
```

### Service URLs

| Service | URL |
|---------|-----|
| Backend API | http://localhost:8080 |
| Mailpit (email UI) | http://localhost:8025 |
| PostgreSQL | localhost:5432 |
| Azurite (blob storage) | http://localhost:10000 |

## Commands

### Development

```bash
make dev                    # Live reload via Air
make test                   # Unit tests (go test -v -race)
make test-integration       # Integration tests (requires POSTGRES_TEST_DSN)
```

### Docker

```bash
make docker-up              # Start services
make docker-up-dev          # Start with extras (pgAdmin)
make docker-down            # Stop (keep volumes)
make docker-down-v          # Stop and remove volumes
```

### Database

```bash
just migrate                # Apply Atlas migrations
just migrate-status         # Show migration status
just migrate-diff <name>    # Generate migration from Ent schema diff
make seed                   # Seed development data
make psql                   # Open psql shell
```

### Code Quality

```bash
make lint                   # Run golangci-lint
make lint-fix               # Auto-fix lint issues
make fmt                    # Format code
make tidy                   # Tidy go modules
```

### Code Generation

```bash
just generate               # Run all generators
just generate-ent           # Ent ORM code
just generate-api           # OpenAPI types
just swag                   # Swagger docs
```

## Authentication & Authorization

Authentication is managed by **Better Auth** in the Next.js web app. The backend validates requests by:

1. Fetching JWKS from `BETTER_AUTH_URL/api/auth/jwks`
2. Verifying Bearer tokens in the session middleware
3. Enforcing role-based access control (admin vs. customer)
4. Supporting device authentication for IoT/station devices via a separate middleware

## Payment Processing

Payrexx handles payment gateway creation for TWINT (Swiss mobile payment):

- Payment gateways are created via the Payrexx API
- Webhook notifications are verified using HMAC signatures
- Payment status updates are processed asynchronously
- Currency: CHF

## Environment Variables

Key configuration (see `.env.example` for full list):

| Variable | Purpose |
|----------|---------|
| `DATABASE_URL` | PostgreSQL connection string |
| `BETTER_AUTH_URL` | Next.js app URL for JWKS discovery |
| `PAYREXX_INSTANCE` | Payrexx instance name |
| `PAYREXX_API_SECRET` | Payrexx request signing |
| `PAYREXX_WEBHOOK_SECRET` | Webhook verification |
| `PUBLIC_BASE_URL` | Frontend URL for redirects |
| `PLUNK_API_KEY` | Transactional email service |
| `SECURITY_TRUSTED_ORIGINS` | CORS allowed origins |
