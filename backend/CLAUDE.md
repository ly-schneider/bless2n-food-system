# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture Overview

This is a Go-based HTTP backend service for the Bless2n Food System using clean architecture principles:

- **Dependency Injection**: Uses Uber FX for dependency injection and lifecycle management
- **Database**: MongoDB with custom repository pattern
- **HTTP Router**: Chi router with middleware support
- **Configuration**: Environment-based configuration with Docker and local development support
- **Domain Layer**: Rich domain models in `internal/domain/` covering food system entities (users, orders, products, devices, etc.)

### Directory Structure

- `cmd/backend/main.go` - Application entry point with FX dependency injection setup
- `internal/app/` - Application wiring and bootstrapping (config, database, handlers, repositories, services)
- `internal/config/` - Configuration management with environment variable loading
- `internal/domain/` - Business entities and domain models
- `internal/repository/` - Data access layer with MongoDB implementations
- `internal/service/` - Business logic layer (auth, email services)
- `internal/handler/` - HTTP request handlers
- `internal/database/` - Database connection and setup

## Development Commands

### Local Development
```bash
# Start MongoDB and supporting services
make docker-up

# Start API server with live-reload (requires services running)
make dev
```

### Docker Development
```bash
# Start all services including backend in Docker
make docker-up-backend

# Restart services (no backend)
make docker-restart

# Restart all services including backend
make docker-restart-backend
```

### Database Management
```bash
# Open MongoDB shell
make mongo-shell

# Stop services but keep data
make docker-down

# Stop services and remove all data (DESTRUCTIVE)
make docker-down-v
```

### Testing
```bash
# Run unit tests
make test

# Run end-to-end tests with test infrastructure
make test-e2e

# Run all tests with coverage (80% threshold)
make test-coverage

# Setup/teardown test infrastructure manually
make test-setup
make test-teardown
```

### Build and Development
```bash
# Manual build
go build -o ./tmp/main ./cmd/backend

# View logs
make logs

# Show running services
make ps

# Clean build artifacts
make clean
```

## Environment Configuration

- Copy `.env.example` to `.env` and configure values
- The app automatically detects Docker vs local environment
- Local development uses `*_LOCAL` variables, Docker uses `*_DOCKER` variables
- MongoDB admin credentials, SMTP settings, JWT secret, and app configuration required

## Key Technologies

- **Web Framework**: Chi router (github.com/go-chi/chi/v5)
- **Database**: MongoDB with official Go driver
- **Validation**: go-playground/validator
- **Logging**: Uber Zap structured logging
- **Dependency Injection**: Uber FX
- **Live Reload**: Air (configured in .air.toml)
- **Email Testing**: Mailpit for local SMTP testing

## Testing Architecture

Comprehensive end-to-end testing architecture with:

- **Test Isolation**: Dedicated test database and services (MongoDB on port 27018, SMTP on 1026)
- **Coverage Enforcement**: 80% minimum coverage threshold with automated verification
- **Docker-based Infrastructure**: Consistent testing environment with `test/docker-compose.test.yml`
- **Complete Auth Testing**: Registration, login, OTP verification, token refresh, logout flows
- **CI/CD Integration**: GitHub Actions workflow with coverage reporting
- **Test Structure**: Located in `test/` directory with helpers, fixtures, and e2e tests

### Test Commands
- `make test`: Unit tests only
- `make test-e2e`: End-to-end tests with infrastructure
- `make test-coverage`: All tests with 80% coverage verification

See `test/README.md` for detailed testing documentation.

## Application Startup Flow

1. `main.go` initializes FX container with providers for config, logger, database, and router
2. Repositories are registered for all domain entities
3. Services (auth, email) are wired up
4. HTTP handlers are registered
5. HTTP server starts on configured port with graceful shutdown

The application follows a clean separation between domain models, repositories, services, and HTTP handlers, making it easy to extend with new features while maintaining testability.