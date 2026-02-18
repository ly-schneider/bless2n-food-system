package integration

import (
	"context"
	"testing"

	"backend/internal/service"

	entDevice "backend/internal/generated/ent/device"
	entInventoryLedger "backend/internal/generated/ent/inventoryledger"
	entOrder "backend/internal/generated/ent/order"
	entProduct "backend/internal/generated/ent/product"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPOSService_GetDeviceByToken(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	cfg := TestConfig()

	paymentSvc := service.NewPaymentService(
		cfg,
		repos.Order,
		repos.OrderLine,
		repos.Product,
		repos.MenuSlot,
		repos.Inventory,
		nil,
	)

	club100Svc := service.NewClub100Service(&MockElvantoService{}, repos.Club100Redemption, repos.Settings, repos.OrderLine)

	svc := service.NewPOSService(cfg, repos.Device, repos.Order, paymentSvc, club100Svc)
	ctx := context.Background()

	t.Run("GetDeviceByToken returns POS device", func(t *testing.T) {
		fixtures.CreateDevice("POS 1", "pos-token-1", entDevice.TypePOS, entDevice.StatusApproved)

		device, err := svc.GetDeviceByToken(ctx, "pos-token-1")
		require.NoError(t, err)
		require.Equal(t, "POS 1", device.Name)
		require.Equal(t, entDevice.TypePOS, device.Type)
	})

	t.Run("GetDeviceByToken returns error for station device", func(t *testing.T) {
		fixtures.CreateDevice("Station 1", "station-token-1", entDevice.TypeSTATION, entDevice.StatusApproved)

		_, err := svc.GetDeviceByToken(ctx, "station-token-1")
		require.Error(t, err)
		require.Equal(t, "device_not_pos", err.Error())
	})

	t.Run("GetDeviceByToken returns error for non-existent token", func(t *testing.T) {
		_, err := svc.GetDeviceByToken(ctx, "non-existent-token")
		require.Error(t, err)
	})
}

func TestPOSService_CreateOrder(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	cfg := TestConfig()

	paymentSvc := service.NewPaymentService(
		cfg,
		repos.Order,
		repos.OrderLine,
		repos.Product,
		repos.MenuSlot,
		repos.Inventory,
		nil,
	)

	club100Svc := service.NewClub100Service(&MockElvantoService{}, repos.Club100Redemption, repos.Settings, repos.OrderLine)

	svc := service.NewPOSService(cfg, repos.Device, repos.Order, paymentSvc, club100Svc)
	ctx := context.Background()

	// Setup test products
	category := fixtures.CreateCategory("Drinks", 1, true)
	cola := fixtures.CreateProduct("Cola", category.ID, 350, entProduct.TypeSimple, nil)
	sprite := fixtures.CreateProduct("Sprite", category.ID, 350, entProduct.TypeSimple, nil)

	// Add inventory
	fixtures.AddInventory(cola.ID, 100, entInventoryLedger.ReasonOpeningBalance)
	fixtures.AddInventory(sprite.ID, 100, entInventoryLedger.ReasonOpeningBalance)

	t.Run("CreateOrder creates order with items", func(t *testing.T) {
		items := []service.POSCheckoutItem{
			{ProductID: cola.ID.String(), Quantity: 2},
			{ProductID: sprite.ID.String(), Quantity: 1},
		}

		orderID, err := svc.CreateOrder(ctx, items, nil)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, orderID)

		// Verify order was created
		order, err := repos.Order.GetByID(ctx, orderID)
		require.NoError(t, err)
		require.Equal(t, int64(1050), order.TotalCents) // 2*350 + 1*350
		require.Equal(t, entOrder.StatusPending, order.Status)
		require.Equal(t, entOrder.OriginPos, order.Origin)
	})

	t.Run("CreateOrder with customer email", func(t *testing.T) {
		items := []service.POSCheckoutItem{
			{ProductID: cola.ID.String(), Quantity: 1},
		}
		email := "customer@test.com"

		orderID, err := svc.CreateOrder(ctx, items, &email)
		require.NoError(t, err)

		order, err := repos.Order.GetByID(ctx, orderID)
		require.NoError(t, err)
		require.NotNil(t, order.ContactEmail)
		require.Equal(t, email, *order.ContactEmail)
	})

	t.Run("CreateOrder with no items fails", func(t *testing.T) {
		_, err := svc.CreateOrder(ctx, []service.POSCheckoutItem{}, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no items")
	})

	t.Run("CreateOrder with invalid product ID fails", func(t *testing.T) {
		items := []service.POSCheckoutItem{
			{ProductID: "invalid-uuid", Quantity: 1},
		}

		_, err := svc.CreateOrder(ctx, items, nil)
		require.Error(t, err)
	})
}

func TestPOSService_PayCash(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	cfg := TestConfig()

	paymentSvc := service.NewPaymentService(
		cfg,
		repos.Order,
		repos.OrderLine,
		repos.Product,
		repos.MenuSlot,
		repos.Inventory,
		nil,
	)

	club100Svc := service.NewClub100Service(&MockElvantoService{}, repos.Club100Redemption, repos.Settings, repos.OrderLine)

	svc := service.NewPOSService(cfg, repos.Device, repos.Order, paymentSvc, club100Svc)
	ctx := context.Background()

	// Create a POS device
	device := fixtures.CreateDevice("POS 1", "pos-token", entDevice.TypePOS, entDevice.StatusApproved)

	t.Run("PayCash processes payment", func(t *testing.T) {
		order := fixtures.CreateOrder(1000, entOrder.StatusPending, entOrder.OriginPos)

		err := svc.PayCash(ctx, order.ID, &device.ID)
		require.NoError(t, err)

		// Verify order status
		updated, err := repos.Order.GetByID(ctx, order.ID)
		require.NoError(t, err)
		require.Equal(t, entOrder.StatusPaid, updated.Status)
	})

	t.Run("PayCash fails for non-pending order", func(t *testing.T) {
		order := fixtures.CreateOrder(1000, entOrder.StatusPaid, entOrder.OriginPos)

		err := svc.PayCash(ctx, order.ID, &device.ID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not_pending")
	})

	t.Run("PayCash fails with invalid order ID", func(t *testing.T) {
		err := svc.PayCash(ctx, uuid.Nil, &device.ID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid order id")
	})
}

func TestPOSService_PayCard(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	cfg := TestConfig()

	paymentSvc := service.NewPaymentService(
		cfg,
		repos.Order,
		repos.OrderLine,
		repos.Product,
		repos.MenuSlot,
		repos.Inventory,
		nil,
	)

	club100Svc := service.NewClub100Service(&MockElvantoService{}, repos.Club100Redemption, repos.Settings, repos.OrderLine)

	svc := service.NewPOSService(cfg, repos.Device, repos.Order, paymentSvc, club100Svc)
	ctx := context.Background()

	device := fixtures.CreateDevice("POS 1", "pos-token", entDevice.TypePOS, entDevice.StatusApproved)

	t.Run("PayCard processes payment", func(t *testing.T) {
		order := fixtures.CreateOrder(1000, entOrder.StatusPending, entOrder.OriginPos)

		err := svc.PayCard(ctx, order.ID, &device.ID)
		require.NoError(t, err)

		// Verify order status
		updated, err := repos.Order.GetByID(ctx, order.ID)
		require.NoError(t, err)
		require.Equal(t, entOrder.StatusPaid, updated.Status)
	})

	t.Run("PayCard fails for non-pending order", func(t *testing.T) {
		order := fixtures.CreateOrder(1000, entOrder.StatusPaid, entOrder.OriginPos)

		err := svc.PayCard(ctx, order.ID, &device.ID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not_pending")
	})
}

func TestPOSService_PayTwint(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	cfg := TestConfig()

	paymentSvc := service.NewPaymentService(
		cfg,
		repos.Order,
		repos.OrderLine,
		repos.Product,
		repos.MenuSlot,
		repos.Inventory,
		nil,
	)

	club100Svc := service.NewClub100Service(&MockElvantoService{}, repos.Club100Redemption, repos.Settings, repos.OrderLine)

	svc := service.NewPOSService(cfg, repos.Device, repos.Order, paymentSvc, club100Svc)
	ctx := context.Background()

	device := fixtures.CreateDevice("POS 1", "pos-token", entDevice.TypePOS, entDevice.StatusApproved)

	t.Run("PayTwint processes payment", func(t *testing.T) {
		order := fixtures.CreateOrder(1000, entOrder.StatusPending, entOrder.OriginPos)

		err := svc.PayTwint(ctx, order.ID, &device.ID)
		require.NoError(t, err)

		// Verify order status
		updated, err := repos.Order.GetByID(ctx, order.ID)
		require.NoError(t, err)
		require.Equal(t, entOrder.StatusPaid, updated.Status)
	})

	t.Run("PayTwint fails for non-pending order", func(t *testing.T) {
		order := fixtures.CreateOrder(1000, entOrder.StatusPaid, entOrder.OriginPos)

		err := svc.PayTwint(ctx, order.ID, &device.ID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not_pending")
	})
}
