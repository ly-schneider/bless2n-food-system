# Architecture

**Analysis Date:** 2026-01-28

## Pattern Overview

**Overall:** Clean Architecture with Progressive PostgreSQL Migration

This is a monorepo containing multiple interconnected applications following Clean Architecture principles in the backend and Next.js App Router in the frontend. The system is undergoing a major database migration from MongoDB/Stripe to PostgreSQL/Payrexx. Customer-facing functionality (orders, payments, POS, stations) has been fully migrated; admin functionality remains on MongoDB.

**Key Characteristics:**
- Clear separation of concerns: Domain → Repository → Service → Handler → HTTP Router
- Dependency Injection via Uber FX for loose coupling and testability
- Dual database support during migration: MongoDB (legacy) and PostgreSQL (new)
- Dual payment providers: Stripe (legacy) and Payrexx (new)
- NeonAuth external authentication (replaces local auth services)
- Multi-tenant capability: Public shop, Admin dashboard, POS terminals, Station devices
- API-first backend (Chi router) with Next.js frontend consuming REST APIs

## Layers

**Domain Layer (Internal Models):**
- Purpose: Business entities and rules
- Location: `bfs-backend/internal/domain/` (MongoDB domain - legacy), `bfs-backend/internal/model/` (PostgreSQL domain - new)
- Contains: Entity definitions, validation, constants, DTOs
- Depends on: Nothing (pure business logic)
- Used by: Services, Handlers, Repositories

Example entities: `order.go`, `product.go`, `user.go`, `category.go`, `pos_device.go`, `station.go`

**Repository Layer (Data Access):**
- Purpose: Abstract database operations from business logic
- Location:
  - MongoDB: `bfs-backend/internal/repository/` (legacy)
  - PostgreSQL: `bfs-backend/internal/postgres/` (new)
- Contains: Database queries, transactions, persistence operations
- Depends on: Domain models
- Used by: Services

Example repositories:
- `internal/postgres/order_repo.go` - Order CRUD, customer listings, admin queries
- `internal/postgres/product_repo.go` - Product listings, menu slots, jeton management
- `internal/postgres/inventory_ledger_repo.go` - Stock tracking and adjustments

**Service Layer (Business Logic):**
- Purpose: Orchestrate business operations, validation, payment processing
- Location: `bfs-backend/internal/service/`
- Contains: Business logic, validation, external integrations (Payrexx, email)
- Depends on: Repositories, Domain models
- Used by: Handlers

PostgreSQL services (new): `category_pg.go`, `order_pg.go`, `payment_pg.go`, `pos_pg.go`, `pos_config_pg.go`, `product_pg.go`, `station_pg.go`

Legacy MongoDB services: `payment.go`, `pos.go`, `order.go`, `station.go`, `jwt.go`, `email.go`, `oidc.go`

**Handler Layer (HTTP Handlers):**
- Purpose: Handle HTTP requests, parse input, call services, return responses
- Location: `bfs-backend/internal/handler/`
- Contains: HTTP endpoint implementations, request/response mapping
- Depends on: Services, Domain models
- Used by: HTTP Router

Example handlers:
- `order.go` / `order_pg.go` - Order listing (authenticated), public order details (no auth)
- `pos.go` / `pos_pg.go` - POS device management, order creation, payment methods
- `payment.go` / `payment_pg.go` - Payment intent creation, webhook handling
- `admin_*.go` - Admin dashboard operations (still using MongoDB)

**HTTP Router Layer:**
- Purpose: Route HTTP requests to handlers, apply middleware
- Location: `bfs-backend/internal/http/router.go`
- Contains: Route definitions, middleware stack, CORS, authentication, CSRF
- Depends on: All handlers, middleware, config
- Used by: HTTP server startup

Routes organized by:
- Public: `/v1/products`, `/v1/orders/{id}`, `/health`, `/.well-known/jwks.json`
- Public device-gated: `/v1/stations/*`, `/v1/pos/*` (via X-Device-Token or X-Pos-Token)
- Authenticated: `/v1/orders` (user's own orders)
- Admin: `/v1/admin/*` (requires JWT bearer token + admin role)
- Payments: `/v1/payments/*` (optional auth, Payrexx integration)

**Middleware Layer:**
- Purpose: Cross-cutting concerns (auth, security, observability)
- Location: `bfs-backend/internal/middleware/` and `bfs-backend/internal/auth/`
- Contains: JWT validation, CSRF protection, security headers, request tracing
- Key middleware:
  - `JWTMiddleware.RequireAuth` - Validates JWT bearer tokens
  - `JWTMiddleware.OptionalAuth` - Allows logged-in users but doesn't require it
  - `NeonAuthMiddleware` - Fetches user claims from external NeonAuth provider
  - `SecurityMiddleware.SecurityHeaders` - HSTS, X-Content-Type-Options, etc.
  - `SecurityMiddleware.CORS` - Configurable CORS origins
  - `observability.ChiMiddleware` - Request tracing and spans

**Configuration & Wiring (Dependency Injection):**
- Purpose: Bootstrap application, wire dependencies
- Location: `bfs-backend/internal/app/`
- Key files:
  - `server.go` - HTTP server lifecycle (OnStart, OnStop)
  - `config.go` - Load environment configuration
  - `logger.go` - Initialize Uber Zap structured logging
  - `database.go` - MongoDB connection (legacy)
  - `postgres_repositories.go` - GORM PostgreSQL connection, all PostgreSQL repositories
  - `repositories.go` - MongoDB repositories (legacy)
  - `services.go` - Service providers (legacy MongoDB services)
  - `services_pg.go` - Service providers for PostgreSQL services
  - `handlers.go` - Handler providers
  - `router.go` - HTTP router setup with all routes and middleware
  - `observability.go` - Tracing and metrics setup

**Entry Point:**
- Location: `bfs-backend/cmd/backend/main.go`
- Uses Uber FX for lifecycle management and dependency resolution
- Provides all dependencies in order: Config → Logger → Databases → Repositories → Services → Handlers → Router → Server
- Invokes: Observability setup, JWKS client startup, HTTP server startup

## Data Flow

**Customer Ordering Flow (Web):**

1. **Menu Load**: Client requests `GET /v1/products` → ProductHandler → ProductService → ProductRepository → PostgreSQL
2. **Order Creation**: Client posts `POST /v1/payments/create-intent` → PaymentHandler → Payrexx API → creates gateway
3. **Order Submission**: After payment succeeds, client posts `POST /v1/orders` or order created via webhook
4. **Order Tracking**: Authenticated user fetches `GET /v1/orders` → OrderHandler → OrderService → OrderRepository
5. **Pickup**: Customer scans QR or station verifies token → `POST /v1/stations/verify-qr` → StationHandler → validation against order redemption status

**POS Terminal Flow:**

1. **Device Request**: POS device registers `POST /v1/pos/requests` (one-time) → POSHandler creates pending device
2. **Admin Approval**: Admin approves device → `POST /v1/admin/pos/requests/{id}/approve` → device gets X-Pos-Token
3. **Order Creation**: POS creates order `POST /v1/pos/orders` (X-Pos-Token header) → POSHandler → OrderService → PostgreSQL
4. **Payment**: POS sends `POST /v1/pos/orders/{id}/pay-cash|card|twint` → PaymentHandler → Payrexx gateway
5. **Redemption**: Station staff scans/redeems item → `POST /v1/stations/redeem` → StationHandler → OrderLineRedemptionRepository

**Admin Operations Flow:**

1. **Authentication**: Admin token in `Authorization: Bearer <JWT>` header
2. **CSRF**: All state-changing requests (POST/PATCH/DELETE) require CSRF token
3. **Role Check**: Middleware validates admin role from JWT claims or falls back to database check
4. **Operation**: Admin handler executes operation using MongoDB repositories (still legacy)
5. **Audit**: Critical operations logged to `AuditRepository` (MongoDB)

**Payment Processing (Payrexx) - NEW:**

1. Client initiates payment via Payment Element (`/v1/payments/create-intent`)
2. PaymentHandler calls PaymentService → PayrexxClient.CreateGateway()
3. Payrexx processes payment asynchronously (card, TWINT, etc.)
4. Payrexx sends webhook to `POST /v1/payments/webhook`
5. PaymentHandler verifies webhook signature, updates order status
6. Order transitions: `pending` → `paid`

## Key Abstractions

**Repository Pattern:**
- Purpose: Abstract database implementation details
- All repositories implement interfaces defined in service layer
- Example: `OrderRepository` interface has `Create`, `GetByID`, `ListByCustomerID`, `ListAdmin`, `UpdateStatus`
- Switching databases: Just inject different repository implementation

**Service Pattern:**
- Purpose: Encapsulate business logic independent of transport/persistence
- Services depend on repositories and are called by handlers
- Example: `OrderService.PrepareOrder()` validates products, calculates totals, builds order+items

**Handler Pattern:**
- Purpose: Convert HTTP requests to domain operations
- Parse URL params, query strings, request bodies
- Extract user context from middleware
- Call services and return responses
- Example handlers:
  - `OrderHandler.ListMyOrders()` - Parse pagination, get user ID from JWT, call service
  - `AdminProductHandler.PatchPrice()` - Validate price, update inventory, call audit

**Middleware Pattern:**
- Purpose: Apply cross-cutting logic without polluting handlers
- Examples:
  - `JWTMiddleware.RequireAuth` - If no valid JWT, return 401
  - `SecurityMiddleware.CORS` - Add Access-Control-* headers
  - `observability.ChiMiddleware` - Create request span

**Dependency Injection via Uber FX:**
- Purpose: Manage object graph and lifecycle
- Each component provided as constructor function returning interface
- FX resolves dependency graph automatically
- Lifecycle hooks: `OnStart` (setup), `OnStop` (cleanup)
- Example: `NewOrderService` depends on `ProductRepository`, FX auto-injects it

**UUID vs ObjectID Split:**
- PostgreSQL layer uses `uuid.UUID` (github.com/google/uuid)
- MongoDB layer uses `bson.ObjectID`
- Repositories return different types; services/handlers stay generic by using interfaces

## Entry Points

**HTTP Server:**
- Location: `bfs-backend/cmd/backend/main.go`
- Triggers: Application startup via Docker or local `make dev`
- Responsibilities: Parse config, initialize FX app, start HTTP server on configured port

**Health Check Endpoint:**
- Location: `GET /health` or `/ping`
- Triggers: Container orchestration, uptime monitoring
- Responsibilities: Return 200 OK to confirm service is alive

**Swagger Documentation:**
- Location: `GET /swagger/*`
- Triggers: Developers browsing API documentation
- Auto-generated from handler godoc comments

**JWKS Endpoint:**
- Location: `GET /.well-known/jwks.json`
- Triggers: External services verifying JWT tokens
- Responsibilities: Return public keys for token verification

**Stripe/Payrexx Webhooks:**
- Location: `POST /v1/payments/webhook`
- Triggers: Payment gateway (Stripe or Payrexx) after transaction
- Responsibilities: Verify signature, update order status, trigger fulfillment

## Error Handling

**Strategy:** Layered error responses using `response.ProblemDetails` (RFC 7807)

**Patterns:**
- Domain validation errors → 400 Bad Request with `ProblemDetails{Type, Title, Detail}`
- Authentication missing → 401 Unauthorized with `WWW-Authenticate` header
- Permission denied → 403 Forbidden with reason
- Not found → 404 with entity type
- Conflict → 409 with details (e.g., duplicate device request)
- Server errors → 500 with request ID for tracing
- Structured logging with Zap: logger.Error("operation failed", zap.Error(err), zap.String("order_id", id))

**Response types:**
- Success: `response.WriteJSON(w, 200, data)`
- Error: `response.WriteError(w, statusCode, errorMessage)`
- List: `domain.ListResponse[T]{Items: []T, Count: int}`

## Cross-Cutting Concerns

**Logging:**
- Framework: Uber Zap with structured fields
- Entry: `internal/app/logger.go` initializes global logger
- Usage: Handlers and services call `zap.L().Info/Error/Warn` with fields
- Format: JSON in production, colored text in development

**Validation:**
- Framework: `validator` Go package (struct tags like `validate:"required,gte=0"`)
- Request DTOs validated by handlers before calling services
- Domain models self-validate via `Validate()` methods
- Custom validation logic in services (e.g., check product availability)

**Authentication:**
- External: NeonAuth (Neon PostgreSQL built-in auth provider)
- Middleware: `NeonAuthMiddleware` fetches user claims from NeonAuth headers
- JWT: Issued by NeonAuth, verified locally using JWKS from `GET /.well-known/jwks.json`
- Roles: Admin, Customer (extracted from JWT claims)
- Device auth: X-Device-Token (station), X-Pos-Token (POS terminal)

**Transactions & Idempotency:**
- GORM transactions for multi-step operations (e.g., create order + adjust inventory)
- Idempotency keys stored in `app.idempotency` table to prevent duplicate payment intents
- Payment service checks idempotency before creating Payrexx gateway

**Observability:**
- Tracing: OpenTelemetry via `observability.ChiMiddleware`
- Request ID: Automatically generated by Chi middleware, added to logs
- Spans: Created per request with operation name, status, duration
- Metrics: Compatible with OpenTelemetry collectors (Vercel OTEL, Azure Monitor)

---

*Architecture analysis: 2026-01-28*
