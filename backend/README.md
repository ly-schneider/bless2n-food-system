# Rentro Backend

A Go-based backend service for the Rentro application, featuring REST APIs, background job processing, and PostgreSQL database integration.

## Architecture

The application follows a clean architecture pattern with the following structure:

- **cmd/**: Application entry points (api, worker)
- **internal/**: Private application code
  - **config/**: Configuration management
  - **db/**: Database connection setup
  - **domain/**: Domain models and entities
  - **handlers/**: HTTP request handlers
  - **logger/**: Logging utilities
  - **queue/**: Background job queue setup
  - **repository/**: Data access layer
  - **service/**: Business logic layer

## Features

- RESTful API with Chi router
- Background job processing with Asynq
- PostgreSQL database with GORM
- Redis for caching and job queues
- Structured logging with Zap
- Docker support for development and production
- Database migrations
- Live-reload for development

## Development Setup

### Prerequisites

- Go 1.24+
- PostgreSQL 16+
- Redis 7+
- Docker & Docker Compose (optional)

### Local Development

1. **Clone and setup**:
   ```bash
   git clone <repository-url>
   cd rentro/backend
   cp .env.example .env
   # Edit .env with your configuration
   ```

2. **Install development tools**:
   ```bash
   make install-tools
   ```

3. **Start services locally**:
   ```bash
   # Option 1: Use Docker for dependencies only
   docker-compose up postgres redis asynqmon -d
   
   # Then run API with live-reload
   make dev
   
   # In another terminal, run worker
   make worker
   ```

4. **Or start everything with Docker**:
   ```bash
   make docker-up
   ```

### Available Commands

```bash
make help           # Show all available commands
make dev            # Start API with live-reload (development only)
make api            # Build and run API (production mode)
make worker         # Build and run worker
make test           # Run tests
make clean          # Clean build artifacts
make install-tools  # Install development tools (Air, migrate)

# Docker commands
make docker-up      # Start all services
make docker-down    # Stop all services
make docker-build   # Build Docker images

# Database migrations
make migrate-up     # Run migrations up
make migrate-down   # Run migrations down
make migrate-create name=migration_name  # Create new migration
```

## Development vs Production

### Development (`make dev`)
- Uses Air for live-reload
- Automatic restart on file changes
- Development logging
- Local environment variables

### Production (`make api` or Docker)
- No live-reload tools
- Optimized builds
- Production logging
- Environment-specific configuration

### Docker Builds
- **Never include Air or live-reload tools**
- Multi-stage builds for smaller images
- Production-ready binaries
- Separate API and Worker containers

## API Endpoints

### Products
- `GET /products` - List all products
- `GET /products/{id}` - Get product by ID
- `POST /products` - Create new product
- `PUT /products/{id}` - Update product
- `DELETE /products/{id}` - Delete product

## Configuration

Environment variables are loaded from:
1. `.env` file (base configuration)
2. `.env.{APP_ENV}` file (environment-specific overrides)
3. System environment variables

Key configuration options:
- `APP_ENV`: Environment (local, development, staging, production)
- `APP_PORT`: API server port
- `POSTGRES_*`: Database configuration
- `REDIS_*`: Redis configuration
- `LOG_LEVEL`: Logging level

## Background Jobs

The application uses Asynq for background job processing:

- **API**: Enqueues jobs (e.g., product creation events)
- **Worker**: Processes jobs (e.g., sending emails, webhooks)
- **Asynqmon**: Web UI for monitoring jobs (http://localhost:8081)

## Database

- **PostgreSQL 16** with GORM ORM
- **Migrations** managed with golang-migrate
- **Connection pooling** and health checks

## Logging

- **Zap** structured logging
- **Development**: Console-friendly output
- **Production**: JSON format for log aggregation

## Testing

```bash
make test  # Run all tests
go test -v ./internal/...  # Run specific package tests
```

## Deployment

### Docker Compose (Staging/Production)

```bash
# Copy and configure environment
cp .env.example .env.production
# Edit .env.production with production values

# Deploy with specific profile
docker-compose --env-file .env.production --profile api up -d
```

### Manual Deployment

```bash
# Build binaries
make api
make worker

# Run migrations
make migrate-up

# Deploy binaries to your infrastructure
```

## Monitoring

- **Health checks**: Built into Docker containers
- **Asynqmon**: Job queue monitoring at http://localhost:8081
- **Logs**: Structured JSON logging for production
- **Metrics**: Ready for Prometheus integration

## Contributing

1. Make sure tests pass: `make test`
2. Use `make dev` for development with live-reload
3. Follow Go conventions and add tests for new features
4. Ensure Docker builds work: `make docker-build`
