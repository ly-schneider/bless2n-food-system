# Technology Stack

**Analysis Date:** 2026-01-28

## Languages

**Primary:**
- Go 1.25 - Backend API and core business logic
- TypeScript 5.8 - Web app, frontend application logic
- SQL (PostgreSQL dialect) - Database schemas and migrations

**Secondary:**
- JavaScript - Build tooling and configuration
- YAML - CI/CD pipeline definitions
- HCL - Infrastructure as Code (Terraform)

## Runtime

**Environment:**
- Go 1.25.x for backend
- Node.js 20.0.0+ for frontend (pnpm package manager)
- PostgreSQL 14+ (production via Neon)
- Docker containerization for local and production deployments

**Package Manager:**
- Go modules (`go.mod`/`go.sum`)
- pnpm 10.19.0 for frontend dependencies
- Lockfile: `go.sum` and `pnpm-lock.yaml` present

## Frameworks

**Core:**
- Next.js 16.1.1 - React 19 SSR/SSG framework with App Router
- Chi v5.2.3 - Lightweight HTTP router for Go backend
- GORM v1.25.12 - PostgreSQL ORM for data persistence

**Testing:**
- Vitest 4.0.16 - Unit testing framework (frontend)
- Playwright 1.57.0 - E2E testing (frontend)
- Storybook 9.1.17 - Component testing and documentation
- Go standard testing + stretchr/testify for assertions (backend)

**Build/Dev:**
- Air v1.63.0 - Hot-reload development server for Go
- Next.js built-in dev server for frontend
- Swagger/Swag - API documentation generation
- golangci-lint v2.5.0 - Go linting
- Tailwind CSS 4.1.18 - Utility-first CSS framework

## Key Dependencies

**Critical:**
- `github.com/jackc/pgx/v5` v5.8.0 - PostgreSQL driver for connection pooling
- `gorm.io/driver/postgres` v1.5.11 - GORM PostgreSQL dialect
- `github.com/stripe/stripe-go/v82` v82.5.1 - Stripe API client (legacy, being replaced)
- `react-hook-form` v7.70.0 - Form state management (frontend)
- `zod` v4.3.5 - TypeScript-first schema validation (frontend)

**Infrastructure:**
- `go.uber.org/fx` v1.24.0 - Dependency injection container (backend)
- `go.uber.org/zap` v1.27.1 - Structured JSON logging (backend)
- `github.com/golang-jwt/jwt/v5` v5.3.0 - JWT token handling
- `golang.org/x/crypto` v0.46.0 - Cryptographic primitives (Argon2 password hashing)

**Observability:**
- `go.opentelemetry.io/otel/*` v1.26-1.37.0 - Distributed tracing (backend)
- `go.opentelemetry.io/otel/exporters/otlp/otlptracehttp` - OTLP exporter for traces
- OpenTelemetry SDK for Node.js (frontend, versions 0.209-2.3.0)

**UI Components:**
- Radix UI (v1.x) - Headless component library
- Lucide React - Icon library
- Shadcn/ui patterns - Pre-built accessible components

**Validation:**
- `github.com/go-playground/validator/v10` v10.30.1 - Struct field validation (backend)
- Zod - Schema validation and type inference (frontend)

## Configuration

**Environment:**
- Dotenv files (`.env`, `.env.{APP_ENV}`) for local configuration
- Environment variable fallbacks for Docker deployments
- Docker detection via `/.dockerenv` file presence
- Config structure in `internal/config/config.go` with typed fields

**Build:**
- `next.config.js` - Next.js build configuration
- `tsconfig.json` - TypeScript compiler options
- `Dockerfile` - Multi-stage build using golang:1.25-alpine and distroless base
- `Makefile` - Backend build, test, and dev commands

## Platform Requirements

**Development:**
- Go 1.25+
- Node.js 20.0.0+
- pnpm 10.19.0
- Docker & Docker Compose for local services
- PostgreSQL (via Docker Compose)
- MongoDB (deprecated, present for backward compatibility)

**Production:**
- Azure Container Apps (deployment target)
- Azure Cosmos DB (MongoDB API) - deprecated
- Azure PostgreSQL via Neon (current database)
- Azure Container Registry for image storage
- Azure Key Vault for secrets management
- Log Analytics workspace for observability

---

*Stack analysis: 2026-01-28*
