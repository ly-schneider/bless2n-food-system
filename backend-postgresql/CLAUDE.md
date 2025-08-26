# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

**Local Development Workflow:**
```bash
make docker-up      # Start PostgreSQL, Redis, and Asynqmon
make dev           # Start API with live-reload using Air (development only)
make docker-down   # Stop all services with cleanup
```

**Database Migrations:**
```bash
make flyway-migrate    # Run database migrations
make flyway-info      # Check migration status
make flyway-validate  # Validate migration files
```

**Docker Operations:**
```bash
make docker-build     # Useful for building the worker image
make clean           # Clean build artifacts and tmp/
```

**Service URLs:**
- API: http://localhost:8080
- Asynqmon (job monitoring): http://localhost:8081
- PostgreSQL: localhost:5432
- Redis: localhost:6379

## Architecture Overview

This Go backend follows **Clean Architecture** with clear layer separation:

**Domain Layer** (`internal/domain/`): Business entities with NanoID14 identifiers and GORM hooks for auto-ID generation.

**Service Layer** (`internal/service/`): Business logic that coordinates between repositories and handles complex operations.

**Repository Layer** (`internal/repository/`): Data access with interfaces for testability using GORM.

**Handler Layer** (`internal/handler/`): HTTP handlers that adapt between HTTP requests and service calls.

**Infrastructure**: HTTP middleware, database connections, configuration, and background job processing.

## Key Patterns

**ID Generation**: Uses NanoID14 (14 characters) instead of UUIDs for 61% space savings. IDs are auto-generated in domain models via `BeforeCreate` GORM hooks.

**Dependency Injection**: Uses Uber FX throughout. New services must be registered in `internal/app/` provider functions.

**Background Jobs**: Asynq-based system with API enqueuing jobs and separate Worker service processing them.

**Authentication**: JWT access tokens (15 minutes) + Argon2 hashed refresh tokens (30 days) with mobile/browser detection.

**Database**: PostgreSQL 17 with Row Level Security (RLS), Flyway migrations, and BRIN indexes for time-series data.

**Error Handling**: Structured errors via `internal/apperrors/` with consistent HTTP status mapping.

## Adding New Features

1. **Domain Model**: Create entity in `internal/domain/` with NanoID14 and GORM hooks
2. **Repository**: Add interface and implementation in `internal/repository/`
3. **Service**: Implement business logic in `internal/service/`
4. **Handler**: Create HTTP handlers in `internal/handler/`
5. **Registration**: Wire up dependencies in `internal/app/` provider functions
6. **Migration**: Add versioned SQL file in `db/migrations/V{number}__{description}.sql`

## Database Migration Patterns

- Use Flyway placeholders for environment-specific values: `${db_admin_user}`, `${app_user}`
- Follow naming: `V{version}__{description}.sql`
- Include proper GORM column definitions for custom types like `nano_id`, `argon2_hash`
- Consider RLS policies and BRIN indexes for time-series columns

## Testing Notes

Currently no test files exist. When implementing tests:
- Air excludes `_test.go` files from live-reload
- Use standard Go testing patterns with `_test.go` suffixes
- Consider integration tests for repository layer with test database
- Test handlers using `httptest` package

## Configuration

Environment-based config with `.env` files:
- `.env` - base configuration
- `.env.local` - local development overrides  
- `.env.staging` - staging overrides

Key environment variables loaded via `internal/config/`.