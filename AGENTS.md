# Repository Guidelines

## Project Structure & Module Organization
- `bfs-backend/`: Go HTTP API (Clean Architecture). Entry at `cmd/backend/`; core under `internal/{domain,service,http,...}`; docs in `docs/`; dev secrets in `secrets/`. Tests live next to code as `*_test.go`.
- `bfs-web-app/`: Next.js app (App Router). UI in `app/`, components in `components/`, shared code in `lib/` and `hooks/`, assets in `public/`.
- `bfs-cloud/`: Terraform IaC. Environments under `envs/{staging,prod}`; reusable modules in `modules/`.
- `bfs-http/`: Yaak/HTTP client collections (`yaak.*.yaml`) for API testing.

## Build, Test, and Development Commands
- Backend: `cd bfs-backend && make help`
  - `make dev`: run API with live‑reload (Air).
  - `make docker-up` / `make docker-up-backend`: start Mongo(+API) via Compose.
  - `make test`: run unit tests; `make swag` to refresh Swagger docs.
- Web app: `cd bfs-web-app`
  - `pnpm dev|build|start|lint` and `pnpm test|test:coverage`.
- Cloud: `cd bfs-cloud/envs/<env> && terraform init && terraform plan|apply`.

## Coding Style & Naming Conventions
- TypeScript/React: 2‑space indent, functional components. Hooks prefixed `use*`.
  - Filenames: components `kebab-case`, utilities `camelCase`. Keep imports ordered.
  - Tools: ESLint + Prettier (`pnpm lint`, `pnpm prettier:fix`).
- Go: idiomatic formatting via `gofmt`/`goimports`; `golangci-lint` via `make lint`.
  - Package layout mirrors `internal/{domain,service,http,...}`.

## Testing Guidelines
- Go: table‑driven tests in `*_test.go`. Run `make test`; target 80%+ coverage for changed packages.
- Web app: Vitest + React Testing Library. Name files `*.test.ts(x)` near subjects. Run `pnpm test` or `pnpm test:coverage`.

## Commit & Pull Request Guidelines
- Commits: Conventional Commits (e.g., `feat(auth): add OTP flow`, `fix: address nil pointer`). Squash trivial fixes; avoid "WIP".
- PRs: include description, linked issues, setup steps, and UI screenshots when applicable. Call out affected packages (e.g., `bfs-backend`, `bfs-web-app`). Keep changes atomic per app.

## Security & Configuration Tips
- Never commit secrets. Use `.env` files and provided templates (e.g., `bfs-backend/.env.example`).
- Backend dev keys go under `bfs-backend/secrets/dev/` (see backend README).
- Prefer local DB via `make docker-up`; clean with `make docker-down` or `docker-down-v` (data loss).

