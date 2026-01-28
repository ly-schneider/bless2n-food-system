# External Integrations

**Analysis Date:** 2026-01-28

## APIs & External Services

**Payment Processing:**
- Payrexx - Primary payment gateway for Swiss market
  - SDK/Client: Custom internal `backend/internal/payrexx/` package
  - Auth: `PAYREXX_INSTANCE`, `PAYREXX_API_SECRET`, `PAYREXX_WEBHOOK_SECRET`
  - Purpose: TWINT, card payments, invoice generation
  - Implementation: `internal/payrexx/client.go`, `internal/payrexx/gateway.go`, `internal/payrexx/webhook.go`
  - Webhook verification: HMAC-MD5 signature validation

- Stripe API - Legacy payment provider (being phased out)
  - SDK/Client: `github.com/stripe/stripe-go/v82`
  - Auth: `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`
  - Purpose: Payment Intents, webhook handling
  - Note: Code comments indicate migration away from Stripe to Payrexx

**Authentication & Identity:**
- Neon Auth - External authentication service (OIDC)
  - SDK/Client: None (standard OIDC JWT validation)
  - Auth: `NEON_AUTH_URL` (JWKS discovery endpoint), optional `NEON_AUTH_AUDIENCE`
  - Purpose: JWT token validation, federated authentication
  - Implementation: `internal/auth/middleware.go`, `internal/auth/jwks.go`
  - Validation: Token issuer verification, audience claim check (if configured)

- Google OAuth 2.0 - Federated login
  - SDK/Client: Client-side OIDC flow (Authorization Code + PKCE)
  - Auth: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`
  - Purpose: User sign-up and sign-in federation
  - Implementation:
    - Backend: `app/api/auth/google/start/route.ts`, `app/api/auth/google/callback/route.ts`
    - Backend validation: POST to `/v1/auth/google/code` via backend
  - Frontend: Google ID token passed to backend for server-side validation

**Email & Messaging:**
- Plunk (Email Service)
  - SDK/Client: HTTP API client (custom implementation)
  - Auth: `PLUNK_API_KEY`
  - Purpose: Transactional email sending
  - Config: `PLUNK_FROM_NAME`, `PLUNK_FROM_EMAIL`, `PLUNK_REPLY_TO`
  - Integration: `internal/config/config.go` PlunkConfig struct

## Data Storage

**Databases:**
- PostgreSQL 14+ (Primary)
  - Connection: `POSTGRES_DSN` (Neon connection string with SSL)
  - Client: `github.com/jackc/pgx/v5` (driver), `gorm.io/driver/postgres` (ORM)
  - Pool Configuration:
    - `POSTGRES_MAX_CONNS` (default: 25)
    - `POSTGRES_MIN_CONNS` (default: 5)
    - `POSTGRES_MAX_CONN_LIFETIME` (default: 1 hour)
    - `POSTGRES_MAX_CONN_IDLE_TIME` (default: 30 minutes)
  - Migrations: Flyway-based, located in `db/flyway/migrations/`
    - V1: Custom types and enums
    - V2: Main schema (products, orders, users, etc.)
    - V3: Payrexx payment columns
    - V4: Idempotency keys for payment deduplication

- MongoDB (Deprecated - Legacy support only)
  - Connection: `MONGO_URI` (optional, fallback support)
  - Client: `go.mongodb.org/mongo-driver/v2`
  - Purpose: Legacy data, being phased out
  - Note: Configuration present but not actively used in new development

**File Storage:**
- Local filesystem only - No cloud storage integration currently

**Caching:**
- Redis v5.10.0 (Frontend dependency listed, backend usage not detected)
  - Purpose: Likely for session/cache management (configuration pending)

## Authentication & Identity

**Auth Provider:**
- Neon Auth (Primary) - OIDC-compliant external service
  - Implementation: JWKS-based JWT validation
  - Middleware: `internal/auth/NeonAuthMiddleware`
  - Issuer validation: URL-based issuer extraction from `NEON_AUTH_URL`
  - Token claims validated: `iss`, `aud` (optional), `exp`

**Legacy Auth (Deprecated):**
- Custom JWT (Ed25519) with local key management
  - Keys: `JWT_PRIV_PEM`, `JWT_PUB_PEM`
  - Issuer: `JWT_ISSUER` - configured as backend URL
  - Note: Comments indicate this is deprecated in favor of Neon Auth

## Monitoring & Observability

**Error Tracking:**
- OpenTelemetry OTLP (OTEL) - Distributed tracing infrastructure
  - Backend Implementation: `internal/observability/` package
  - Exporters: `otlptrace/otlptracehttp` for HTTP/gRPC traces
  - Frontend SDK: OpenTelemetry SDK for Node.js with Vercel OTEL integration
  - Configuration: Environment-based via `OTEL_*` variables
  - Resource attributes: Deployment environment tagging

**Logs:**
- Structured JSON logging (backend)
  - Framework: Uber Zap (`go.uber.org/zap`)
  - Format: JSON structured logs for machine parsing
  - Level control: `LOG_LEVEL` (default: info)
  - Development mode: `LOG_DEVELOPMENT` flag for pretty-printing

## CI/CD & Deployment

**Hosting:**
- Azure Container Apps (multi-region deployment)
  - Auto-scaling: 0-20 replicas per environment
  - Environments: staging, production with separate VNets
  - Registry: Azure Container Registry (GHCR in GitHub workflows)

**CI Pipeline:**
- GitHub Actions workflow-based
  - Phases: Pre-flight checks, CI testing, Security scanning, Build, Deploy
  - Test sharding: 4 shards (backend), 3 shards (frontend)
  - Coverage threshold: 70% enforced
  - Build output: Docker images pushed to GHCR

**Local Development Services:**
- Docker Compose services:
  - `mongo` - MongoDB 8 (port 27017, deprecated)
  - `mailpit` - Email testing UI (ports 8025 UI, 1025 SMTP)
  - `mongo-express` - MongoDB web UI (port 8081)
  - Optional: backend, web-app containers

## Environment Configuration

**Required env vars:**
- Database: `POSTGRES_DSN`, `MONGO_URI` (optional), `MONGO_DATABASE`
- Auth: `NEON_AUTH_URL`, `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`
- Payments: `PAYREXX_INSTANCE`, `PAYREXX_API_SECRET`, `PAYREXX_WEBHOOK_SECRET`
- Email: `PLUNK_API_KEY`, `PLUNK_FROM_EMAIL`, `PLUNK_FROM_NAME`
- Logging: `LOG_LEVEL`, `LOG_DEVELOPMENT`
- Security: `SECURITY_TRUSTED_ORIGINS`, `SECURITY_ENABLE_HSTS`, `SECURITY_ENABLE_CSP`
- Application: `APP_ENV`, `APP_PORT`, `PUBLIC_BASE_URL`, `STATION_QR_SECRET`

**Secrets location:**
- Development: `.env` and `.env.{APP_ENV}` files (git-ignored)
- Production: Azure Key Vault (referenced via managed identities)
- Docker: Environment variables passed via deployment manifests
- Frontend: `NEXT_PUBLIC_*` variables for client-side access (limited scope)

## Webhooks & Callbacks

**Incoming:**
- Payrexx Payment Webhooks
  - Endpoint: `POST /v1/payments/webhook` (mapped via handler)
  - Verification: HMAC-MD5 signature validation using `PAYREXX_API_SECRET`
  - Payload: Payment status updates, invoice notifications
  - Handler: `internal/handler/payment_pg.go` PaymentWebhookHandler

- Stripe Webhooks (Legacy)
  - Endpoint: Webhook receiver present but deprecated
  - Verification: Signature validation using `STRIPE_WEBHOOK_SECRET`
  - Status: Being replaced by Payrexx webhooks

**Outgoing:**
- Frontend → Backend API
  - Base URL: `NEXT_PUBLIC_API_BASE_URL` (client-side) or `BACKEND_INTERNAL_URL` (server-side)
  - Pattern: Fetch-based with JWT bearer tokens
  - Utility: `lib/http.ts` error message extraction
- Backend → Neon Auth
  - JWKS endpoint discovery for token validation (automatic caching)

## API Documentation

**Swagger/OpenAPI:**
- Auto-generated from Go code annotations
- Command: `swag init -g cmd/backend/main.go`
- Endpoint: Available at `/swagger/index.html` (HTTP layer)
- Tool: `github.com/swaggo/swag` and `swaggo/http-swagger`

---

*Integration audit: 2026-01-28*
