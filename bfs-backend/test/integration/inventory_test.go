package integration

import (
	"context"
	"testing"

	"backend/internal/generated/ent/inventoryledger"
	"backend/internal/generated/ent/order"
	productEnum "backend/internal/generated/ent/product"
	pgRepo "backend/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestInventoryRepository_CRUD(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	ctx := context.Background()

	// Setup test product
	category := fixtures.CreateCategory("Drinks", 1, true)
	product := fixtures.CreateProduct("Cola", category.ID, 350, productEnum.TypeSimple, nil)

	t.Run("Create inventory entry", func(t *testing.T) {
		entry, err := repos.Inventory.Create(ctx, product.ID, 100, inventoryledger.ReasonOpeningBalance, nil, nil, nil, nil)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, entry.ID)
	})

	t.Run("GetCurrentStock calculates total", func(t *testing.T) {
		// Initial stock should be 100 from previous test
		stock, err := repos.Inventory.GetCurrentStock(ctx, product.ID)
		require.NoError(t, err)
		require.Equal(t, 100, stock)

		// Add more inventory
		fixtures.AddInventory(product.ID, 50, inventoryledger.ReasonManualAdjust)

		stock, err = repos.Inventory.GetCurrentStock(ctx, product.ID)
		require.NoError(t, err)
		require.Equal(t, 150, stock)

		// Subtract inventory (sale)
		fixtures.AddInventory(product.ID, -30, inventoryledger.ReasonSale)

		stock, err = repos.Inventory.GetCurrentStock(ctx, product.ID)
		require.NoError(t, err)
		require.Equal(t, 120, stock)
	})

	t.Run("GetCurrentStockBatch returns multiple products", func(t *testing.T) {
		tdb.Cleanup(t)

		category := fixtures.CreateCategory("Drinks", 1, true)
		cola := fixtures.CreateProduct("Cola", category.ID, 350, productEnum.TypeSimple, nil)
		sprite := fixtures.CreateProduct("Sprite", category.ID, 350, productEnum.TypeSimple, nil)
		fanta := fixtures.CreateProduct("Fanta", category.ID, 350, productEnum.TypeSimple, nil)

		fixtures.AddInventory(cola.ID, 100, inventoryledger.ReasonOpeningBalance)
		fixtures.AddInventory(sprite.ID, 50, inventoryledger.ReasonOpeningBalance)
		// Fanta has no inventory

		stocks, err := repos.Inventory.GetCurrentStockBatch(ctx, []uuid.UUID{cola.ID, sprite.ID, fanta.ID})
		require.NoError(t, err)
		require.Len(t, stocks, 3)
		require.Equal(t, 100, stocks[cola.ID])
		require.Equal(t, 50, stocks[sprite.ID])
		require.Equal(t, 0, stocks[fanta.ID]) // Should default to 0
	})

	t.Run("CreateMany creates multiple entries", func(t *testing.T) {
		tdb.Cleanup(t)

		category := fixtures.CreateCategory("Drinks", 1, true)
		cola := fixtures.CreateProduct("Cola", category.ID, 350, productEnum.TypeSimple, nil)
		sprite := fixtures.CreateProduct("Sprite", category.ID, 350, productEnum.TypeSimple, nil)

		entries := []pgRepo.InventoryLedgerCreateParams{
			{ProductID: cola.ID, Delta: 100, Reason: inventoryledger.ReasonOpeningBalance},
			{ProductID: sprite.ID, Delta: 75, Reason: inventoryledger.ReasonOpeningBalance},
		}

		_, err := repos.Inventory.CreateMany(ctx, entries)
		require.NoError(t, err)

		// Verify both were created
		stock1, err := repos.Inventory.GetCurrentStock(ctx, cola.ID)
		require.NoError(t, err)
		require.Equal(t, 100, stock1)

		stock2, err := repos.Inventory.GetCurrentStock(ctx, sprite.ID)
		require.NoError(t, err)
		require.Equal(t, 75, stock2)
	})

	t.Run("GetByProductID returns all entries for product", func(t *testing.T) {
		tdb.Cleanup(t)

		category := fixtures.CreateCategory("Drinks", 1, true)
		product := fixtures.CreateProduct("Cola", category.ID, 350, productEnum.TypeSimple, nil)

		fixtures.AddInventory(product.ID, 100, inventoryledger.ReasonOpeningBalance)
		fixtures.AddInventory(product.ID, 50, inventoryledger.ReasonManualAdjust)
		fixtures.AddInventory(product.ID, -10, inventoryledger.ReasonSale)

		entries, err := repos.Inventory.GetByProductID(ctx, product.ID)
		require.NoError(t, err)
		require.Len(t, entries, 3)
	})

	t.Run("SumByProductIDs returns totals for multiple products", func(t *testing.T) {
		tdb.Cleanup(t)

		category := fixtures.CreateCategory("Drinks", 1, true)
		cola := fixtures.CreateProduct("Cola", category.ID, 350, productEnum.TypeSimple, nil)
		sprite := fixtures.CreateProduct("Sprite", category.ID, 350, productEnum.TypeSimple, nil)

		fixtures.AddInventory(cola.ID, 100, inventoryledger.ReasonOpeningBalance)
		fixtures.AddInventory(cola.ID, -20, inventoryledger.ReasonSale)
		fixtures.AddInventory(sprite.ID, 50, inventoryledger.ReasonOpeningBalance)

		sums, err := repos.Inventory.SumByProductIDs(ctx, []uuid.UUID{cola.ID, sprite.ID})
		require.NoError(t, err)
		require.EqualValues(t, 80, sums[cola.ID])
		require.EqualValues(t, 50, sums[sprite.ID])
	})

	t.Run("Empty product list returns empty map", func(t *testing.T) {
		stocks, err := repos.Inventory.GetCurrentStockBatch(ctx, []uuid.UUID{})
		require.NoError(t, err)
		require.Empty(t, stocks)

		sums, err := repos.Inventory.SumByProductIDs(ctx, []uuid.UUID{})
		require.NoError(t, err)
		require.Empty(t, sums)
	})
}

func TestInventoryManagement_OrderFlow(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	ctx := context.Background()

	// Setup
	category := fixtures.CreateCategory("Drinks", 1, true)
	cola := fixtures.CreateProduct("Cola", category.ID, 350, productEnum.TypeSimple, nil)

	// Initial inventory
	fixtures.AddInventory(cola.ID, 100, inventoryledger.ReasonOpeningBalance)

	t.Run("Inventory decreases with sales", func(t *testing.T) {
		initialStock, err := repos.Inventory.GetCurrentStock(ctx, cola.ID)
		require.NoError(t, err)
		require.Equal(t, 100, initialStock)

		// Simulate sale
		order := fixtures.CreateOrder(700, order.StatusPaid, order.OriginShop)
		fixtures.AddInventory(cola.ID, -2, inventoryledger.ReasonSale)

		// Link to order (optional)
		_, err = repos.Inventory.Create(ctx, cola.ID, 0, inventoryledger.ReasonSale, &order.ID, nil, nil, nil)
		require.NoError(t, err)

		newStock, err := repos.Inventory.GetCurrentStock(ctx, cola.ID)
		require.NoError(t, err)
		require.Equal(t, 98, newStock)
	})

	t.Run("Inventory increases with refunds", func(t *testing.T) {
		beforeRefund, err := repos.Inventory.GetCurrentStock(ctx, cola.ID)
		require.NoError(t, err)

		// Simulate refund
		fixtures.AddInventory(cola.ID, 2, inventoryledger.ReasonRefund)

		afterRefund, err := repos.Inventory.GetCurrentStock(ctx, cola.ID)
		require.NoError(t, err)
		require.Equal(t, beforeRefund+2, afterRefund)
	})

	t.Run("Manual adjustments work correctly", func(t *testing.T) {
		beforeAdjust, err := repos.Inventory.GetCurrentStock(ctx, cola.ID)
		require.NoError(t, err)

		// Positive adjustment
		fixtures.AddInventory(cola.ID, 10, inventoryledger.ReasonManualAdjust)

		afterPositive, err := repos.Inventory.GetCurrentStock(ctx, cola.ID)
		require.NoError(t, err)
		require.Equal(t, beforeAdjust+10, afterPositive)

		// Negative adjustment (correction)
		fixtures.AddInventory(cola.ID, -5, inventoryledger.ReasonCorrection)

		afterNegative, err := repos.Inventory.GetCurrentStock(ctx, cola.ID)
		require.NoError(t, err)
		require.Equal(t, afterPositive-5, afterNegative)
	})
}

func TestInventoryManagement_LowStock(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	ctx := context.Background()

	category := fixtures.CreateCategory("Drinks", 1, true)
	cola := fixtures.CreateProduct("Cola", category.ID, 350, productEnum.TypeSimple, nil)

	t.Run("Stock can go to zero", func(t *testing.T) {
		fixtures.AddInventory(cola.ID, 10, inventoryledger.ReasonOpeningBalance)
		fixtures.AddInventory(cola.ID, -10, inventoryledger.ReasonSale)

		stock, err := repos.Inventory.GetCurrentStock(ctx, cola.ID)
		require.NoError(t, err)
		require.Equal(t, 0, stock)
	})

	t.Run("Stock can go negative (for tracking oversells)", func(t *testing.T) {
		tdb.Cleanup(t)

		category := fixtures.CreateCategory("Drinks", 1, true)
		sprite := fixtures.CreateProduct("Sprite", category.ID, 350, productEnum.TypeSimple, nil)

		fixtures.AddInventory(sprite.ID, 5, inventoryledger.ReasonOpeningBalance)
		fixtures.AddInventory(sprite.ID, -7, inventoryledger.ReasonSale) // Oversell

		stock, err := repos.Inventory.GetCurrentStock(ctx, sprite.ID)
		require.NoError(t, err)
		require.Equal(t, -2, stock)
	})
}
