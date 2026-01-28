# Codebase Structure

**Analysis Date:** 2026-01-28

## Directory Layout

```
bless2n-food-system/
├── bfs-backend/                    # Go HTTP backend (Clean Architecture)
│   ├── cmd/backend/main.go         # Entry point
│   ├── internal/
│   │   ├── domain/                 # MongoDB domain models (legacy)
│   │   ├── model/                  # PostgreSQL domain models (new)
│   │   ├── repository/             # MongoDB repositories (legacy)
│   │   ├── postgres/               # PostgreSQL repositories (new)
│   │   ├── service/                # Business logic services
│   │   ├── handler/                # HTTP endpoint handlers
│   │   ├── http/                   # Router and route definitions
│   │   ├── middleware/             # HTTP middleware (JWT, CSRF, security)
│   │   ├── auth/                   # Authentication (NeonAuth)
│   │   ├── app/                    # Dependency injection & wiring
│   │   ├── config/                 # Configuration loading
│   │   ├── response/               # Response formatting helpers
│   │   ├── database/               # Database initialization
│   │   ├── payrexx/                # Payrexx payment gateway client
│   │   ├── observability/          # OpenTelemetry tracing
│   │   ├── utils/                  # Utility functions
│   │   └── seed/                   # Development seed data (MongoDB)
│   ├── db/
│   │   ├── flyway/migrations/      # SQL migrations (versioned)
│   │   └── provisioning/           # Database setup scripts
│   ├── docs/                       # Swagger API documentation
│   ├── test/                       # Test files
│   ├── go.mod & go.sum             # Go dependencies
│   ├── Makefile                    # Build commands
│   └── docker-compose.yml          # Local development services
│
├── bfs-web-app/                    # Next.js 15 web application
│   ├── app/                        # App Router (file-based routing)
│   │   ├── (site)/                 # Public ordering interface
│   │   │   ├── layout.tsx
│   │   │   ├── page.tsx
│   │   │   ├── food/checkout/      # Checkout flow (product selection, payment)
│   │   │   └── profile/page.tsx
│   │   ├── admin/                  # Admin dashboard
│   │   │   ├── layout.tsx
│   │   │   ├── page.tsx
│   │   │   ├── orders/             # Order management
│   │   │   ├── products/           # Product management
│   │   │   ├── categories/         # Category management
│   │   │   ├── stations/           # Station management
│   │   │   ├── pos/                # POS device management
│   │   │   ├── menu/               # Menu editor
│   │   │   ├── invites/            # Admin invite management
│   │   │   ├── users/              # User management
│   │   │   ├── sessions/           # Session management
│   │   │   └── jetons/             # Jeton/voucher management
│   │   ├── pos/                    # POS terminal interface
│   │   │   ├── layout.tsx
│   │   │   └── page.tsx
│   │   ├── station/                # Station device interface
│   │   │   ├── layout.tsx
│   │   │   └── page.tsx
│   │   ├── (auth)/                 # Auth routes
│   │   │   ├── login/page.tsx
│   │   │   ├── signup/page.tsx
│   │   │   └── forgot-password/page.tsx
│   │   ├── api/                    # Server actions & API routes
│   │   └── health/route.ts         # Health check endpoint
│   ├── components/                 # Reusable React components
│   │   ├── ui/                     # Radix UI + Tailwind components
│   │   ├── layout/                 # Header, Footer, Navigation
│   │   ├── admin/                  # Admin-specific components
│   │   ├── pos/                    # POS terminal components
│   │   ├── cart/                   # Shopping cart UI
│   │   ├── menu/                   # Menu display components
│   │   ├── payment/                # Payment form components (Stripe)
│   │   └── auth/                   # Auth UI components
│   ├── lib/                        # Utility functions & helpers
│   │   ├── api.ts                  # API client wrapper
│   │   ├── api/                    # API endpoints
│   │   ├── auth/                   # Auth utilities
│   │   ├── server/                 # Server-side utilities
│   │   ├── stripe.ts               # Stripe SDK setup
│   │   ├── csrf.ts                 # CSRF token handling
│   │   └── utils.ts                # General utilities
│   ├── hooks/                      # React custom hooks
│   ├── contexts/                   # React Context providers
│   │   └── cart-context.tsx        # Shopping cart state
│   ├── types/                      # TypeScript types & interfaces
│   ├── styles/                     # Global CSS & Tailwind config
│   ├── public/                     # Static assets (images, fonts)
│   ├── .storybook/                 # Storybook component showcase config
│   ├── package.json                # npm dependencies & scripts
│   ├── tsconfig.json               # TypeScript configuration
│   └── next.config.js              # Next.js configuration
│
├── bfs-android-app/                # Android mobile application
│   └── app/src/                    # Source code
│
├── bfs-cloud/                      # Azure Terraform infrastructure
│   ├── modules/                    # Reusable Terraform modules
│   │   ├── containerapp/           # Container Apps (web/backend)
│   │   ├── cosmos_mongo/           # MongoDB database
│   │   ├── stack/                  # Full environment stack
│   │   └── ...
│   ├── envs/
│   │   ├── staging/                # Staging environment config
│   │   └── production/             # Production environment config
│   └── defaults/                   # Default variable values
│
├── bfs-docs/                       # Fumadocs documentation site
│   ├── app/                        # Documentation pages
│   ├── content/                    # Markdown documentation
│   └── lib/                        # Doc utilities
│
├── bfs-http/                       # Yaak HTTP request collections
│   └── *.http                      # Manual API testing requests
│
├── docker-compose.yml              # Local development: MongoDB, Mailpit, Mongo Express
├── CLAUDE.md                       # Project guidelines for Claude
└── MIGRATION-PROGRESS.md           # PostgreSQL migration status tracking
```

## Directory Purposes

**bfs-backend/cmd/backend/**
- Purpose: Application entry point
- Contains: `main.go` with FX app bootstrap
- Key files: `main.go` (startup, healthcheck flag handling)

**bfs-backend/internal/domain/**
- Purpose: Business entities and rules (MongoDB era)
- Contains: Order, Product, User, Category, Station models with BSON tags
- Key files: `order.go`, `product.go`, `user.go`, `category.go`, `pos_device.go`
- Status: Being replaced by `internal/model/` for PostgreSQL migration

**bfs-backend/internal/model/**
- Purpose: PostgreSQL domain models with GORM
- Contains: Entity structs with GORM tags, UUIDs instead of ObjectIDs
- Key files: `order.go`, `product.go`, `device.go`, `idempotency.go`, `enums.go`
- Status: New PostgreSQL-native models

**bfs-backend/internal/repository/**
- Purpose: MongoDB data access layer (legacy)
- Contains: CRUD operations, queries, transactions against MongoDB
- Key files: `order.go`, `product.go`, `category.go`, `user.go`, `refresh_token.go`
- Status: Being superseded by `internal/postgres/`

**bfs-backend/internal/postgres/**
- Purpose: PostgreSQL data access layer (new)
- Contains: GORM-based repositories, SQL operations, transactions
- Key files: `order_repo.go`, `product_repo.go`, `device_repo.go`, `order_line_repo.go`
- Status: Active - customer-facing code fully migrated here

**bfs-backend/internal/service/**
- Purpose: Business logic orchestration
- Contains: Order preparation, payment processing, category management, email sending
- Key files:
  - `order_pg.go` - Order validation, status transitions
  - `payment_pg.go` - Payrexx payment gateway integration
  - `product_pg.go` - Product queries with menu slots
  - `station_pg.go` - Station verification and device management
  - `pos_pg.go` - POS device and order processing
  - `jwt.go` - Token generation for admin invites
  - `email.go` - Email template rendering and sending (SMTP)
  - Legacy: `payment.go` (Stripe), `order.go` (MongoDB), etc.
- DI Wiring: `services_pg.go` for FX providers

**bfs-backend/internal/handler/**
- Purpose: HTTP request handling
- Contains: Request parsing, validation, service calls, response formatting
- Key files:
  - `order.go` / `order_pg.go` - Order listing and public details
  - `pos.go` / `pos_pg.go` - POS device and order endpoints
  - `payment.go` / `payment_pg.go` - Payment intent creation and webhooks
  - `admin_*.go` - Admin dashboard operations (still MongoDB)
  - `station.go` / `station_pg.go` - Station verification and redemption
- Pattern: Each handler has a struct with dependencies, methods are handlers (func(w, r))

**bfs-backend/internal/http/**
- Purpose: Router configuration and HTTP setup
- Contains: Route definitions, middleware stacking, CORS setup
- Key files: `router.go` (main route tree)
- Structure: Chi router with grouped routes by version and resource

**bfs-backend/internal/middleware/**
- Purpose: HTTP middleware for cross-cutting concerns
- Contains: JWT validation, CSRF protection, security headers
- Key files: `jwt.go` (bearer token parsing), `csrf.go` (token validation)
- Applied in: `internal/http/router.go` via `r.Use(middleware)`

**bfs-backend/internal/auth/**
- Purpose: Authentication mechanisms
- Contains: NeonAuth middleware integration, JWT claims extraction
- Key files: `neon.go` (external auth provider)
- Pattern: Middleware extracts user context and stores in request context

**bfs-backend/internal/app/**
- Purpose: Dependency injection configuration and lifecycle
- Contains: FX provider functions, app initialization
- Key files:
  - `server.go` - HTTP server lifecycle hooks
  - `config.go` - Configuration loading and validation
  - `logger.go` - Zap logger setup
  - `database.go` - MongoDB connection (legacy)
  - `postgres_repositories.go` - GORM and all PostgreSQL repository providers
  - `repositories.go` - MongoDB repository providers (legacy)
  - `services.go` - Legacy MongoDB service providers
  - `services_pg.go` - PostgreSQL service providers
  - `handlers.go` - Handler providers
  - `router.go` - Router setup with all routes
- Pattern: Each file has `New*` function returning constructed dependency

**bfs-backend/internal/config/**
- Purpose: Configuration management
- Contains: Environment variable parsing and validation
- Key files: `config.go`
- Env vars: `APP_PORT`, `MONGO_URI` (legacy), `DATABASE_URL` (PostgreSQL), `JWT_*`, `STRIPE_*`, `PAYREXX_*`, etc.

**bfs-backend/internal/response/**
- Purpose: HTTP response formatting
- Contains: JSON writing, error response format (RFC 7807 ProblemDetails)
- Key files: `response.go`
- Functions: `WriteJSON()`, `WriteError()`, `WriteCreated()`

**bfs-backend/internal/payrexx/**
- Purpose: Payrexx payment gateway client
- Contains: HTTP client, HMAC-MD5 signing, gateway CRUD, webhook verification
- Key files: `client.go`, `gateway.go`, `webhook.go`
- Authentication: API key + HMAC-MD5 signature on requests

**bfs-backend/internal/observability/**
- Purpose: Tracing and metrics
- Contains: OpenTelemetry setup, request tracing middleware
- Key files: `tracing.go`, `chi.go` (Chi middleware integration)

**bfs-backend/db/flyway/migrations/**
- Purpose: Database schema versioning
- Contains: SQL migration files (V1__schema.sql, V3__payrexx_columns.sql, etc.)
- Naming: Flyway convention: V{number}__{description}.sql
- Usage: Run automatically on app startup via Flyway

**bfs-web-app/app/(site)/**
- Purpose: Public customer ordering interface
- Routes: `/` (home), `/food/checkout/*` (product selection), `/profile`
- Key components: Menu display, cart, checkout flow, order tracking
- Provider: `CartProvider` (global cart state)

**bfs-web-app/app/admin/**
- Purpose: Admin dashboard
- Routes: `/admin` (dashboard), `/admin/orders`, `/admin/products`, `/admin/users`, etc.
- Key pages: Order management, product CRUD, user management, session revocation
- Auth: Requires JWT bearer token with admin role

**bfs-web-app/app/pos/**
- Purpose: POS terminal interface
- Routes: `/pos` (single-page app)
- Key components: Product list with stock, order creation, payment methods (cash/card/TWINT)
- Auth: X-Pos-Token header (device token from backend)

**bfs-web-app/app/station/**
- Purpose: Station device interface (kiosk/display)
- Routes: `/station` (single-page app)
- Key components: Order pickup display, QR verification, item redemption
- Auth: X-Device-Token header

**bfs-web-app/components/ui/**
- Purpose: Reusable UI components (Radix UI + Tailwind CSS)
- Contains: Button, Dialog, Input, Select, Table, Tabs, etc.
- Pattern: Server Components by default, `"use client"` for interactivity
- Styling: Tailwind CSS with custom Tailwind config

**bfs-web-app/lib/api.ts**
- Purpose: API client wrapper
- Contains: Base fetch wrapper with auth header injection, error handling
- Endpoints: Wraps all backend `/v1/*` endpoints

**bfs-web-app/lib/auth/**
- Purpose: Client-side authentication utilities
- Contains: JWT token storage, extraction, refresh logic
- Files: Token management, Google OAuth helpers

**bfs-web-app/lib/server/**
- Purpose: Server-side utilities (Server Components only)
- Contains: Server-only API calls, secrets access
- Pattern: `"use server"` marked files

**bfs-web-app/contexts/cart-context.tsx**
- Purpose: Global shopping cart state
- Pattern: React Context + useReducer
- Shared: All (site) layout components access via `useCart()`

## Key File Locations

**Entry Points:**
- `bfs-backend/cmd/backend/main.go` - Backend HTTP server startup
- `bfs-web-app/app/layout.tsx` - Root layout (font, metadata)
- `bfs-web-app/app/(site)/layout.tsx` - Public site layout (header, footer, cart)
- `bfs-web-app/app/admin/layout.tsx` - Admin layout

**Configuration:**
- `bfs-backend/.env.example` - Backend environment template
- `bfs-web-app/.env.local.example` - Frontend environment template
- `bfs-backend/internal/config/config.go` - Config struct and loading
- `bfs-web-app/next.config.js` - Next.js build configuration

**Core Logic:**
- `bfs-backend/internal/service/order_pg.go` - Order business logic (PostgreSQL)
- `bfs-backend/internal/service/payment_pg.go` - Payment processing (Payrexx)
- `bfs-web-app/contexts/cart-context.tsx` - Shopping cart state management
- `bfs-web-app/lib/api.ts` - API client

**Testing:**
- `bfs-backend/test/integration/` - Integration tests
- `bfs-backend/test/e2e/` - End-to-end tests
- `bfs-web-app/__tests__/` or `*.test.ts` files - Vitest unit tests
- `bfs-web-app/e2e/` - Playwright E2E tests

**Database:**
- `bfs-backend/db/flyway/migrations/` - SQL schema versions
- `bfs-backend/internal/postgres/device_repo.go` - Device CRUD
- `bfs-backend/internal/model/` - GORM entity definitions

**Payments:**
- `bfs-backend/internal/payrexx/client.go` - Payrexx HTTP client
- `bfs-backend/internal/payrexx/gateway.go` - Gateway creation/update
- `bfs-backend/internal/payrexx/webhook.go` - Webhook signature verification
- `bfs-web-app/lib/stripe.ts` - Stripe Elements setup

## Naming Conventions

**Files:**
- Handlers: `{resource}.go` or `{resource}_pg.go` (e.g., `order.go`, `order_pg.go`)
- Services: `{feature}.go` or `{feature}_pg.go` (e.g., `payment.go`, `payment_pg.go`)
- Repositories: `{entity}_repo.go` (e.g., `order_repo.go`, `product_repo.go`)
- Models: `{entity}.go` (e.g., `order.go`, `product.go`)
- Pages (Next.js): `page.tsx` (route segments), `layout.tsx` (layouts), `route.ts` (API routes)
- Components (Next.js): PascalCase (e.g., `CartItem.tsx`, `AdminProductTable.tsx`)

**Directories:**
- Feature-based in handlers: Group by resource (`order.go`, `payment.go`) not by operation
- Version in routes: `/v1/` prefix for API versioning
- Layout groups: Parentheses `(site)`, `(auth)` for route organization (not shown in URL)

**Functions:**
- Handlers: `(h *TypeHandler) Method(w, r)` - receiver method pattern
- Services: `(s *serviceType) Method(ctx context.Context, ...)` - receiver method pattern
- Constructors: `New{Type}(deps...) *Type` (e.g., `NewOrderHandler`)
- Interfaces: `{Type}Interface` or just `{Type}` (e.g., `OrderRepository`)

**Constants & Types:**
- Domain models: PascalCase enums (`OrderStatusPending`, `UserRoleAdmin`)
- Type aliases: PascalCase (e.g., `type Cents int64`)
- Private types: lowercase (e.g., `type orderService struct`)

## Where to Add New Code

**New Feature (Order-related):**
- PostgreSQL model: `bfs-backend/internal/model/{entity}.go`
- Repository: `bfs-backend/internal/postgres/{entity}_repo.go` (implement interface in service package)
- Service: `bfs-backend/internal/service/{feature}_pg.go`
- Handler: `bfs-backend/internal/handler/{feature}_pg.go`
- Routes: Add under `v1.Route("/orders", ...)` in `bfs-backend/internal/http/router.go`
- Tests: `bfs-backend/test/integration/{feature}_test.go`

**New Page (Admin):**
- Route: Create `bfs-web-app/app/admin/{feature}/page.tsx`
- Layout (if needed): `bfs-web-app/app/admin/{feature}/layout.tsx`
- Components: Create `bfs-web-app/components/admin/{Feature}Table.tsx`, etc.
- API client: Add to `bfs-web-app/lib/api.ts` or create `bfs-web-app/lib/api/{feature}.ts`

**New Reusable Component:**
- Location: `bfs-web-app/components/` with feature subfolder
- Storybook: `bfs-web-app/components/{Feature}.stories.tsx`
- Tests: `bfs-web-app/components/{Feature}.test.tsx`
- Pattern: Prefer Server Components, mark interactive ones with `"use client"`

**New Utility Function:**
- Shared (backend): `bfs-backend/internal/utils/{feature}.go`
- Client-side: `bfs-web-app/lib/{feature}.ts` or `bfs-web-app/lib/{category}/{feature}.ts`
- Hooks: `bfs-web-app/hooks/use{Feature}.ts`

**New Migration (Database):**
- Location: `bfs-backend/db/flyway/migrations/V{next_number}__{description}.sql`
- Run: `make db-reset` (in bfs-backend)
- Pattern: Create tables, indexes, constraints using PostgreSQL syntax

## Special Directories

**bfs-backend/secrets/**
- Purpose: JWT key storage (development)
- Generated: Yes (by `make` setup)
- Committed: No (in .gitignore)
- Structure:
  - `secrets/dev/jwt-priv.pem` - Development private key
  - `secrets/dev/jwt-pub.pub.pem` - Development public key
  - `secrets/staging/` - Staging keys
  - `secrets/production/` - Production keys

**bfs-backend/tmp/**
- Purpose: Build artifacts and caches
- Generated: Yes (by `make dev`, Go toolchain)
- Committed: No (in .gitignore)
- Contents: Compiled binaries, Go module cache

**bfs-backend/docs/**
- Purpose: Generated Swagger/OpenAPI documentation
- Generated: Yes (by swag CLI)
- Committed: No (generated from godoc comments)
- Access: `GET /swagger/*` when deployed

**bfs-web-app/.next/**
- Purpose: Next.js build output and caching
- Generated: Yes (by `pnpm build` or `pnpm dev`)
- Committed: No (in .gitignore)

**bfs-web-app/node_modules/**
- Purpose: npm dependencies
- Generated: Yes (by `pnpm install`)
- Committed: No (in .gitignore)
- Lockfile: `pnpm-lock.yaml` (committed, pinned versions)

---

*Structure analysis: 2026-01-28*
