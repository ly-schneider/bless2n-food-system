# Bless2n Food System Backend

A comprehensive Go-based HTTP backend service for the Bless2n Food System

## 📋 Table of Contents

- [Features](#-features)
- [Architecture Overview](#-architecture-overview)
- [Prerequisites](#-prerequisites)
- [Development Setup](#-development-setup)
- [Available Commands](#-available-commands)
- [Project Structure](#-project-structure)
- [Testing](#-testing)

## ✨ Features

### Core Functionality
- **RESTful API**: Built with Chi router for high-performance HTTP routing
- **Clean Architecture**: Separation of concerns with domain, repository, service, and handler layers
- **Authentication & Authorization**: JWT-based authentication with role-based access control
- **User Management**: Support for admin and customer roles with verification system
- **Order Management**: Complete order processing with status tracking
- **Product Catalog**: Product and category management with bundle support
- **Device Management**: IoT device integration for food stations
- **Inventory Tracking**: Real-time inventory ledger system
- **Audit Logging**: Comprehensive audit trail for all operations

### Technical Features
- **Dependency Injection**: Uber FX for clean dependency management
- **Database**: MongoDB with custom repository pattern
- **Email Service**: SMTP integration with development testing support
- **Configuration Management**: Environment-based configuration
- **Live Reload**: Air integration for development productivity
- **Docker Support**: Complete containerization with Docker Compose
- **Logging**: Structured logging with Uber Zap
- **Validation**: Request validation with go-playground/validator
- **Security**: Argon2 password hashing and JWT tokens

## 🏗 Architecture Overview

This backend follows **Clean Architecture** principles with clear separation of concerns:

```
┌─────────────────────────────────────────┐
│              HTTP Layer                 │
│  (Chi Router, Middleware, Handlers)     │
├─────────────────────────────────────────┤
│             Service Layer               │
│    (Business Logic, Auth, Email)        │
├─────────────────────────────────────────┤
│           Repository Layer              │
│        (Data Access, MongoDB)           │
├─────────────────────────────────────────┤
│             Domain Layer                │
│      (Entities, Business Rules)         │
└─────────────────────────────────────────┘
```

### Key Architectural Decisions

- **Dependency Injection**: Uber FX manages all dependencies and lifecycle
- **Repository Pattern**: Clean abstraction over MongoDB operations
- **Domain-Driven Design**: Rich domain models with business logic
- **Configuration**: Environment-based with Docker and local development support
- **Error Handling**: Centralized error handling with proper HTTP status codes

## 📋 Prerequisites

### Required
- **Go 1.24.3+**: Latest Go version for optimal performance
- **Docker & Docker Compose**: For running services (MongoDB, Mailpit, etc.)
- **Make**: For running development commands

### Optional
- **Air**: For live reload during development (installed automatically)
- **MongoDB Compass**: For database visualization
- **Postman/Insomnia**: For API testing

## 🔧 Development Setup

### Local Development (Recommended)

This approach runs MongoDB in Docker while running the backend locally with live reload:

```bash
# 1. Environment setup
cp .env.example .env
# Configure your .env file (see Configuration section)

# 2. Generate JWT keys (REQUIRED)
mkdir -p secrets/dev
openssl genpkey -algorithm Ed25519 -out secrets/dev/ed25519.pem
openssl pkey -in secrets/dev/ed25519.pem -pubout -out secrets/dev/ed25519.pub.pem

# 3. Start supporting services (MongoDB, Mailpit, etc.)
make docker-up

# 4a. Run backend with live reload
make dev
```

**Advantages:**
- Fast rebuilds with Air
- Easy debugging and profiling
- Direct access to Go tooling
- Better IDE integration

### Full Docker Development

Run everything in Docker containers:

```bash
# 1. Setup environment and JWT keys (same as above)
cp .env.example .env
mkdir -p secrets/dev
openssl genpkey -algorithm Ed25519 -out secrets/dev/ed25519.pem
openssl pkey -in secrets/dev/ed25519.pem -pubout -out secrets/dev/ed25519.pub.pem

# 2. Start all services including backend
make docker-up-backend
```

**Use cases:**
- Production-like testing
- Consistent environment across team
- CI/CD pipeline testing

### Service URLs

When services are running:

- **Backend API**: http://localhost:8080
- **Mongo Express**: http://localhost:8081 (Database UI)
- **Mailpit**: http://localhost:8025 (Email testing UI)
- **MongoDB**: localhost:27017 (Direct connection)

## 📝 Available Commands

Run `make` or `make help` to see all available commands with descriptions.

### Key Commands
```bash
make dev                    # Start with live-reload
make docker-up              # Start supporting services
make docker-up-backend      # Start all services (including backend)
make logs                   # View service logs
make mongo-shell           # Access MongoDB shell
make clean                  # Clean build artifacts
```

## 📁 Project Structure

```
backend/
├── cmd/backend/           # Application entry point
├── internal/
│   ├── app/              # Application wiring & DI setup
│   ├── config/           # Configuration management
│   ├── database/         # Database connection
│   ├── domain/           # Business entities
│   ├── handler/          # HTTP handlers
│   ├── middleware/       # HTTP middleware
│   ├── repository/       # Data access layer
│   ├── service/          # Business logic
│   ├── utils/            # Utility functions
│   └── http/             # HTTP routing
├── secrets/dev/          # Development JWT keys
├── .air.toml             # Live reload config
├── .env.example          # Environment template
├── docker-compose.yml    # Docker services
├── Dockerfile            # Container definition
├── go.mod & go.sum       # Go dependencies
├── Makefile              # Development commands
└── README.md
```

## 🧪 Testing

### Test Structure (Planned)

```
backend/
└── test/
    ├── integration/           # Integration tests
    ├── e2e/                   # e2e tests
    └── fixtures/              # Test data and fixtures
```
