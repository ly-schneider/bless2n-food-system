package integration

import (
	"context"
	"testing"

	"backend/internal/generated/ent/inventoryledger"
	"backend/internal/generated/ent/order"
	"backend/internal/generated/ent/orderline"
	"backend/internal/generated/ent/product"
	"backend/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPaymentService_PrepareAndCreateOrder(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	cfg := TestConfig()

	svc := service.NewPaymentService(
		cfg,
		repos.Order,
		repos.OrderLine,
		repos.Product,
		repos.MenuSlot,
		repos.Inventory,
		nil,
	)
	ctx := context.Background()

	// Setup test products
	category := fixtures.CreateCategory("Drinks", 1, true)
	cola := fixtures.CreateProduct("Cola", category.ID, 350, product.TypeSimple, nil)
	sprite := fixtures.CreateProduct("Sprite", category.ID, 350, product.TypeSimple, nil)

	// Add inventory
	fixtures.AddInventory(cola.ID, 100, inventoryledger.ReasonOpeningBalance)
	fixtures.AddInventory(sprite.ID, 100, inventoryledger.ReasonOpeningBalance)

	t.Run("PrepareAndCreateOrder creates order with simple products", func(t *testing.T) {
		input := service.CreateCheckoutInput{
			Items: []service.CheckoutItemInput{
				{ProductID: cola.ID.String(), Quantity: 2},
				{ProductID: sprite.ID.String(), Quantity: 1},
			},
		}

		prep, err := svc.PrepareAndCreateOrder(ctx, input, nil, nil)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, prep.OrderID)
		require.Equal(t, int64(1050), prep.TotalCents) // 2*350 + 1*350
		require.Len(t, prep.LineItems, 2)

		// Verify order was created
		orderObj, err := repos.Order.GetByID(ctx, prep.OrderID)
		require.NoError(t, err)
		require.Equal(t, order.StatusPending, orderObj.Status)
		require.Equal(t, order.OriginShop, orderObj.Origin)
	})

	t.Run("PrepareAndCreateOrder with user ID", func(t *testing.T) {
		userID := uuid.Must(uuid.NewV7()).String()
		input := service.CreateCheckoutInput{
			Items: []service.CheckoutItemInput{
				{ProductID: cola.ID.String(), Quantity: 1},
			},
		}

		prep, err := svc.PrepareAndCreateOrder(ctx, input, &userID, nil)
		require.NoError(t, err)
		require.Equal(t, &userID, prep.UserID)

		orderObj, err := repos.Order.GetByID(ctx, prep.OrderID)
		require.NoError(t, err)
		require.NotNil(t, orderObj.CustomerID)
		require.Equal(t, userID, *orderObj.CustomerID)
	})

	t.Run("PrepareAndCreateOrder with attempt ID", func(t *testing.T) {
		attemptID := "test-attempt-123"
		input := service.CreateCheckoutInput{
			Items: []service.CheckoutItemInput{
				{ProductID: cola.ID.String(), Quantity: 1},
			},
		}

		prep, err := svc.PrepareAndCreateOrder(ctx, input, nil, &attemptID)
		require.NoError(t, err)

		orderObj, err := repos.Order.GetByID(ctx, prep.OrderID)
		require.NoError(t, err)
		require.NotNil(t, orderObj.PaymentAttemptID)
		require.Equal(t, attemptID, *orderObj.PaymentAttemptID)
	})

	t.Run("PrepareAndCreateOrder reserves inventory", func(t *testing.T) {
		// Check initial stock
		initialStock, err := repos.Inventory.GetCurrentStock(ctx, cola.ID)
		require.NoError(t, err)

		input := service.CreateCheckoutInput{
			Items: []service.CheckoutItemInput{
				{ProductID: cola.ID.String(), Quantity: 5},
			},
		}

		_, err = svc.PrepareAndCreateOrder(ctx, input, nil, nil)
		require.NoError(t, err)

		// Check stock was reduced
		newStock, err := repos.Inventory.GetCurrentStock(ctx, cola.ID)
		require.NoError(t, err)
		require.Equal(t, initialStock-5, newStock)
	})

	t.Run("PrepareAndCreateOrder with no items fails", func(t *testing.T) {
		input := service.CreateCheckoutInput{Items: []service.CheckoutItemInput{}}

		_, err := svc.PrepareAndCreateOrder(ctx, input, nil, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no items")
	})

	t.Run("PrepareAndCreateOrder with invalid product ID fails", func(t *testing.T) {
		input := service.CreateCheckoutInput{
			Items: []service.CheckoutItemInput{
				{ProductID: "invalid-uuid", Quantity: 1},
			},
		}

		_, err := svc.PrepareAndCreateOrder(ctx, input, nil, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid productId")
	})

	t.Run("PrepareAndCreateOrder with non-existent product fails", func(t *testing.T) {
		input := service.CreateCheckoutInput{
			Items: []service.CheckoutItemInput{
				{ProductID: uuid.Must(uuid.NewV7()).String(), Quantity: 1},
			},
		}

		_, err := svc.PrepareAndCreateOrder(ctx, input, nil, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unknown product")
	})
}

func TestPaymentService_MenuCheckout(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	cfg := TestConfig()

	svc := service.NewPaymentService(
		cfg,
		repos.Order,
		repos.OrderLine,
		repos.Product,
		repos.MenuSlot,
		repos.Inventory,
		nil,
	)
	ctx := context.Background()

	// Setup menu product
	category := fixtures.CreateCategory("Menus", 1, true)
	optionCat := fixtures.CreateCategory("Options", 2, true)

	cola := fixtures.CreateProduct("Cola", optionCat.ID, 350, product.TypeSimple, nil)
	fries := fixtures.CreateProduct("Fries", optionCat.ID, 400, product.TypeSimple, nil)

	fixtures.AddInventory(cola.ID, 50, inventoryledger.ReasonOpeningBalance)
	fixtures.AddInventory(fries.ID, 50, inventoryledger.ReasonOpeningBalance)

	menu := fixtures.CreateProduct("Combo Menu", category.ID, 1500, product.TypeMenu, nil)

	drinkSlot := fixtures.CreateMenuSlot(menu.ID, "Drink", 1)
	sideSlot := fixtures.CreateMenuSlot(menu.ID, "Side", 2)

	fixtures.CreateMenuSlotOption(drinkSlot.ID, cola.ID)
	fixtures.CreateMenuSlotOption(sideSlot.ID, fries.ID)

	t.Run("PrepareAndCreateOrder with menu configuration", func(t *testing.T) {
		input := service.CreateCheckoutInput{
			Items: []service.CheckoutItemInput{
				{
					ProductID: menu.ID.String(),
					Quantity:  1,
					Configuration: map[string]string{
						drinkSlot.ID.String(): cola.ID.String(),
						sideSlot.ID.String():  fries.ID.String(),
					},
				},
			},
		}

		prep, err := svc.PrepareAndCreateOrder(ctx, input, nil, nil)
		require.NoError(t, err)
		require.Equal(t, int64(1500), prep.TotalCents) // Menu price only

		// Verify order lines were created
		lines, err := repos.OrderLine.GetByOrderID(ctx, prep.OrderID)
		require.NoError(t, err)

		// Should have 3 lines: menu parent + 2 components
		require.Len(t, lines, 3)

		var bundleCount, componentCount int
		for _, line := range lines {
			if line.LineType == orderline.LineTypeBundle {
				bundleCount++
				require.Equal(t, "Combo Menu", line.Title)
			}
			if line.LineType == orderline.LineTypeComponent {
				componentCount++
				require.NotNil(t, line.ParentLineID)
			}
		}
		require.Equal(t, 1, bundleCount)
		require.Equal(t, 2, componentCount)
	})

	t.Run("PrepareAndCreateOrder with invalid slot configuration fails", func(t *testing.T) {
		invalidSlotID := uuid.Must(uuid.NewV7())
		input := service.CreateCheckoutInput{
			Items: []service.CheckoutItemInput{
				{
					ProductID: menu.ID.String(),
					Quantity:  1,
					Configuration: map[string]string{
						invalidSlotID.String(): cola.ID.String(),
					},
				},
			},
		}

		_, err := svc.PrepareAndCreateOrder(ctx, input, nil, nil)
		require.Error(t, err)
	})

	t.Run("PrepareAndCreateOrder with invalid option in slot fails", func(t *testing.T) {
		// Try to put fries in the drink slot (not allowed)
		input := service.CreateCheckoutInput{
			Items: []service.CheckoutItemInput{
				{
					ProductID: menu.ID.String(),
					Quantity:  1,
					Configuration: map[string]string{
						drinkSlot.ID.String(): fries.ID.String(), // Fries not allowed in drink slot
					},
				},
			},
		}

		_, err := svc.PrepareAndCreateOrder(ctx, input, nil, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "product not allowed")
	})
}

func TestPaymentService_FindPendingOrderByAttemptID(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	cfg := TestConfig()

	svc := service.NewPaymentService(
		cfg,
		repos.Order,
		repos.OrderLine,
		repos.Product,
		repos.MenuSlot,
		repos.Inventory,
		nil,
	)
	ctx := context.Background()

	// Setup
	category := fixtures.CreateCategory("Drinks", 1, true)
	cola := fixtures.CreateProduct("Cola", category.ID, 350, product.TypeSimple, nil)
	fixtures.AddInventory(cola.ID, 100, inventoryledger.ReasonOpeningBalance)

	t.Run("FindPendingOrderByAttemptID finds pending order", func(t *testing.T) {
		attemptID := "attempt-123"
		input := service.CreateCheckoutInput{
			Items: []service.CheckoutItemInput{
				{ProductID: cola.ID.String(), Quantity: 1},
			},
		}

		prep, err := svc.PrepareAndCreateOrder(ctx, input, nil, &attemptID)
		require.NoError(t, err)

		found, err := svc.FindPendingOrderByAttemptID(ctx, attemptID)
		require.NoError(t, err)
		require.Equal(t, prep.OrderID, found.ID)
	})

	t.Run("FindPendingOrderByAttemptID returns error for non-existent", func(t *testing.T) {
		_, err := svc.FindPendingOrderByAttemptID(ctx, "non-existent-attempt")
		require.Error(t, err)
	})

	t.Run("FindPendingOrderByAttemptID returns error for empty ID", func(t *testing.T) {
		_, err := svc.FindPendingOrderByAttemptID(ctx, "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing attempt id")
	})
}

func TestPaymentService_MarkOrderPaidByPayrexx(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	cfg := TestConfig()

	svc := service.NewPaymentService(
		cfg,
		repos.Order,
		repos.OrderLine,
		repos.Product,
		repos.MenuSlot,
		repos.Inventory,
		nil,
	)
	ctx := context.Background()

	t.Run("MarkOrderPaidByPayrexx updates order status", func(t *testing.T) {
		orderObj := fixtures.CreateOrder(1000, order.StatusPending, order.OriginShop)
		gatewayID := 12345
		transactionID := 67890
		email := "customer@test.com"

		err := svc.MarkOrderPaidByPayrexx(ctx, orderObj.ID, gatewayID, transactionID, &email)
		require.NoError(t, err)

		updated, err := repos.Order.GetByID(ctx, orderObj.ID)
		require.NoError(t, err)
		require.Equal(t, order.StatusPaid, updated.Status)
		require.NotNil(t, updated.PayrexxGatewayID)
		require.Equal(t, gatewayID, *updated.PayrexxGatewayID)
		require.NotNil(t, updated.PayrexxTransactionID)
		require.Equal(t, transactionID, *updated.PayrexxTransactionID)
		require.NotNil(t, updated.ContactEmail)
		require.Equal(t, email, *updated.ContactEmail)
	})

	t.Run("MarkOrderPaidByPayrexx returns error for non-existent order", func(t *testing.T) {
		err := svc.MarkOrderPaidByPayrexx(ctx, uuid.Must(uuid.NewV7()), 123, 456, nil)
		require.Error(t, err)
	})
}

func TestPaymentService_CleanupPendingOrderByID(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	cfg := TestConfig()

	svc := service.NewPaymentService(
		cfg,
		repos.Order,
		repos.OrderLine,
		repos.Product,
		repos.MenuSlot,
		repos.Inventory,
		nil,
	)
	ctx := context.Background()

	// Setup
	category := fixtures.CreateCategory("Drinks", 1, true)
	cola := fixtures.CreateProduct("Cola", category.ID, 350, product.TypeSimple, nil)
	fixtures.AddInventory(cola.ID, 100, inventoryledger.ReasonOpeningBalance)

	t.Run("CleanupPendingOrderByID deletes pending order and releases inventory", func(t *testing.T) {
		input := service.CreateCheckoutInput{
			Items: []service.CheckoutItemInput{
				{ProductID: cola.ID.String(), Quantity: 5},
			},
		}

		prep, err := svc.PrepareAndCreateOrder(ctx, input, nil, nil)
		require.NoError(t, err)

		// Check inventory was reserved
		stockBefore, err := repos.Inventory.GetCurrentStock(ctx, cola.ID)
		require.NoError(t, err)

		// Cleanup the order
		err = svc.CleanupPendingOrderByID(ctx, prep.OrderID)
		require.NoError(t, err)

		// Order should be deleted
		_, err = repos.Order.GetByID(ctx, prep.OrderID)
		require.Error(t, err)

		// Inventory should be released
		stockAfter, err := repos.Inventory.GetCurrentStock(ctx, cola.ID)
		require.NoError(t, err)
		require.Equal(t, stockBefore+5, stockAfter)
	})

	t.Run("CleanupPendingOrderByID does nothing for paid order", func(t *testing.T) {
		orderObj := fixtures.CreateOrder(1000, order.StatusPaid, order.OriginShop)

		err := svc.CleanupPendingOrderByID(ctx, orderObj.ID)
		require.NoError(t, err)

		// Order should still exist
		found, err := repos.Order.GetByID(ctx, orderObj.ID)
		require.NoError(t, err)
		require.Equal(t, order.StatusPaid, found.Status)
	})
}

func TestPaymentService_TWINTLimits(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	cfg := TestConfig()

	svc := service.NewPaymentService(
		cfg,
		repos.Order,
		repos.OrderLine,
		repos.Product,
		repos.MenuSlot,
		repos.Inventory,
		nil,
	)
	ctx := context.Background()

	// Setup expensive product
	category := fixtures.CreateCategory("Expensive", 1, true)
	expensive := fixtures.CreateProduct("Expensive Item", category.ID, 600000, product.TypeSimple, nil) // 6000 CHF
	fixtures.AddInventory(expensive.ID, 100, inventoryledger.ReasonOpeningBalance)

	t.Run("PrepareAndCreateOrder fails for single item exceeding TWINT limit", func(t *testing.T) {
		input := service.CreateCheckoutInput{
			Items: []service.CheckoutItemInput{
				{ProductID: expensive.ID.String(), Quantity: 1},
			},
		}

		_, err := svc.PrepareAndCreateOrder(ctx, input, nil, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "TWINT max")
	})

	t.Run("PrepareAndCreateOrder fails for total exceeding TWINT limit", func(t *testing.T) {
		affordable := fixtures.CreateProduct("Affordable", category.ID, 100000, product.TypeSimple, nil) // 1000 CHF
		fixtures.AddInventory(affordable.ID, 100, inventoryledger.ReasonOpeningBalance)

		input := service.CreateCheckoutInput{
			Items: []service.CheckoutItemInput{
				{ProductID: affordable.ID.String(), Quantity: 6}, // 6000 CHF total
			},
		}

		_, err := svc.PrepareAndCreateOrder(ctx, input, nil, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "TWINT max")
	})
}
