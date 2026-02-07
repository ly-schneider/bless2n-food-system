package integration

import (
	"context"
	"testing"
	"time"

	entOrder "backend/internal/generated/ent/order"
	"backend/internal/generated/ent/orderline"
	entProduct "backend/internal/generated/ent/product"
	"backend/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOrderService_GetByID(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)

	svc := service.NewOrderService(repos.Order, repos.OrderLine, repos.Inventory, nil)
	ctx := context.Background()

	// Setup test data
	order := fixtures.CreateOrder(1500, entOrder.StatusPending, entOrder.OriginShop)

	t.Run("GetByID returns order", func(t *testing.T) {
		result, err := svc.GetByID(ctx, order.ID)
		require.NoError(t, err)
		require.Equal(t, order.ID, result.ID)
		require.Equal(t, int64(1500), result.TotalCents)
		require.Equal(t, entOrder.StatusPending, result.Status)
	})

	t.Run("GetByID returns error for non-existent order", func(t *testing.T) {
		_, err := svc.GetByID(ctx, uuid.Must(uuid.NewV7()))
		require.Error(t, err)
	})
}

func TestOrderService_GetOrderLines(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)

	svc := service.NewOrderService(repos.Order, repos.OrderLine, repos.Inventory, nil)
	ctx := context.Background()

	// Setup test data
	category := fixtures.CreateCategory("Drinks", 1, true)
	product := fixtures.CreateProduct("Cola", category.ID, 350, entProduct.TypeSimple, nil)
	order := fixtures.CreateOrder(700, entOrder.StatusPending, entOrder.OriginShop)
	fixtures.CreateOrderLine(order.ID, product.ID, "Cola", 2, 350, orderline.LineTypeSimple)

	t.Run("GetOrderLines returns lines", func(t *testing.T) {
		lines, err := svc.GetOrderLines(ctx, order.ID)
		require.NoError(t, err)
		require.Len(t, lines, 1)
		require.Equal(t, "Cola", lines[0].Title)
		require.Equal(t, 2, lines[0].Quantity)
	})

	t.Run("GetOrderLines returns empty for order without lines", func(t *testing.T) {
		emptyOrder := fixtures.CreateOrder(0, entOrder.StatusPending, entOrder.OriginShop)
		lines, err := svc.GetOrderLines(ctx, emptyOrder.ID)
		require.NoError(t, err)
		require.Empty(t, lines)
	})
}

func TestOrderService_ListByCustomerID(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)

	svc := service.NewOrderService(repos.Order, repos.OrderLine, repos.Inventory, nil)
	ctx := context.Background()

	// Setup test data
	customerID := uuid.Must(uuid.NewV7())
	fixtures.CreateOrderWithCustomer(1000, entOrder.StatusPaid, entOrder.OriginShop, customerID.String())
	fixtures.CreateOrderWithCustomer(1500, entOrder.StatusPending, entOrder.OriginShop, customerID.String())
	fixtures.CreateOrderWithCustomer(2000, entOrder.StatusPaid, entOrder.OriginShop, customerID.String())

	// Create an order for a different customer
	otherCustomer := uuid.Must(uuid.NewV7())
	fixtures.CreateOrderWithCustomer(500, entOrder.StatusPaid, entOrder.OriginShop, otherCustomer.String())

	t.Run("ListByCustomerID returns customer orders", func(t *testing.T) {
		orders, total, err := svc.ListByCustomerID(ctx, customerID.String(), 20, 0)
		require.NoError(t, err)
		require.Equal(t, int64(3), total)
		require.Len(t, orders, 3)
	})

	t.Run("ListByCustomerID with pagination", func(t *testing.T) {
		orders, total, err := svc.ListByCustomerID(ctx, customerID.String(), 2, 0)
		require.NoError(t, err)
		require.Equal(t, int64(3), total)
		require.Len(t, orders, 2)

		orders, total, err = svc.ListByCustomerID(ctx, customerID.String(), 2, 2)
		require.NoError(t, err)
		require.Equal(t, int64(3), total)
		require.Len(t, orders, 1)
	})

	t.Run("ListByCustomerID returns empty for customer without orders", func(t *testing.T) {
		orders, total, err := svc.ListByCustomerID(ctx, uuid.Must(uuid.NewV7()).String(), 20, 0)
		require.NoError(t, err)
		require.Equal(t, int64(0), total)
		require.Empty(t, orders)
	})

	t.Run("ListByCustomerID respects limit boundaries", func(t *testing.T) {
		orders, _, err := svc.ListByCustomerID(ctx, customerID.String(), 0, 0)
		require.NoError(t, err)
		require.True(t, len(orders) > 0)

		orders, _, err = svc.ListByCustomerID(ctx, customerID.String(), 200, 0)
		require.NoError(t, err)
		require.LessOrEqual(t, len(orders), 100)
	})
}

func TestOrderService_ListAdmin(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)

	svc := service.NewOrderService(repos.Order, repos.OrderLine, repos.Inventory, nil)
	ctx := context.Background()

	// Setup test data with different statuses
	fixtures.CreateOrder(1000, entOrder.StatusPaid, entOrder.OriginShop)
	fixtures.CreateOrder(1500, entOrder.StatusPending, entOrder.OriginPos)
	fixtures.CreateOrder(2000, entOrder.StatusCancelled, entOrder.OriginShop)

	t.Run("ListAdmin returns all orders", func(t *testing.T) {
		params := service.OrderListParams{Limit: 50, Offset: 0}
		orders, total, err := svc.ListAdmin(ctx, params)
		require.NoError(t, err)
		require.Equal(t, int64(3), total)
		require.Len(t, orders, 3)
	})

	t.Run("ListAdmin filters by status", func(t *testing.T) {
		status := entOrder.StatusPending
		params := service.OrderListParams{Status: &status, Limit: 50}
		orders, total, err := svc.ListAdmin(ctx, params)
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Len(t, orders, 1)
		require.Equal(t, entOrder.StatusPending, orders[0].Status)
	})

	t.Run("ListAdmin with pagination", func(t *testing.T) {
		params := service.OrderListParams{Limit: 2, Offset: 0}
		orders, total, err := svc.ListAdmin(ctx, params)
		require.NoError(t, err)
		require.Equal(t, int64(3), total)
		require.Len(t, orders, 2)

		params = service.OrderListParams{Limit: 2, Offset: 2}
		orders, _, err = svc.ListAdmin(ctx, params)
		require.NoError(t, err)
		require.Len(t, orders, 1)
	})

	t.Run("ListAdmin filters by date range", func(t *testing.T) {
		now := time.Now()
		from := now.Add(-1 * time.Hour).Format(time.RFC3339)
		to := now.Add(1 * time.Hour).Format(time.RFC3339)

		params := service.OrderListParams{From: &from, To: &to, Limit: 50}
		orders, total, err := svc.ListAdmin(ctx, params)
		require.NoError(t, err)
		require.Equal(t, int64(3), total)
		require.Len(t, orders, 3)

		// Test with past date range (no orders)
		pastFrom := now.Add(-48 * time.Hour).Format(time.RFC3339)
		pastTo := now.Add(-24 * time.Hour).Format(time.RFC3339)
		params = service.OrderListParams{From: &pastFrom, To: &pastTo, Limit: 50}
		orders, total, err = svc.ListAdmin(ctx, params)
		require.NoError(t, err)
		require.Equal(t, int64(0), total)
		require.Empty(t, orders)
	})
}

func TestOrderService_UpdateStatus(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)

	svc := service.NewOrderService(repos.Order, repos.OrderLine, repos.Inventory, nil)
	ctx := context.Background()

	t.Run("Valid status transitions", func(t *testing.T) {
		testCases := []struct {
			name    string
			from    entOrder.Status
			to      entOrder.Status
			wantErr bool
		}{
			{"pending to paid", entOrder.StatusPending, entOrder.StatusPaid, false},
			{"pending to cancelled", entOrder.StatusPending, entOrder.StatusCancelled, false},
			{"paid to refunded", entOrder.StatusPaid, entOrder.StatusRefunded, false},
			{"paid to cancelled", entOrder.StatusPaid, entOrder.StatusCancelled, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				tdb.Cleanup(t)
				ord := fixtures.CreateOrder(1000, tc.from, entOrder.OriginShop)

				err := svc.UpdateStatus(ctx, ord.ID, tc.to)
				if tc.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					updated, err := svc.GetByID(ctx, ord.ID)
					require.NoError(t, err)
					require.Equal(t, tc.to, updated.Status)
				}
			})
		}
	})

	t.Run("Invalid status transitions", func(t *testing.T) {
		testCases := []struct {
			name string
			from entOrder.Status
			to   entOrder.Status
		}{
			{"cancelled to paid", entOrder.StatusCancelled, entOrder.StatusPaid},
			{"refunded to pending", entOrder.StatusRefunded, entOrder.StatusPending},
			{"pending to refunded", entOrder.StatusPending, entOrder.StatusRefunded},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				tdb.Cleanup(t)
				ord := fixtures.CreateOrder(1000, tc.from, entOrder.OriginShop)

				err := svc.UpdateStatus(ctx, ord.ID, tc.to)
				require.Error(t, err)
				require.Contains(t, err.Error(), "invalid status transition")
			})
		}
	})

	t.Run("UpdateStatus returns error for non-existent order", func(t *testing.T) {
		err := svc.UpdateStatus(ctx, uuid.Must(uuid.NewV7()), entOrder.StatusPaid)
		require.Error(t, err)
	})
}
