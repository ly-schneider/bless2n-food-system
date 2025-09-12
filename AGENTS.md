# Repository Guidelines

## Project Structure & Module Organization
- `ordering-system/`: Next.js + Supabase ordering UI (App Router). Code in `app/`, UI in `components/`, shared in `lib/` and `utils/`, DB seeds/migrations in `supabase/`.
- `bfs-backend/`: Go HTTP API with Clean Architecture. Entry at `cmd/backend/`, core in `internal/` (domain, service, http, etc.). Docker and `Makefile` provided.
- `webpos2-main/`: Next.js POS client. Source in `src/`, public assets in `public/`.
- `bfs-cloud/`: Terraform IaC (envs under `envs/`).
- `bfs-http/`: Yaak/HTTP client collections for API testing.

## Build, Test, and Development Commands
- Ordering UI: `cd ordering-system && npm run dev|build|start` (uses Turbopack). Env from `ordering-system/.env`.
- POS app: `cd webpos2-main && npm run dev|build|start|lint`.
- Backend API: `cd bfs-backend && make help`.
  - `make dev`: run API with live‑reload (Air).
  - `make docker-up` / `make docker-up-backend`: start services (Mongo, Mailpit) with/without API.
  - `make test` / `make test-coverage`: unit tests; coverage threshold 80%.

## Coding Style & Naming Conventions
- TypeScript/React: 2‑space indent, functional components, hooks prefix `use*`. Prefer `kebab-case` filenames for components and `camelCase` for utilities. Run `npm run lint` where available.
- Formatting: Prettier is used (see project configs). Keep imports ordered and avoid unused exports.
- Go: idiomatic Go formatting (`gofmt`, `goimports`). Package layout mirrors `internal/{domain,service,http,...}`.
- Commits: Conventional Commits enforced via pre‑commit (`feat:`, `fix:`, `refactor:`, etc.; see `git-conventional-commits.yaml`).

## Testing Guidelines
- Go unit tests live alongside code in `*_test.go`. Use table‑driven tests and subtests. Run `make test` locally; for CI, prefer `make test-coverage` (fails if <80%).
- Frontend: add tests where applicable (e.g., React Testing Library). Name files `*.test.tsx` near the component.

## Commit & Pull Request Guidelines
- Commits: concise imperative subject, scope when helpful (`feat(auth): ...`). Avoid "WIP"; squash trivial commits.
- PRs: include clear description, linked issues, setup steps, and screenshots for UI. Note affected packages (`ordering-system`, `bfs-backend`, etc.). Keep changes atomic per app.

## Security & Configuration Tips
- Never commit secrets. Use `.env` files and templates (`ordering-system/.env.example`, `bfs-backend/.env.example`).
- Backend keys: generate Ed25519 dev keys under `bfs-backend/secrets/dev/` (see backend README).
- Supabase: run SQL in `ordering-system/supabase/` to seed/migrate as needed.
