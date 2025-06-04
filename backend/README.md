# Rentro Backend

A Go-based backend service for the Rentro application, featuring REST APIs, background job processing, and PostgreSQL database integration.

## Features

- RESTful API with Chi router
- Background job processing with Asynq
- PostgreSQL database with GORM
- Redis for caching and job queues
- Structured logging with Zap
- Docker support for development and production
- Database migrations with Flyway
- Live-reload for development

![ory](https://www.ory.sh/docs/assets/images/1-42e65393379b7f7ddc3f9a05474f27ac.png)

## Development Setup

### Prerequisites

- Go 1.24+
- PostgreSQL 16+
- Redis 7+
- Docker & Docker Compose (optional)
- Java 11+ (for Flyway)

### Local Development

1. **Clone and setup**:
   ```bash
   git clone <repository-url>
   cd rentro/backend
   cp .env.example .env
   # Edit .env with your configuration
   ```

2. **Start services locally**:
   ```bash
   # Run docker
   make docker-up
   
   # Then run API with live-reload
   make dev
   ```

### Available Commands

```bash
make help           # Show all available commands
make dev            # Start API with live-reload (development only)
make clean          # Clean build artifacts

# Database migrations
make flyway-migrate # Run pending migrations
make flyway-info    # Show migration status
make flyway-validate# Validate migration files

# Docker commands
make docker-up      # Start all services
make docker-down    # Stop all services
make docker-build   # Build Docker images
```

## Development vs Production

### Development (`make dev`)
- Uses Air for live-reload
- Automatic restart on file changes
- Development logging
- Local environment variables

### Docker Builds
- **Never include Air or live-reload tools**
- Multi-stage builds for smaller images
- Production-ready binaries
- Separate API and Worker containers

## Background Jobs

The application uses Asynq for background job processing:

- **API**: Enqueues jobs (e.g., product creation events)
- **Worker**: Processes jobs (e.g., sending emails, webhooks)
- **Asynqmon**: Web UI for monitoring jobs (http://localhost:8081)

## Database

- **PostgreSQL 16** with GORM ORM
- **Migrations** managed with Flyway
- **Connection pooling** and health checks

### Running Migrations

You can run migrations independently without restarting containers. This implies that the docker-compose is running:

```bash
# Run all pending migrations
make flyway-migrate

# Check migration status
make flyway-info

# Validate migration files
make flyway-validate
```

## Migrations

Flyway manages database schema changes with versioned SQL scripts located in `db/migration/`:

- `V001__create_uuid_extension.sql`
- `V002__create_roles.sql`
- etc.

### Migration Naming Convention
- Version: `V{version}__{description}.sql`
- Undo: `U{version}__{description}.sql` (optional)
- Repeatable: `R__{description}.sql`

## Logging

- **Zap** structured logging
- **Development**: Console-friendly output
- **Production**: JSON format for log aggregation