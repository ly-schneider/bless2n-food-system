# Coding Conventions

**Analysis Date:** 2026-01-28

## Naming Patterns

**Files:**
- Go backend: `snake_case` for files (e.g., `product_handler.go`, `admin_category.go`)
- TypeScript/React: `kebab-case` for files (e.g., `cookie-banner.tsx`, `auth-context.tsx`, `utils.test.ts`)
- Components and hooks: Use descriptive names with purpose indicator (e.g., `use-authorized-fetch`, `cookie-banner`)
- Test files: `[name].test.ts` or `[name].spec.ts` format

**Functions/Methods:**
- Go: CamelCase starting with lowercase for unexported, UPPERCASE for exported methods (Go convention)
  - Example: `func (h *ProductHandler) ListProducts(...)`
  - Example: `func NewProductHandler(...) *ProductHandler`
- TypeScript/React: camelCase for all functions and methods
  - Example: `const fetchAuth = useAuthorizedFetch()`
  - Example: `export function AuthProvider({ children }: ...)`
  - Example: `async () => { const res = await fetchAuth(...) }`

**Variables:**
- Go: camelCase (unexported), UPPERCASE constants
  - Example: `var categoryID *string`, `const JetonsCollection = "jetons"`
  - Example: `type ProductType string` followed by `const ProductTypeSimple ProductType = "simple"`
- TypeScript/React: camelCase for all variables
  - Example: `const [items, setItems] = useState<Product[]>([])`
  - Example: `const SESSION_KEY = "bfs.auth.session"` for constants
  - Example: `const NO_JETON_VALUE = "__none__"`

**Types:**
- Go interfaces: PascalCase, end with `Interface` or use action-based name
  - Example: `type ProductService interface { ListProducts(...) }`
  - Example: `type ProductRepository interface { FindByID(...) }`
- Go structs: PascalCase
  - Example: `type Product struct { ID bson.ObjectID }`, `type ProductDTO struct { ... }`
- TypeScript types: PascalCase using `type` or `interface` keywords
  - Example: `type AuthState = { accessToken: string | null }`
  - Example: `export interface Product { id: string; categoryId: string }`
  - Example: `type AuthContextType = AuthState & { setAuth: (...) => void }`

**Constants:**
- Go: SCREAMING_SNAKE_CASE in package scope
  - Example: `const JetonsCollection = "jetons"`
- TypeScript: SCREAMING_SNAKE_CASE for meaningful constants
  - Example: `const SESSION_KEY = "bfs.auth.session"`, `const COOKIE_NAME = "ga_consent"`

## Code Style

**Formatting:**
- Prettier is used for TypeScript/React (configured via `prettier.config.js`)
- Configuration:
  - `trailingComma: "es5"`
  - `tabWidth: 2`
  - `printWidth: 120`
  - `semi: false` (no semicolons)
  - Uses `prettier-plugin-tailwindcss` for automatic Tailwind class ordering
- Go: `gofmt` for automatic formatting (make target: `make fmt`)
- Run linting before commit

**Linting:**
- TypeScript/React: ESLint (config: `eslint.config.mjs`)
  - Engine: TypeScript ESLint with flat config
  - Extends: `@next/eslint-plugin-next`, `react-hooks`, `import`, `storybook`
  - Key rules:
    - `@typescript-eslint/no-unused-vars`: Warn for unused vars (allow `_` prefix for intentional ignores)
    - `sort-imports`: Error on unsorted imports
    - `import/order`: Warn on import organization (groups: external, builtin, internal, sibling, parent, index)
    - Import aliases in `tsconfig.json`: `@/*` maps to project root
- Go: golangci-lint (make target: `make lint`, `make lint-fix`)

## Import Organization

**Order (enforced by ESLint):**
1. External packages (node_modules)
2. Builtin modules
3. Internal paths (project source, including `@/` alias)
4. Sibling imports
5. Parent imports
6. Index imports

**Path Aliases (TypeScript):**
- `@/*` resolves to project root for absolute imports
- Example: `import { Button } from "@/components/ui/button"`
- Example: `import { useAuth } from "@/contexts/auth-context"`
- Example: `import type { User } from "@/types"`

**Go imports:**
- Group imports: standard library, third-party, internal packages
- Example:
  ```go
  import (
      "context"
      "net/http"
      "backend/internal/domain"
      "backend/internal/response"
      "github.com/go-playground/validator/v10"
      "go.uber.org/zap"
  )
  ```

## Error Handling

**TypeScript/React Patterns:**
- Async/await with try-catch blocks
  ```typescript
  try {
    const res = await fetch("/api/auth/refresh", { method: "POST" })
    if (!res.ok) {
      clearAuth()
      return false
    }
    const data = (await res.json()) as { access_token: string; expires_in: number }
    setAuth(data.access_token, data.expires_in)
    return true
  } catch {
    clearAuth()
    return false
  }
  ```
- Optional error handling (catch without variable when not used)
  ```typescript
  try {
    sessionStorage.setItem(SESSION_KEY, JSON.stringify(s))
  } catch {}
  ```
- HTTP error checking with explicit status codes
  ```typescript
  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`)
  }
  ```
- Error message parsing utility: `readErrorMessage(response)` for extracting details from failed responses

**Go Patterns:**
- Wrap errors with context using `fmt.Errorf`
  ```go
  if err := doSomething(); err != nil {
    return nil, fmt.Errorf("failed to check category: %w", err)
  }
  ```
- Explicit error checking with `if err != nil`
  ```go
  if err != nil {
    h.logger.Error("Failed to list products", zap.Error(err))
    response.WriteError(w, http.StatusInternalServerError, err.Error())
    return
  }
  ```
- Use `errors.New()` for simple error messages
  ```go
  if cat == nil {
    return nil, errors.New("category not found")
  }
  ```
- Error responses via `response.WriteError()` helper
  ```go
  response.WriteError(w, http.StatusInternalServerError, "failed to list categories")
  ```

## Logging

**Framework:**
- Go: Uber Zap (`go.uber.org/zap`)
- TypeScript/React: `console` (no structured logging)

**Go Patterns:**
- Use `zap.L()` for global logger or injected logger instance
- Include context with structured fields
  ```go
  h.logger.Error("Failed to list products", zap.Error(err))
  zap.L().Error("admin list categories failed", zap.Error(err), zap.String("method", r.Method), zap.String("path", r.URL.Path))
  ```
- Log at appropriate levels: Error for failures, Info for significant events
- Include request context when available: method, path, error details

**TypeScript/React Patterns:**
- Minimal logging (framework handles observability)
- Use console for development debug (minimal in production code)
- Example: Cookie banner consent tracking via custom events
  ```typescript
  window.dispatchEvent(new CustomEvent("ga-consent-changed", { detail: { value: true } }))
  ```

## Comments

**When to Comment:**
- Complex business logic that isn't self-evident
- Handler functions: Include godoc comments with Swagger annotations
- Non-obvious type constraints or format requirements
- Workarounds or temporary solutions

**JSDoc/TSDoc (Go):**
- Go: Required for exported functions/types with godoc format
  ```go
  // ListProducts godoc
  // @Summary List products
  // @Description List all products with optional category filtering
  // @Tags products
  // @Param category_id query string false "Filter by category ID"
  // @Success 200 {object} domain.ListResponse[domain.ProductDTO]
  // @Router /v1/products [get]
  func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request)
  ```
- JSDoc: Optional for TypeScript (used for complex utility functions)
  ```typescript
  // Formats a price in cents to Swiss Francs.
  // Examples:
  //  - 500  -> "CHF 5.-"
  //  - 550  -> "CHF 5.50"
  export function formatChf(cents: number): string
  ```

## Function Design

**Size:**
- Keep functions focused and single-purpose
- Go handlers typically 20-40 lines (request → service → response pattern)
- TypeScript/React hooks often use composition (useEffect, useState) with 30-60 line components

**Parameters:**
- Go: Use receiver methods on structs for dependency injection
  ```go
  func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request)
  ```
- TypeScript: Use props object destructuring for React components
  ```typescript
  export function AuthProvider({ children }: { children: React.ReactNode })
  ```
- TypeScript: Use callbacks for state setters and async functions
  ```typescript
  const setAuth = useCallback((t: string, expiresIn: number, u?: User) => { ... }, [])
  ```

**Return Values:**
- Go: Use tuples with error as last return value
  ```go
  func (r *productService) ListProducts(...) (*domain.ListResponse[domain.ProductDTO], error)
  ```
- TypeScript: Use typed returns with generics
  ```typescript
  const refresh = useCallback(async (): Promise<boolean> => { ... })
  ```
- React: Components return JSX elements, hooks return state/callbacks

## Module Design

**Exports:**
- Go: Exported names (PascalCase) are explicitly exported
  ```go
  type ProductService interface { ... }
  func NewProductHandler(...) *ProductHandler { ... }
  ```
- TypeScript: Use `export` keyword for public API
  ```typescript
  export function useAuth()
  export type AuthState = { ... }
  export default function CookieBanner()
  ```

**Barrel Files:**
- TypeScript: `types/index.ts` exports common types for convenience
  ```typescript
  export type { User, Product, Order }
  ```
- Used sparingly to avoid circular dependencies

## Type Safety

**TypeScript:**
- `strict: true` in tsconfig - all code must pass strict type checking
- `noUncheckedIndexedAccess: true` - prevents unsafe array/object access
- Use `satisfies` operator for type validation without narrowing
- Generic constraints for reusable utilities
  ```typescript
  export function cn(...inputs: ClassValue[])
  ```

**Go:**
- Use interfaces for dependency injection (not concrete types)
  ```go
  type ProductService interface { ... }
  type productService struct { productRepo ProductRepository }
  ```
- Use custom types for domain concepts
  ```go
  type ProductType string
  const ProductTypeSimple ProductType = "simple"
  ```

---

*Convention analysis: 2026-01-28*
