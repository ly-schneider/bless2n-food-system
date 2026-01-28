# Codebase Concerns

**Analysis Date:** 2026-01-28

## Tech Debt

**Incomplete PostgreSQL Migration (Phase 5.5 Blocker):**
- Issue: Admin handlers still use MongoDB repositories while core services have migrated to PostgreSQL
- Files: `bfs-backend/internal/handler/admin_product.go`, `admin_menu.go`, `admin_order.go`, `admin_user.go`, `admin_station.go`, `admin_pos.go`, `admin_invite.go`, `admin_sessions.go`
- Impact:
  - 73 files still reference MongoDB (repositories, domain types with BSON tags, seed data)
  - Increases operational complexity (must run both MongoDB and PostgreSQL)
  - Migration cannot complete until admin handlers are rewritten for PostgreSQL
  - Schema changes require updating both databases
- Fix approach: Create PostgreSQL versions of all 8 admin handlers (Phase 5.5), update DI wiring in `bfs-backend/internal/http/router.go`, then delete MongoDB entirely (Phase 6)

**Dual Database Configuration:**
- Issue: Both MongoDB and PostgreSQL connection strings required and initialized in `bfs-backend/internal/config/config.go`
- Files: `bfs-backend/internal/config/config.go` (lines 42-45, 125-127), `bfs-backend/main.go`
- Impact:
  - Increased infrastructure burden
  - Both databases must be monitored, backed up, and scaled
  - Connection pool management overhead
  - Confusing for developers which database handles what
- Fix approach: After Phase 5.5 is complete, remove MongoDB config entirely in Phase 6

**Legacy Stripe Payment Provider Coexisting with Payrexx:**
- Issue: Both Stripe and Payrexx SDKs loaded, but only Payrexx should be used for new payments
- Files: `bfs-backend/internal/service/payment.go` (Stripe), `bfs-backend/internal/service/payment_pg.go` (Payrexx), `bfs-backend/internal/config/config.go` (both configs loaded)
- Impact:
  - Confusing which payment system handles what
  - POS payment handlers use Payrexx while old Stripe code remains
  - Dependencies not removed from `go.mod`
- Fix approach: Phase 6 removal plan includes removing Stripe SDK, but requires completing admin handler migration first

**Unused PostgreSQL Repository Imports:**
- Issue: PostgreSQL repositories passed but unused in router initialization
- Files: `bfs-backend/internal/http/router.go` (lines 54-66) - all marked with underscore prefix `_`
- Impact: Confusing DI wiring, unclear future intent
- Fix approach: Remove unused imports once repositories are actively used

## Known Bugs

**Android WebView Security Gap:**
- Symptoms: POS activity accepts arbitrary URLs and loads them with JavaScript enabled and no origin verification
- Files: `bfs-android-app/app/src/main/java/ch/leys/bless2n/PosWebActivity.kt` (line 43)
- Trigger: Any malicious app can launch PosWebActivity with crafted URL, executing arbitrary JavaScript in WebView context
- Impact: Medium - WebView is restricted by `allowFileAccess=false` and `allowContentAccess=false`, but network attacks possible
- Workaround: Currently none; default URL points to example domain
- Fix approach: Implement whitelist validation for allowed origins, only load HTTPS URLs, validate against config

**Android Print Function Stub:**
- Symptoms: `PosBridge.print()` defined but not implemented (empty body)
- Files: `bfs-android-app/app/src/main/java/ch/leys/bless2n/PosWebActivity.kt` (line 74-76)
- Trigger: When POS calls JavaScript print() method
- Impact: Low - silently ignores print requests, no crash, but receipts won't print
- Workaround: Use manual cash/card payment flow without printing
- Fix approach: Implement ESC/POS rendering and send via Bluetooth RFCOMM (see TODO comment referencing MainActivity.printReceipt)

**Android Activity Callback Missing:**
- Symptoms: SumUp SDK payment results not returned to WebView
- Files: `bfs-android-app/app/src/main/java/ch/leys/bless2n/PosWebActivity.kt` (line 82)
- Trigger: When user completes payment in SumUp activity
- Impact: Medium - Payment may complete but JavaScript never notified, UI appears stuck
- Workaround: None; user must manually verify payment
- Fix approach: Implement onActivityResult mapping to call `evaluateJavascript()` to notify WebView

**Silent Exception Swallowing in Android Payment:**
- Symptoms: payWithCard method catches all exceptions and silently ignores them
- Files: `bfs-android-app/app/src/main/java/ch/leys/bless2n/PosWebActivity.kt` (line 70)
- Trigger: Any error parsing JSON or calling SumUp API
- Impact: Medium - Errors invisible to user and developers, difficult to debug payment issues
- Workaround: None
- Fix approach: Log caught exceptions and notify JavaScript layer of payment errors

## Security Considerations

**WebView JavaScript Interface Exposure:**
- Risk: `PosBridge` JavaScript interface exposes native payment and print functions to any loaded URL
- Files: `bfs-android-app/app/src/main/java/ch/leys/bless2n/PosWebActivity.kt` (line 41)
- Current mitigation: Restricted file access (`allowFileAccess=false`), enforced HTTPS content checking (`mixedContentMode = MIXED_CONTENT_NEVER_ALLOW`)
- Recommendations:
  - Implement origin/domain whitelist validation before loading URL
  - Sign payment requests with HMAC to prevent tampering
  - Validate JSON payload structure before parsing
  - Log all payment attempts for audit trail

**JWT Key Generation and Management:**
- Risk: JWT keys generated at dev startup using shell commands, must exist before backend runs
- Files: `bfs-backend/` (requires manual `openssl` commands per CLAUDE.md)
- Current mitigation: Keys stored in `.env` or `secrets/dev/` directory
- Recommendations:
  - Document key rotation process
  - Implement JWT key versioning for rotation without downtime
  - Consider key expiration and refresh strategy
  - Ensure dev keys are never committed to production

**NeonAuth URL Dependency:**
- Risk: NeonAuth JWKS discovery requires external service; no fallback for offline operation
- Files: `bfs-backend/internal/config/config.go` (line 130), `bfs-backend/internal/auth/context.go`
- Current mitigation: NEON_AUTH_URL is optional but used when provided
- Recommendations:
  - Cache JWKS locally with TTL to handle brief outages
  - Implement circuit breaker for JWKS endpoint
  - Document failure modes and retry logic

**SQL Injection via Pattern Matching:**
- Risk: LIKE pattern concatenation in queries (though parameterized)
- Files: `bfs-backend/internal/postgres/order_repo.go` (search pattern), `product_repo.go`
- Current mitigation: Using parameterized queries with `?` placeholders (GORM)
- Recommendations: Continue using parameterized queries; avoid raw SQL

## Performance Bottlenecks

**Large File Uploads Without Streaming:**
- Problem: Email service may load large attachment files entirely into memory
- Files: `bfs-backend/internal/service/email.go` (line 300)
- Current state: Not identified in code, potential issue if implemented
- Improvement path: Implement streaming file reads, chunked uploads

**N+1 Query Problems in Admin Handlers:**
- Problem: Listing operations may fetch individual items in loops without batch loading
- Files: `bfs-backend/internal/handler/admin_menu.go` (563 lines), `admin_product.go` (461 lines)
- Cause: MongoDB repository methods called per item without eager loading
- Improvement path: Add batch fetch methods to repositories, eager load relationships

**MongoDB to PostgreSQL Migration Performance Risk:**
- Problem: Admin handlers still query MongoDB which has 73 file dependencies; queries unoptimized for PostgreSQL
- Files: All admin handlers still using `repository` package (MongoDB)
- Cause: Incomplete migration; no indexes optimized for new access patterns
- Improvement path: Complete migration to PostgreSQL, create appropriate indexes post-migration

**Payment Webhook Processing Without Queue:**
- Problem: Stripe webhook handler processes synchronously; high-volume payments could block
- Files: `bfs-backend/internal/handler/payment.go` (line 308), Payrexx webhook in `payment_pg.go`
- Cause: Webhook handlers directly update database
- Improvement path: Implement async job queue (Redis, Bull, etc.) for webhook processing

## Fragile Areas

**Admin Menu Handler:**
- Files: `bfs-backend/internal/handler/admin_menu.go` (563 lines - largest handler)
- Why fragile:
  - Very large single file handling menu CRUD, slots, items, reordering
  - Complex state management (slot reordering with ordering field)
  - Multiple repository interactions without transactions
  - MongoDB-based, will require major refactor for PostgreSQL migration
- Safe modification: Add comprehensive tests before changes; implement transaction wrapper
- Test coverage: 0 tests (2 test files total in entire backend, only for utils)

**Payment Service with Dual Providers:**
- Files: `bfs-backend/internal/service/payment.go` (757 lines - largest service), `payment_pg.go` (429 lines)
- Why fragile:
  - Complex state machine for order preparation, payment intent creation, cleanup
  - Handles both Stripe (legacy) and Payrexx (current) with different flows
  - No idempotency for POS cash/card payments (only Payrexx has idempotency table)
  - Multiple cleanup operations that could leave orphaned data
- Safe modification: Write integration tests for payment flows; test with real webhooks locally
- Test coverage: 0 payment tests

**Admin Role Authorization Logic:**
- Files: `bfs-backend/internal/http/router.go` (lines 149-170)
- Why fragile:
  - Inline middleware checking token role AND falling back to database lookup
  - Falls back to MongoDB user repository during RBAC check
  - Role change effect timing unclear (token vs DB consistency)
  - Mixed NeonAuth token claims with local DB checks
- Safe modification: Extract to dedicated middleware, add integration tests for role changes
- Test coverage: 0 role authorization tests

**Idempotency Implementation - Incomplete:**
- Files: `bfs-backend/internal/model/idempotency.go`, `internal/postgres/idempotency_repo.go`, `internal/service/payment_pg.go`
- Why fragile:
  - Idempotency table created but not wired into all payment flows
  - POS cash/card payments lack idempotency handling
  - No cleanup job for expired idempotency records
  - Unclear which endpoints require idempotency
- Safe modification: Document idempotency requirements, implement cleanup job
- Test coverage: 0 idempotency tests

## Scaling Limits

**MongoDB Connection Pooling:**
- Current capacity: Default driver pools (unclear limits)
- Limit: Not documented; will become unavailable during admin migration
- Scaling path: Complete PostgreSQL migration to eliminate

**PostgreSQL Connection Pool:**
- Current capacity: Min 5, Max 25 connections configured in `bfs-backend/internal/config/config.go`
- Limit: Typical cloud databases allow 100-200 simultaneous connections; admin handler migration will increase load
- Scaling path: Monitor connection pool saturation; consider connection pooling service (PgBouncer) if scaling further

**Stripe API Rate Limits:**
- Current capacity: Stripe allows ~100 requests/second for Stripe accounts
- Limit: High-volume payment intents during peak hours could hit limits
- Scaling path: Implement exponential backoff retry; monitor API usage dashboard; consider batch operations

**Payrexx Gateway Limits:**
- Current capacity: Not documented
- Limit: Unknown; Payrexx API likely has undocumented rate limits
- Scaling path: Contact Payrexx for SLA; implement circuit breaker pattern for API calls

## Dependencies at Risk

**Stripe SDK v82:**
- Risk: Stripe only maintained actively; legacy versions unsupported. Version 82 may become stale as newer versions released
- Files: `bfs-backend/go.mod`, `bfs-backend/internal/service/payment.go`
- Impact: Security vulnerabilities in old SDK versions; API deprecation risk
- Current state: Used only for legacy code path; Payrexx replaces it
- Migration plan: Remove entirely in Phase 6 (post-admin handler migration)

**MongoDB Driver v2:**
- Risk: Large dependency with CVEs; version 2 EOL timeline unknown
- Files: `bfs-backend/go.mod`, 20 repository files, 73 domain types with BSON tags
- Impact: Transitive dependency vulnerabilities; upgrade pain
- Current state: Only used by admin handlers (deprecated)
- Migration plan: Remove entire MongoDB dependency in Phase 6

**Uber FX Dependency Injection:**
- Risk: FX is powerful but complex; few developers understand DI scopes and lifecycle
- Files: `bfs-backend/internal/app/` (services.go, handlers.go, router.go)
- Impact: Difficult to debug DI wiring; potential runtime initialization errors
- Current state: Well-established in codebase
- Recommendation: Add testing for DI graph; document FX patterns

## Missing Critical Features

**No Transaction Support for Complex Operations:**
- Problem: Admin menu operations (reordering slots, attaching items) lack transaction wrappers
- Files: `bfs-backend/internal/handler/admin_menu.go`, `admin_product.go`
- Blocks: Data consistency guarantees for multi-step operations
- Impact: High - partial failures can leave data in inconsistent state
- Fix approach: Implement transaction wrapper utility; use for all multi-operation handlers

**No Rate Limiting for Public Endpoints:**
- Problem: Public order endpoints (`GET /v1/orders/{id}`, `POST /v1/payments/webhook`) lack rate limiting
- Files: `bfs-backend/internal/http/router.go` (lines 139, 269)
- Blocks: Protection against brute force, DoS attacks
- Impact: High - webhook endpoint could be abused to spam external services
- Fix approach: Implement per-IP rate limiting middleware; use Redis for distributed rate limiting

**No Audit Log Retention Policy:**
- Problem: Audit records created but no purging strategy defined
- Files: `bfs-backend/internal/repository/audit.go`
- Blocks: Compliance requirements; database growth management
- Impact: Medium - audit table could grow unbounded
- Fix approach: Implement TTL/purging strategy; add configuration for retention period

**No CSV Export Pagination:**
- Problem: Admin order CSV export may load entire dataset into memory
- Files: `bfs-backend/internal/handler/admin_order.go` (line 190)
- Blocks: Exporting large datasets (100k+ orders)
- Impact: Medium - could cause OOM on large exports
- Fix approach: Implement streaming CSV export with chunked database queries

**Missing POS Device Provisioning Verification:**
- Problem: POS devices created from requests but no validation that device keys are cryptographically secure
- Files: `bfs-backend/internal/handler/pos_pg.go`, `pos.go`
- Blocks: Security assurance for device authenticity
- Impact: Medium - rogue devices could potentially access POS interface
- Fix approach: Add device certificate validation or HMAC-based device key verification

## Test Coverage Gaps

**Zero Backend Unit Tests:**
- What's not tested: All business logic in repositories, services, handlers
- Files: Only 2 test files exist (`internal/utils/hash_test.go`, `crypto_test.go`)
- Risk: High - Any refactoring to payment processing, order management, or admin operations could introduce bugs unnoticed
- Priority: High
- Recommendation: Target 70% coverage; start with payment and order services (highest risk)

**No Integration Tests for Payment Flow:**
- What's not tested: End-to-end payment creation → webhook → order completion
- Files: `bfs-backend/internal/service/payment_pg.go`, `internal/handler/payment_pg.go`
- Risk: High - Payment bugs directly impact revenue; unclear if Payrexx webhook handling works correctly
- Priority: High
- Recommendation: Add test harness for Payrexx API mocking; test idempotency

**No E2E Tests for Admin Operations:**
- What's not tested: Admin role checks, RBAC enforcement, data consistency after complex operations
- Files: All `admin_*.go` handlers
- Risk: High - Admin operations could be exploited; role changes could have inconsistent effects
- Priority: High
- Recommendation: Add E2E tests using test database; verify RBAC before/after migrations

**No Android Integration Tests:**
- What's not tested: WebView communication, SumUp payment callback integration, print function
- Files: `bfs-android-app/app/src/main/java/ch/leys/bless2n/PosWebActivity.kt`
- Risk: Medium - Payment flow broken only discovered by users
- Priority: Medium
- Recommendation: Add Espresso tests for WebView interaction

**PostgreSQL Migration Verification Tests:**
- What's not tested: Data consistency between MongoDB and PostgreSQL during migration
- Files: Entire `bfs-backend/internal/postgres/` directory
- Risk: High - Silent data corruption during migration
- Priority: Critical (must complete before going live)
- Recommendation: Write migration verification scripts comparing MongoDB → PostgreSQL records

---

*Concerns audit: 2026-01-28*
