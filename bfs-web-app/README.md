# BFS Web App

Next.js web application serving four distinct interfaces for the Bless2n Food System — public ordering, admin dashboard, POS terminal, and station display.

## Architecture

Built on **Next.js 16 with React 19 Server Components**, the app leverages the App Router for nested layouts and server-side data fetching. Authentication is handled entirely within the app via Better Auth, exposing a JWKS endpoint for the Go backend to validate sessions.

```
app/
├── (auth)/         Authentication flows (login, register, password reset)
├── (site)/         Public ordering — menu browsing, cart, QR code ordering, TWINT checkout
├── admin/          Admin dashboard — orders, inventory, analytics, devices, settings
├── pos/            POS terminal — cashier interface for on-site SumUp payments
├── station/        Station display — kitchen preparation queue for food operators
├── health/         Health check endpoint
└── api/auth/       Better Auth API routes (sessions, OAuth, JWKS)
```

### Key Design Decisions

- **React 19 Server Components** for server-side data fetching and reduced client bundle
- **Better Auth** as the single auth layer — sessions, Google OAuth, email/password, JWKS endpoint
- **Radix UI + Tailwind CSS 4** — accessible primitives with utility-first styling
- **React Hook Form + Zod** — type-safe form validation
- **QR code integration** — html5-qrcode for scanning, qrcode for generation
- **Sentry** for production error tracking with source maps

## Tech Stack

| Category | Technology |
|----------|-----------|
| Framework | Next.js 16 (App Router) |
| UI | React 19, Server Components |
| Language | TypeScript 5.8 (strict) |
| Styling | Tailwind CSS 4 |
| Components | Radix UI |
| Forms | React Hook Form + Zod 4 |
| Auth | Better Auth (sessions, Google OAuth) |
| Charts | Recharts |
| QR Codes | html5-qrcode, qrcode |
| Notifications | Sonner |
| Dates | date-fns |
| WebSocket | ws |
| Caching | Redis |
| Error Tracking | Sentry |
| Testing | Vitest, React Testing Library, Playwright |
| Component Dev | Storybook 9 |
| Package Manager | pnpm 10 |
| Container | Distroless Node.js (non-root) |

## Project Structure

```
bfs-web-app/
├── app/                  Next.js App Router (routes above)
├── components/           Reusable React components
├── hooks/                Custom React hooks
├── lib/                  Utilities and helpers
├── contexts/             React Context providers
├── styles/               Global styles
├── public/               Static assets
├── .storybook/           Storybook configuration
├── next.config.ts        Next.js config (Sentry, bundle analyzer)
├── playwright.config.ts  E2E test config
├── Dockerfile            Multi-stage distroless build
└── package.json
```

## Prerequisites

- **Node.js 20+**
- **pnpm** (specified in package.json)
- Running backend for full functionality (see root README)

## Development Setup

```bash
pnpm install
pnpm dev                # http://localhost:3000
```

### Application Interfaces

| Interface | URL | Description |
|-----------|-----|-------------|
| Public Ordering | http://localhost:3000 | Customer-facing menu and checkout |
| Admin Dashboard | http://localhost:3000/admin | Order management, inventory, analytics |
| POS Terminal | http://localhost:3000/pos | Cashier SumUp payment interface |
| Station Display | http://localhost:3000/station | Kitchen preparation queue |

## Commands

### Development

```bash
pnpm dev                # Start dev server with hot reload
pnpm build              # Production build
pnpm start              # Start production server
```

### Quality

```bash
pnpm typecheck          # TypeScript type checking
pnpm lint               # ESLint
pnpm lint:fix           # Auto-fix lint issues
pnpm prettier           # Check formatting
pnpm prettier:fix       # Fix formatting
```

### Testing

```bash
pnpm test               # Vitest unit tests
pnpm test:watch         # Watch mode
pnpm test:coverage      # Coverage report
pnpm e2e:headless       # Playwright E2E tests
pnpm e2e:ui             # Playwright UI mode
```

### Storybook

```bash
pnpm storybook          # Start on port 6006
pnpm build-storybook    # Build static Storybook
pnpm test-storybook     # Run Storybook tests
```

### Analysis

```bash
pnpm analyze            # Bundle size analysis
```

## Authentication

Better Auth manages all authentication within the web app:

- **Session-based auth** with secure cookies
- **Google OAuth** via Better Auth social providers
- **Email/password** registration and login
- **JWKS endpoint** (`/api/auth/jwks`) consumed by the Go backend for token validation

The backend never handles auth directly — it validates tokens by fetching the public keys from the JWKS endpoint.

## Environment Variables

Key configuration (see `.env.example` for full list):

| Variable | Purpose |
|----------|---------|
| `NEXT_PUBLIC_API_BASE_URL` | Backend API URL (client-side) |
| `BACKEND_INTERNAL_URL` | Backend URL (server-side, bypasses public network) |
| `BETTER_AUTH_SECRET` | Session signing secret |
| `DATABASE_URL` | PostgreSQL for Better Auth |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret |
| `SENTRY_AUTH_TOKEN` | Sentry source map uploads |
