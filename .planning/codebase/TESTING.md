# Testing Patterns

**Analysis Date:** 2026-01-28

## Test Framework

**Runner (TypeScript/React):**
- Vitest (version ^4.0.16)
- Config: `vitest.config.ts`
- Environment: jsdom (browser DOM simulation)
- Setup file: `vitest.setup.ts` (imports Testing Library jest-dom matchers)

**Runner (Go Backend):**
- Go's built-in `testing` package
- Run via `make test` or `go test -v -race ./internal/...`
- Standard table-driven test pattern

**Assertion Library:**
- TypeScript/React: Vitest's `expect()` with Testing Library matchers
- Go: No external assertion library, using standard `if` statements and `t.Fatalf()`

**Run Commands (TypeScript/React):**
```bash
pnpm test                 # Run all tests once
pnpm test:watch          # Watch mode
pnpm test:ui             # UI mode with visual dashboard
pnpm test:coverage       # Coverage report
```

**Run Commands (Go):**
```bash
make test                 # Run all tests with race detector
go test ./internal/...    # Run specific package tests
```

## Test File Organization

**Location (TypeScript/React):**
- Co-located with source files in same directory
- Example: `lib/utils.test.ts` sits alongside `lib/utils.ts`
- E2E tests: Separate `e2e/` directory at project root

**Location (Go):**
- Co-located with source code in same package
- Example: `internal/utils/hash_test.go` in same `utils` package as `hash.go`

**Naming (TypeScript/React):**
- Pattern: `[filename].test.ts` or `[filename].test.tsx`
- Example: `utils.test.ts`, `lib/utils.test.ts`
- Playwright E2E: Any files in `e2e/` directory matching glob pattern `**/*test.ts`

**Naming (Go):**
- Pattern: `[filename]_test.go`
- Example: `hash_test.go`, `crypto_test.go`
- Keep test files in same package as code

**Structure:**
```
bfs-web-app/
├── lib/
│   ├── utils.ts
│   └── utils.test.ts       # Co-located test
├── app/
│   └── health/
│       └── route.ts
└── e2e/
    └── [e2e test files]    # Playwright tests

bfs-backend/
├── internal/
│   └── utils/
│       ├── hash.go
│       ├── hash_test.go    # Co-located test
│       ├── crypto.go
│       └── crypto_test.go
```

## Test Structure

**TypeScript/React Pattern:**
```typescript
import { describe, expect, it } from "vitest"
import { cn, formatChf } from "./utils"

describe("utils", () => {
  it("cn merges class names", () => {
    expect(cn("a", false && "b", "c")).toBe("a c")
  })

  it("formatChf formats cents to CHF", () => {
    expect(formatChf(0)).toContain("CHF")
    expect(formatChf(1990)).toMatch(/19/) // 19.90
  })
})
```

**Go Pattern (Table-Driven Tests):**
```go
package utils

import "testing"

func TestHashAndVerifyOTPArgon2(t *testing.T) {
  code := "123456"
  phc, err := HashOTPArgon2(code)
  if err != nil {
    t.Fatalf("hash error: %v", err)
  }
  if phc == "" {
    t.Fatalf("empty phc string")
  }
  ok, err := VerifyOTPArgon2(code, phc)
  if err != nil {
    t.Fatalf("verify error: %v", err)
  }
  if !ok {
    t.Fatalf("expected verification ok")
  }
  ok2, _ := VerifyOTPArgon2("000000", phc)
  if ok2 {
    t.Fatalf("expected verification to fail for wrong code")
  }
}

func TestGenerateRandomURLSafe(t *testing.T) {
  s, err := GenerateRandomURLSafe(16)
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if s == "" {
    t.Fatalf("expected non-empty string")
  }
  if strings.ContainsAny(s, "+/= ") {
    t.Fatalf("string is not URL-safe: %q", s)
  }
  if _, err := base64.RawURLEncoding.DecodeString(s); err != nil {
    t.Fatalf("not base64url: %v", err)
  }
}
```

**Setup/Teardown:**
- TypeScript: Use `beforeEach()` / `afterEach()` from Vitest for setup
- Go: Use `t.Setup()` or setUp functions in test functions (inline setup common)

**Assertion Pattern:**
- TypeScript: `expect(value).toBe(expected)`, `expect(value).toMatch(pattern)`, `expect(value).toContain(substring)`
- Go: Manual checking with `if` statements, fail with `t.Fatalf()`

## Mocking

**Framework (TypeScript/React):**
- Vitest's built-in mocking via `vi.mock()` (Vitest API)
- No Jest, uses Vitest which has compatible mocking

**Patterns (TypeScript/React):**
```typescript
// Mock modules
vi.mock("@/lib/api", () => ({
  fetchData: vi.fn(),
}))

// Mock functions
const mockFetch = vi.fn().mockResolvedValue({
  ok: true,
  json: async () => ({ data: "test" }),
})
```

**Framework (Go):**
- No mocking framework - use interfaces and dependency injection
- Create mock structs implementing expected interfaces

**Patterns (Go):**
```go
type MockProductRepository struct {
  // Implement ProductRepository interface
}

func (m *MockProductRepository) ListProducts(ctx context.Context, ...) (*domain.ListResponse, error) {
  // Return test data
}
```

**What to Mock:**
- TypeScript: External API calls, timers, browser APIs
- Go: Repository/database calls, external services
- Both: Network requests in unit tests

**What NOT to Mock:**
- Core business logic that's being tested
- Utility functions (pure functions with no side effects)
- Standard library functions
- Go: Actual error returns from repositories in happy path tests

## Fixtures and Factories

**Test Data (TypeScript/React):**
- Currently minimal fixture usage
- Inline test data creation in test files
- Example: Simple object creation for assertions
  ```typescript
  const testProduct = { id: "123", name: "Test", priceCents: 1000 }
  ```

**Test Data (Go):**
- No fixtures directory currently established
- Inline data creation in test functions
- Example: Creating domain objects for testing
  ```go
  product := &domain.Product{
    ID:       bson.NewObjectID(),
    Name:     "Test Product",
    PriceCents: Cents(1000),
  }
  ```

**Location for Fixtures:**
- TypeScript: Would go in `test/fixtures/` or co-located as helper files
- Go: Would go in `test/fixtures/` or inline in test files

## Coverage

**Requirements:**
- No hard coverage threshold enforced in CI
- Coverage report available via `pnpm test:coverage` (TypeScript)
- Coverage data generated to HTML report

**View Coverage:**
```bash
# TypeScript
pnpm test:coverage     # Run with coverage, generates report in coverage/

# Go
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Current State:**
- TypeScript: Minimal test coverage (~1 test file only)
- Go: Sparse coverage (2 test files in `internal/utils/`)
- Most coverage gaps in handlers, services, repositories

## Test Types

**Unit Tests:**
- TypeScript/React: Test utility functions, hooks in isolation
  - Location: `lib/utils.test.ts`
  - Scope: Pure functions, component utilities
  - Approach: Small, focused tests with mocked dependencies
- Go: Test individual functions and types
  - Location: `internal/*/` with `_test.go` files
  - Scope: Hash/crypto utilities, domain logic
  - Approach: Table-driven tests, explicit error checks

**Integration Tests:**
- TypeScript/React: Not currently implemented
  - Would test API calls, context providers, multi-component flows
- Go: Not currently implemented in standard locations
  - Would test service layer with real repositories or mocks
  - Potential location: `test/integration/`

**E2E Tests:**
- Framework: Playwright
- Config: `playwright.config.ts`
- Location: `e2e/` directory
- Scope: Full user flows across web application
- Base URL: `http://127.0.0.1:3000` (localhost dev server)
- Browser support: Chromium (CI), Chromium/Firefox/Safari (local)
- Retry: 2 retries on CI only
- Parallelization: Parallel local, single worker on CI
- Tracing: Enabled on first retry for debugging
- Run command: `pnpm e2e:headless` (CI) or `pnpm e2e:ui` (interactive)

## Common Patterns

**Async Testing (TypeScript):**
```typescript
// Await async functions in tests
it("fetches data", async () => {
  const result = await someAsyncFunction()
  expect(result).toBeDefined()
})

// Mock async responses
vi.mock("@/lib/api", () => ({
  fetchData: vi.fn().mockResolvedValue({ success: true }),
}))
```

**Async Testing (Go):**
```go
// Pass context.Background() or context.TODO() to functions
func TestListProducts(t *testing.T) {
  ctx := context.Background()
  result, err := service.ListProducts(ctx, nil, 50, 0)
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if result == nil {
    t.Fatalf("expected non-nil result")
  }
}
```

**Error Testing (TypeScript):**
```typescript
it("handles errors gracefully", () => {
  const mockFetch = vi.fn().mockRejectedValue(new Error("Network error"))
  expect(async () => {
    await fetchWithMock()
  }).rejects.toThrow()
})
```

**Error Testing (Go):**
```go
func TestErrorHandling(t *testing.T) {
  result, err := someFunction()
  if err == nil {
    t.Fatalf("expected error, got nil")
  }
  if result != nil {
    t.Fatalf("expected nil result on error")
  }
}
```

## Storybook Testing

**Framework:** Storybook 9.1.17 with test-runner
**Commands:**
```bash
pnpm storybook           # Start Storybook on port 6006
pnpm build-storybook     # Build static Storybook
pnpm test-storybook      # Run Storybook tests (requires storybook running)
```

**Purpose:** Component-driven development, visual testing, documentation
**Location:** Component story files alongside components
**Pattern:** Define UI variations via stories, automatically test rendering

## Coverage Targets

**TypeScript/React:**
- Critical paths: Hooks, utilities, API integration helpers
- Components: Render tests, state management
- Gap: Most UI components untested; E2E tests cover user flows

**Go:**
- Critical: Crypto/hash utilities, domain validation
- Gaps: No service layer tests, no repository tests, no handler tests
- Needed: Integration tests for business logic, table-driven tests for handlers

## Test Environment Setup

**TypeScript/React (`vitest.setup.ts`):**
```typescript
import "@testing-library/jest-dom"  // Testing Library matchers
```

**Go (`go.mod`):**
- Uses standard `testing` package, no setup required
- Requires context for async operations

---

*Testing analysis: 2026-01-28*
