// Package integration provides integration test infrastructure.
package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"backend/internal/config"
	"backend/internal/generated/ent"
	"backend/internal/generated/ent/device"
	"backend/internal/generated/ent/inventoryledger"
	"backend/internal/generated/ent/order"
	"backend/internal/generated/ent/orderline"
	"backend/internal/generated/ent/product"
	pgRepo "backend/internal/repository"
	"backend/internal/service"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// TestDB wraps an Ent client and raw sql.DB for integration tests.
type TestDB struct {
	Client *ent.Client
	DB     *sql.DB
}

// NewTestDB creates a new test database connection using Ent.
// It expects POSTGRES_TEST_DSN environment variable to be set.
// If not set, tests will be skipped.
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()

	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_TEST_DSN not set; skipping integration tests")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		t.Fatalf("failed to ping test database: %v", err)
	}

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))

	tdb := &TestDB{Client: client, DB: db}

	// Create the app schema and run Ent auto-migration
	tdb.setupSchema(t)

	return tdb
}

func (tdb *TestDB) setupSchema(t *testing.T) {
	t.Helper()

	if err := tdb.Client.Schema.Create(context.Background()); err != nil {
		t.Fatalf("failed to run Ent schema migration: %v", err)
	}
}

// Cleanup truncates all tables for a fresh test state.
func (tdb *TestDB) Cleanup(t *testing.T) {
	t.Helper()

	// Tables ordered to respect foreign key constraints
	tables := []string{
		"idempotency",
		"order_line_redemption",
		"inventory_ledger",
		"order_payment",
		"order_line",
		"\"order\"",
		"menu_slot_option",
		"menu_slot",
		"device_product",
		"device_binding",
		"device",
		"club100_free_product",
		"product",
		"jeton",
		"category",
		"settings",
		"session",
		"admin_invite",
		"verification",
		"\"user\"",
	}

	ctx := context.Background()
	for _, table := range tables {
		if _, err := tdb.DB.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			// Ignore errors for non-existent tables
			if !strings.Contains(err.Error(), "does not exist") {
				t.Logf("warning: failed to truncate %s: %v", table, err)
			}
		}
	}
}

// Close closes the database connection.
func (tdb *TestDB) Close() {
	tdb.Client.Close()
	tdb.DB.Close()
}

// Repositories holds all repository instances for tests.
type Repositories struct {
	Category           pgRepo.CategoryRepository
	Product            *pgRepo.ProductRepository
	Jeton              pgRepo.JetonRepository
	MenuSlot           pgRepo.MenuSlotRepository
	MenuSlotOption     pgRepo.MenuSlotOptionRepository
	Order              pgRepo.OrderRepository
	OrderLine          pgRepo.OrderLineRepository
	OrderRedemption    pgRepo.OrderLineRedemptionRepository
	Club100Redemption  pgRepo.Club100RedemptionRepository
	Inventory          pgRepo.InventoryLedgerRepository
	Device             pgRepo.DeviceRepository
	DeviceProduct      pgRepo.DeviceProductRepository
	Settings           pgRepo.SettingsRepository
	Idempotency        pgRepo.IdempotencyRepository
}

// NewRepositories creates all repository instances from an Ent client.
func NewRepositories(client *ent.Client) *Repositories {
	return &Repositories{
		Category:          pgRepo.NewCategoryRepository(client),
		Product:           pgRepo.NewProductRepository(client),
		Jeton:             pgRepo.NewJetonRepository(client),
		MenuSlot:          pgRepo.NewMenuSlotRepository(client),
		MenuSlotOption:    pgRepo.NewMenuSlotOptionRepository(client),
		Order:             pgRepo.NewOrderRepository(client),
		OrderLine:         pgRepo.NewOrderLineRepository(client),
		OrderRedemption:   pgRepo.NewOrderLineRedemptionRepository(client),
		Club100Redemption: pgRepo.NewClub100RedemptionRepository(client),
		Inventory:         pgRepo.NewInventoryLedgerRepository(client),
		Device:            pgRepo.NewDeviceRepository(client),
		DeviceProduct:     pgRepo.NewDeviceProductRepository(client),
		Settings:          pgRepo.NewSettingsRepository(client),
		Idempotency:       pgRepo.NewIdempotencyRepository(client),
	}
}

// Fixtures provides helpers to create test data.
type Fixtures struct {
	repos *Repositories
	ctx   context.Context
}

// NewFixtures creates a new Fixtures helper.
func NewFixtures(repos *Repositories) *Fixtures {
	return &Fixtures{repos: repos, ctx: context.Background()}
}

// CreateCategory creates a test category.
func (f *Fixtures) CreateCategory(name string, position int, isActive bool) *ent.Category {
	cat, err := f.repos.Category.Create(f.ctx, name, position, isActive)
	if err != nil {
		panic(fmt.Sprintf("failed to create category: %v", err))
	}
	return cat
}

// CreateJeton creates a test jeton.
func (f *Fixtures) CreateJeton(name, color string) *ent.Jeton {
	jeton, err := f.repos.Jeton.Create(f.ctx, name, color)
	if err != nil {
		panic(fmt.Sprintf("failed to create jeton: %v", err))
	}
	return jeton
}

// CreateProduct creates a test product.
func (f *Fixtures) CreateProduct(name string, categoryID uuid.UUID, priceCents int64, productType product.Type, jetonID *uuid.UUID) *ent.Product {
	p, err := f.repos.Product.Create(f.ctx, categoryID, productType, name, priceCents, true, nil, jetonID)
	if err != nil {
		panic(fmt.Sprintf("failed to create product: %v", err))
	}
	return p
}

// CreateOrder creates a test order.
func (f *Fixtures) CreateOrder(totalCents int64, status order.Status, origin order.Origin) *ent.Order {
	ord, err := f.repos.Order.Create(f.ctx, totalCents, status, origin, nil, nil, nil, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create order: %v", err))
	}
	return ord
}

// CreateOrderWithCustomer creates an order with a customer ID.
func (f *Fixtures) CreateOrderWithCustomer(totalCents int64, status order.Status, origin order.Origin, customerID string) *ent.Order {
	ord, err := f.repos.Order.Create(f.ctx, totalCents, status, origin, &customerID, nil, nil, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create order: %v", err))
	}
	return ord
}

// CreateOrderLine creates a test order line.
func (f *Fixtures) CreateOrderLine(orderID, productID uuid.UUID, title string, quantity int, priceCents int64, lineType orderline.LineType) *ent.OrderLine {
	line, err := f.repos.OrderLine.Create(f.ctx, orderID, lineType, productID, title, quantity, priceCents, nil, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create order line: %v", err))
	}
	return line
}

// CreateDevice creates a test device.
func (f *Fixtures) CreateDevice(name, deviceKey string, deviceType device.Type, status device.Status) *ent.Device {
	d, err := f.repos.Device.Create(f.ctx, name, deviceKey, deviceType, status, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create device: %v", err))
	}
	return d
}

// AddInventory adds inventory for a product.
func (f *Fixtures) AddInventory(productID uuid.UUID, delta int, reason inventoryledger.Reason) *ent.InventoryLedger {
	entry, err := f.repos.Inventory.Create(f.ctx, productID, delta, reason, nil, nil, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create inventory entry: %v", err))
	}
	return entry
}

// CreateMenuSlot creates a test menu slot.
func (f *Fixtures) CreateMenuSlot(menuProductID uuid.UUID, name string, sequence int) *ent.MenuSlot {
	slot, err := f.repos.MenuSlot.Create(f.ctx, menuProductID, name, sequence)
	if err != nil {
		panic(fmt.Sprintf("failed to create menu slot: %v", err))
	}
	return slot
}

// CreateMenuSlotOption creates a test menu slot option.
func (f *Fixtures) CreateMenuSlotOption(slotID, optionProductID uuid.UUID) *ent.MenuSlotOption {
	opt, err := f.repos.MenuSlotOption.Create(f.ctx, slotID, optionProductID)
	if err != nil {
		panic(fmt.Sprintf("failed to create menu slot option: %v", err))
	}
	return opt
}

// AssignProductToDevice assigns a product to a device (station).
func (f *Fixtures) AssignProductToDevice(deviceID, productID uuid.UUID) {
	if _, err := f.repos.DeviceProduct.Create(f.ctx, deviceID, productID); err != nil {
		panic(fmt.Sprintf("failed to assign product to device: %v", err))
	}
}

// TestConfig returns a minimal config for testing.
func TestConfig() config.Config {
	return config.Config{
		App: config.AppConfig{
			AppEnv:        "test",
			PublicBaseURL: "http://localhost:3000",
		},
		Payrexx: config.PayrexxConfig{
			InstanceName: "",
			APISecret:    "",
		},
	}
}

// TimeWithin checks if two times are within a duration of each other.
func TimeWithin(a, b time.Time, d time.Duration) bool {
	diff := a.Sub(b)
	if diff < 0 {
		diff = -diff
	}
	return diff <= d
}

// MockElvantoService is a no-op ElvantoService for testing.
type MockElvantoService struct{}

func (m *MockElvantoService) SearchPeople(_ context.Context) ([]service.ElvantoPerson, error) {
	return nil, nil
}

func (m *MockElvantoService) IsConfigured() bool {
	return false
}
