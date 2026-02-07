package integration

import (
	"context"
	"testing"

	entDevice "backend/internal/generated/ent/device"
	entOrder "backend/internal/generated/ent/order"
	"backend/internal/generated/ent/orderline"
	"backend/internal/generated/ent/product"
	"backend/internal/service"

	"github.com/stretchr/testify/require"
)

func TestStationService_GetStationByKey(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	cfg := TestConfig()

	svc := service.NewStationService(
		cfg,
		tdb.Client,
		repos.Device,
		repos.DeviceProduct,
		repos.OrderLine,
		repos.OrderRedemption,
		repos.Idempotency,
	)
	ctx := context.Background()

	t.Run("GetStationByKey returns station device", func(t *testing.T) {
		fixtures.CreateDevice("Station 1", "station-key-1", entDevice.TypeSTATION, entDevice.StatusApproved)

		dev, err := svc.GetStationByKey(ctx, "station-key-1")
		require.NoError(t, err)
		require.Equal(t, "Station 1", dev.Name)
		require.Equal(t, entDevice.TypeSTATION, dev.Type)
	})

	t.Run("GetStationByKey returns error for POS device", func(t *testing.T) {
		fixtures.CreateDevice("POS 1", "pos-key-1", entDevice.TypePOS, entDevice.StatusApproved)

		_, err := svc.GetStationByKey(ctx, "pos-key-1")
		require.Error(t, err)
		require.Equal(t, "device_not_station", err.Error())
	})

	t.Run("GetStationByKey returns error for non-existent key", func(t *testing.T) {
		_, err := svc.GetStationByKey(ctx, "non-existent-key")
		require.Error(t, err)
	})
}

func TestStationService_AssignedItemsForOrder(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	cfg := TestConfig()

	svc := service.NewStationService(
		cfg,
		tdb.Client,
		repos.Device,
		repos.DeviceProduct,
		repos.OrderLine,
		repos.OrderRedemption,
		repos.Idempotency,
	)
	ctx := context.Background()

	// Setup test data
	category := fixtures.CreateCategory("Drinks", 1, true)
	cola := fixtures.CreateProduct("Cola", category.ID, 350, product.TypeSimple, nil)
	sprite := fixtures.CreateProduct("Sprite", category.ID, 350, product.TypeSimple, nil)
	burger := fixtures.CreateProduct("Burger", category.ID, 1200, product.TypeSimple, nil)

	station := fixtures.CreateDevice("Drinks Station", "station-key", entDevice.TypeSTATION, entDevice.StatusApproved)

	// Assign only cola and sprite to station
	fixtures.AssignProductToDevice(station.ID, cola.ID)
	fixtures.AssignProductToDevice(station.ID, sprite.ID)

	// Create order with all three products
	ord := fixtures.CreateOrder(1900, entOrder.StatusPaid, entOrder.OriginShop)
	fixtures.CreateOrderLine(ord.ID, cola.ID, "Cola", 2, 350, orderline.LineTypeSimple)
	fixtures.CreateOrderLine(ord.ID, sprite.ID, "Sprite", 1, 350, orderline.LineTypeSimple)
	fixtures.CreateOrderLine(ord.ID, burger.ID, "Burger", 1, 1200, orderline.LineTypeSimple)

	t.Run("AssignedItemsForOrder returns only station-assigned items", func(t *testing.T) {
		items, err := svc.AssignedItemsForOrder(ctx, station.ID, ord.ID)
		require.NoError(t, err)
		require.Len(t, items, 2)

		// Should only have cola and sprite, not burger
		var titles []string
		for _, item := range items {
			titles = append(titles, item.Title)
		}
		require.Contains(t, titles, "Cola")
		require.Contains(t, titles, "Sprite")
		require.NotContains(t, titles, "Burger")
	})

	t.Run("AssignedItemsForOrder returns empty for station with no assigned products", func(t *testing.T) {
		emptyStation := fixtures.CreateDevice("Empty Station", "empty-key", entDevice.TypeSTATION, entDevice.StatusApproved)

		items, err := svc.AssignedItemsForOrder(ctx, emptyStation.ID, ord.ID)
		require.NoError(t, err)
		require.Empty(t, items)
	})
}

func TestStationService_RedeemAssigned(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	cfg := TestConfig()

	svc := service.NewStationService(
		cfg,
		tdb.Client,
		repos.Device,
		repos.DeviceProduct,
		repos.OrderLine,
		repos.OrderRedemption,
		repos.Idempotency,
	)
	ctx := context.Background()

	// Setup test data
	category := fixtures.CreateCategory("Drinks", 1, true)
	cola := fixtures.CreateProduct("Cola", category.ID, 350, product.TypeSimple, nil)
	sprite := fixtures.CreateProduct("Sprite", category.ID, 350, product.TypeSimple, nil)

	station := fixtures.CreateDevice("Drinks Station", "station-key", entDevice.TypeSTATION, entDevice.StatusApproved)
	fixtures.AssignProductToDevice(station.ID, cola.ID)
	fixtures.AssignProductToDevice(station.ID, sprite.ID)

	ord := fixtures.CreateOrder(700, entOrder.StatusPaid, entOrder.OriginShop)
	fixtures.CreateOrderLine(ord.ID, cola.ID, "Cola", 1, 350, orderline.LineTypeSimple)
	fixtures.CreateOrderLine(ord.ID, sprite.ID, "Sprite", 1, 350, orderline.LineTypeSimple)

	t.Run("RedeemAssigned redeems all assigned items", func(t *testing.T) {
		result, err := svc.RedeemAssigned(ctx, station.ID, ord.ID, "")
		require.NoError(t, err)

		require.Equal(t, ord.ID.String(), result["orderId"])
		require.Equal(t, station.ID.String(), result["stationId"])
		require.EqualValues(t, 2, result["matched"])
		require.EqualValues(t, 2, result["redeemed"])
		require.NotNil(t, result["redeemedAt"])
	})

	t.Run("RedeemAssigned is idempotent with same key", func(t *testing.T) {
		tdb.Cleanup(t)

		// Recreate test data
		cat := fixtures.CreateCategory("Drinks", 1, true)
		c := fixtures.CreateProduct("Cola", cat.ID, 350, product.TypeSimple, nil)
		stn := fixtures.CreateDevice("Station", "key", entDevice.TypeSTATION, entDevice.StatusApproved)
		fixtures.AssignProductToDevice(stn.ID, c.ID)
		o := fixtures.CreateOrder(350, entOrder.StatusPaid, entOrder.OriginShop)
		fixtures.CreateOrderLine(o.ID, c.ID, "Cola", 1, 350, orderline.LineTypeSimple)

		idemKey := "test-idem-key-123"

		// First call
		result1, err := svc.RedeemAssigned(ctx, stn.ID, o.ID, idemKey)
		require.NoError(t, err)
		require.EqualValues(t, 1, result1["redeemed"])

		// Second call with same key should return same result
		result2, err := svc.RedeemAssigned(ctx, stn.ID, o.ID, idemKey)
		require.NoError(t, err)
		require.EqualValues(t, 1, result2["redeemed"])
		require.Equal(t, result1["redeemedAt"], result2["redeemedAt"])
	})

	t.Run("RedeemAssigned does not re-redeem already redeemed items", func(t *testing.T) {
		tdb.Cleanup(t)

		// Recreate test data
		cat := fixtures.CreateCategory("Drinks", 1, true)
		c := fixtures.CreateProduct("Cola", cat.ID, 350, product.TypeSimple, nil)
		stn := fixtures.CreateDevice("Station", "key", entDevice.TypeSTATION, entDevice.StatusApproved)
		fixtures.AssignProductToDevice(stn.ID, c.ID)
		o := fixtures.CreateOrder(350, entOrder.StatusPaid, entOrder.OriginShop)
		fixtures.CreateOrderLine(o.ID, c.ID, "Cola", 1, 350, orderline.LineTypeSimple)

		// First redemption
		result1, err := svc.RedeemAssigned(ctx, stn.ID, o.ID, "")
		require.NoError(t, err)
		require.EqualValues(t, 1, result1["redeemed"])

		// Second redemption without idempotency key should not redeem again
		result2, err := svc.RedeemAssigned(ctx, stn.ID, o.ID, "")
		require.NoError(t, err)
		require.EqualValues(t, 0, result2["redeemed"]) // Already redeemed
		require.EqualValues(t, 1, result2["matched"])  // Still matches
	})
}
