# Bless2n Food System

A full-stack food ordering platform built for live event food stations in Thun, Switzerland. Handles the complete lifecycle from customer ordering and TWINT payment processing to kitchen fulfillment, inventory tracking, and POS terminal operations.

## Architecture

```
                    ┌──────────────────────────────────────┐
                    │          Azure Container Apps        │
                    │        (auto-scale, scale-to-zero)   │
                    └──────────┬──────────────┬────────────┘
                               │              │
                    ┌──────────▼───┐    ┌─────▼────────────┐
                    │  Next.js 16  │    │    Go Backend    │
                    │  React 19    │    │    Chi Router    │
                    │  TypeScript  │◄──►│    Clean Arch    │
                    │  Tailwind 4  │    │    Uber FX DI    │
                    └──────┬───────┘    └──────┬───────────┘
                           │                   │
             ┌─────────────┼───────────────────┼───────────┐
             │             │                   │           │
      ┌──────▼─────────┐ ┌─▼──────┐  ┌─────────▼───┐  ┌────▼────────┐
      │ Better Auth    │ │ Sentry │  │ PostgreSQL  │  │   Payrexx   │
      │ (Sessions,     │ │        │  │   (NeonDB)  │  │   (TWINT)   │
      │  Google OAuth) │ │        │  │             │  │             │
      └────────────────┘ └────────┘  └─────────────┘  └─────────────┘
```

**Four distinct interfaces** serve different user roles:

| Interface | Path | Purpose |
|-----------|------|---------|
| Public Ordering | `/` | Customer-facing menu, cart, QR code ordering, TWINT checkout |
| Admin Dashboard | `/admin` | Order management, inventory, analytics, device configuration |
| POS Terminal | `/pos` | Cashier interface for on-site payments via SumUp |
| Station Display | `/station` | Kitchen/preparation queue for food station operators |

## Tech Stack

| Component | Stack |
|-----------|-------|
| **Backend** | Go 1.25, Chi, PostgreSQL, Ent ORM, Atlas Migrations, Uber FX |
| **Web App** | Next.js 16, React 19, TypeScript, Tailwind CSS 4, Radix UI |
| **Android** | Kotlin, Jetpack Compose, SumUp SDK, ZXing, Thermal Printer |
| **Infrastructure** | Terraform, Azure Container Apps, NeonDB, Key Vault, Blob Storage |
| **Auth** | Better Auth (sessions + Google OAuth), JWKS validation, RBAC |
| **Payments** | Payrexx (TWINT), SumUp (card terminal) |
| **CI/CD** | GitHub Actions, trunk-based development, immutable image promotion |
| **Observability** | Sentry, OpenTelemetry, Uber Zap, Azure Log Analytics |

## Repository Structure

```
bfs-backend/          Go HTTP API — Clean Architecture, Payrexx payments, RBAC
bfs-web-app/          Next.js web app — ordering, admin, POS, station interfaces
bfs-android-app/      Android POS terminal — SumUp payments, thermal printing
bfs-cloud/            Terraform IaC — Azure Container Apps, auto-scaling
bfs-docs/             Documentation site — Fumadocs, admin handbook, dev guide
```

## Quick Start

### Prerequisites

- Go 1.25+
- Node.js 20+ with pnpm
- Docker & Docker Compose

### Full Stack Development

```bash
# Terminal 1 — Backend
cd bfs-backend
cp .env.example .env
make docker-up
make migrate-local
make seed
make dev              # API at http://localhost:8080

# Terminal 2 — Web App
cd bfs-web-app
pnpm install
pnpm dev              # App at http://localhost:3000
```

## Deployment

The project uses a **trunk-based, immutable-tag release pipeline**:

1. **Push to `main`** triggers staging deployment — Docker images are built, migrations run, infrastructure updated via Terraform Cloud
2. **Create a GitHub Release** (`vX.Y.Z`) triggers production — the exact staging image is promoted via `crane tag` (no rebuild, byte-for-byte identical)
3. After production deploy, the `VERSION` file is auto-bumped and committed with `[skip cd]`

A single `VERSION` file is the source of truth across all projects, image tags, and git tags.

### Environments

| Environment | Trigger | Scaling |
|-------------|---------|---------|
| **Staging** | Push to `main` | 0-20 replicas, lower thresholds |
| **Production** | GitHub Release `vX.Y.Z` | 0-20 replicas, tuned thresholds |

Both environments support **scale-to-zero** for cost efficiency.

## Key Engineering Decisions

- **Better Auth over custom auth**: Session-based authentication managed in Next.js, backend validates via JWKS — clean separation of concerns
- **Payrexx for Swiss payments**: Native TWINT support with HMAC-verified webhooks
- **Atlas + Ent for migrations**: Declarative schema in Go, versioned SQL migrations, CI linting
- **Immutable releases**: Production always deploys the exact same image digest that ran in staging
- **Clean Architecture in Go**: Strict layer separation with Uber FX dependency injection
- **Scale-to-zero Container Apps**: Zero cost when idle, automatic burst scaling under load

## Project Documentation

Detailed READMEs are available in each sub-project:

- [`bfs-backend/README.md`](bfs-backend/README.md) — API architecture, development setup, testing
- [`bfs-web-app/README.md`](bfs-web-app/README.md) — Frontend architecture, routes, component system
- [`bfs-android-app/README.md`](bfs-android-app/README.md) — Android POS, build variants, hardware integration
- [`bfs-cloud/README.md`](bfs-cloud/README.md) — Infrastructure modules, scaling, deployment phases
- [`bfs-docs/README.md`](bfs-docs/README.md) — Documentation site setup and content structure
- [`bfs-http/README.md`](bfs-http/README.md) — API request collections for testing

## License

This project is source-available. All rights reserved. See [LICENSE](LICENSE) for details.
