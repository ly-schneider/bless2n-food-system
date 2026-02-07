# Integration Tests

This directory contains integration tests for the PostgreSQL-migrated backend services.

## Prerequisites

1. A running PostgreSQL instance (version 14+)
2. The `POSTGRES_TEST_DSN` environment variable set

## Running Tests

### Using Make

```bash
# From the bfs-backend directory
POSTGRES_TEST_DSN="postgres://user:pass@localhost:5432/testdb?sslmode=disable" make test-integration
```

### Using Go directly

```bash
cd bfs-backend
POSTGRES_TEST_DSN="postgres://user:pass@localhost:5432/testdb?sslmode=disable" go test -v ./test/integration/...
```

### Using Docker Compose

If you have the local development stack running:

```bash
# Start the database
make docker-up

# Run tests against the local PostgreSQL
POSTGRES_TEST_DSN="postgres://postgres:postgres@localhost:5432/bfs_test?sslmode=disable" make test-integration
```

## Test Structure

The integration tests are organized by service:

- `testdb.go` - Test infrastructure (database setup, fixtures, helpers)
- `category_service_test.go` - CategoryServicePG tests
- `product_service_test.go` - ProductServicePG tests
- `order_service_test.go` - OrderServicePG tests
- `payment_service_test.go` - PaymentServicePG tests (with mocked Payrexx)
- `pos_service_test.go` - POSServicePG tests
- `station_service_test.go` - StationServicePG tests
- `pos_config_service_test.go` - POSConfigServicePG tests
- `inventory_test.go` - Inventory management tests

## Test Database Setup

Tests automatically:
1. Create the `app` schema if it doesn't exist
2. Run Ent auto-migration (creates tables, enums, indexes)
4. Truncate tables between test runs

Each test file uses `tdb.Cleanup(t)` to ensure a clean state.

## Key Testing Patterns

### Table-Driven Tests

```go
testCases := []struct {
    name    string
    from    model.OrderStatus
    to      model.OrderStatus
    wantErr bool
}{
    {"pending to paid", model.OrderStatusPending, model.OrderStatusPaid, false},
    // ...
}

for _, tc := range testCases {
    t.Run(tc.name, func(t *testing.T) {
        // test logic
    })
}
```

### Fixtures

Use the `Fixtures` helper for test data setup:

```go
fixtures := NewFixtures(repos)
category := fixtures.CreateCategory("Drinks", 1, true)
product := fixtures.CreateProduct("Cola", category.ID, 350, model.ProductTypeSimple, nil)
fixtures.AddInventory(product.ID, 100, model.InventoryReasonOpeningBalance)
```

### Assertions

Tests use testify/require for assertions:

```go
require.NoError(t, err)
require.Equal(t, expected, actual)
require.Len(t, items, 3)
require.Contains(t, err.Error(), "expected_substring")
```

## Coverage

These tests cover:

- Category CRUD operations and pagination
- Product listing with filters, menu products with slots/options
- Order creation, status transitions, customer/admin listing
- POS device registration and payment methods (cash, card, TWINT)
- Station verification, QR code generation/validation, item redemption
- Payment preparation, order creation, Payrexx integration (mocked)
- Jeton management and POS fulfillment mode configuration
- Inventory ledger operations and stock calculations

## Notes

- Tests skip automatically if `POSTGRES_TEST_DSN` is not set
- The test database is shared across test files; use `tdb.Cleanup(t)` for isolation
- External services (Payrexx) are not called; those paths require integration with actual APIs
- Tests run with `-race` flag by default for race condition detection
