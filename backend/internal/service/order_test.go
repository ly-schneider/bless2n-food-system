package service

import (
    "context"
    "testing"

    "backend/internal/domain"
    "backend/internal/testutil"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

func setupOrderService() (*orderService, *testutil.MockOrderRepository, *testutil.MockOrderItemRepository, *testutil.MockProductRepository, *testutil.MockProductBundleComponentRepository, *testutil.MockInventoryLedgerRepository, *testutil.MockUserRepository, *testutil.MockStationProductRepository, *testutil.MockStationRepository, *testutil.MockDeviceRepository) {
    orderRepo := &testutil.MockOrderRepository{}
    orderItemRepo := &testutil.MockOrderItemRepository{}
    productRepo := &testutil.MockProductRepository{}
    bundleRepo := &testutil.MockProductBundleComponentRepository{}
    ledgerRepo := &testutil.MockInventoryLedgerRepository{}
    userRepo := &testutil.MockUserRepository{}
    stationProductRepo := &testutil.MockStationProductRepository{}
    stationRepo := &testutil.MockStationRepository{}
    deviceRepo := &testutil.MockDeviceRepository{}

    svc := &orderService{
        orderRepo:           orderRepo,
        orderItemRepo:       orderItemRepo,
        productRepo:         productRepo,
        bundleComponentRepo: bundleRepo,
        inventoryLedgerRepo: ledgerRepo,
        userRepo:            userRepo,
        stationProductRepo:  stationProductRepo,
        stationRepo:         stationRepo,
        deviceRepo:          deviceRepo,
    }
    return svc, orderRepo, orderItemRepo, productRepo, bundleRepo, ledgerRepo, userRepo, stationProductRepo, stationRepo, deviceRepo
}

func TestOrderService_CreateOrder_SimpleProduct_Success(t *testing.T) {
    svc, orderRepo, orderItemRepo, productRepo, _, ledgerRepo, _, _, _, _ := setupOrderService()

    pID := primitive.NewObjectID()
    product := &domain.Product{ID: pID, Name: "Water", Type: domain.ProductTypeSimple, Price: 3.5, IsActive: true}
    productRepo.On("GetByID", mock.Anything, pID).Return(product, nil)
    ledgerRepo.On("GetCurrentStock", mock.Anything, pID).Return(10, nil)

    orderRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Order")).Return(nil)
    orderItemRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.OrderItem")).Return(nil)
    ledgerRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.InventoryLedger")).Return(nil)

    email := "buyer@example.com"
    req := CreateOrderRequest{
        ContactEmail: &email,
        Items:        []CreateOrderItemRequest{{ProductID: pID.Hex(), Quantity: 2}},
    }
    resp, err := svc.CreateOrder(context.Background(), req)
    require.NoError(t, err)
    require.NotNil(t, resp)
    assert.True(t, resp.Success)
    assert.Equal(t, "Order created successfully", resp.Message)

    productRepo.AssertExpectations(t)
    orderRepo.AssertExpectations(t)
    orderItemRepo.AssertExpectations(t)
    ledgerRepo.AssertExpectations(t)
}

func TestOrderService_CreateOrder_SimpleProduct_InsufficientStock(t *testing.T) {
    svc, _, _, productRepo, _, ledgerRepo, _, _, _, _ := setupOrderService()

    pID := primitive.NewObjectID()
    product := &domain.Product{ID: pID, Name: "Water", Type: domain.ProductTypeSimple, Price: 3.5, IsActive: true}
    productRepo.On("GetByID", mock.Anything, pID).Return(product, nil)
    ledgerRepo.On("GetCurrentStock", mock.Anything, pID).Return(1, nil)

    email := "buyer@example.com"
    req := CreateOrderRequest{ContactEmail: &email, Items: []CreateOrderItemRequest{{ProductID: pID.Hex(), Quantity: 2}}}
    resp, err := svc.CreateOrder(context.Background(), req)
    require.Error(t, err)
    assert.Contains(t, err.Error(), "insufficient stock")
    assert.Nil(t, resp)
}

func TestOrderService_CreateOrder_InactiveProduct(t *testing.T) {
    svc, _, _, productRepo, _, _, _, _, _, _ := setupOrderService()
    pID := primitive.NewObjectID()
    product := &domain.Product{ID: pID, Name: "Water", Type: domain.ProductTypeSimple, Price: 3.5, IsActive: false}
    productRepo.On("GetByID", mock.Anything, pID).Return(product, nil)

    email := "buyer@example.com"
    req := CreateOrderRequest{ContactEmail: &email, Items: []CreateOrderItemRequest{{ProductID: pID.Hex(), Quantity: 1}}}
    resp, err := svc.CreateOrder(context.Background(), req)
    require.Error(t, err)
    assert.Contains(t, err.Error(), "not active")
    assert.Nil(t, resp)
}

func TestOrderService_CreateOrder_BundleComponentInsufficient(t *testing.T) {
    svc, _, _, productRepo, bundleRepo, ledgerRepo, _, _, _, _ := setupOrderService()
    bundleID := primitive.NewObjectID()
    compID := primitive.NewObjectID()
    bundle := &domain.Product{ID: bundleID, Name: "Combo", Type: domain.ProductTypeBundle, Price: 10.0, IsActive: true}
    comp := &domain.Product{ID: compID, Name: "Part", Type: domain.ProductTypeSimple, Price: 4.0, IsActive: true}

    productRepo.On("GetByID", mock.Anything, bundleID).Return(bundle, nil)
    productRepo.On("GetByID", mock.Anything, compID).Return(comp, nil)
    bundleRepo.On("GetByBundleID", mock.Anything, bundleID).Return([]*domain.ProductBundleComponent{{BundleID: bundleID, ComponentProductID: compID, Quantity: 2}}, nil)
    // Inventory check is called first for the bundle product itself, then for each component
    ledgerRepo.On("GetCurrentStock", mock.Anything, bundleID).Return(10, nil)
    ledgerRepo.On("GetCurrentStock", mock.Anything, compID).Return(1, nil)

    email := "buyer@example.com"
    req := CreateOrderRequest{ContactEmail: &email, Items: []CreateOrderItemRequest{{ProductID: bundleID.Hex(), Quantity: 1}}}
    resp, err := svc.CreateOrder(context.Background(), req)
    require.Error(t, err)
    assert.Contains(t, err.Error(), "insufficient stock for bundle component")
    assert.Nil(t, resp)
}

func TestOrderService_GetOrder_NotFound(t *testing.T) {
    svc, orderRepo, orderItemRepo, _, _, _, _, _, _, _ := setupOrderService()
    id := primitive.NewObjectID()
    orderRepo.On("GetByID", mock.Anything, id).Return(nil, nil)

    resp, err := svc.GetOrder(context.Background(), id.Hex())
    require.Error(t, err)
    assert.Contains(t, err.Error(), "not found")
    assert.Nil(t, resp)

    // invalid format
    resp, err = svc.GetOrder(context.Background(), "bad-id")
    require.Error(t, err)
    assert.Contains(t, err.Error(), "invalid order ID format")
    assert.Nil(t, resp)

    orderItemRepo.AssertExpectations(t)
}

func TestOrderService_DeleteOrder_Pending_ReversesInventory(t *testing.T) {
    svc, orderRepo, orderItemRepo, _, _, ledgerRepo, _, _, _, _ := setupOrderService()
    id := primitive.NewObjectID()
    itemProduct := primitive.NewObjectID()
    order := &domain.Order{ID: id, Status: domain.OrderStatusPending}
    items := []*domain.OrderItem{{ID: primitive.NewObjectID(), OrderID: id, ProductID: itemProduct, Type: domain.OrderItemTypeSimple, Quantity: 2}}

    orderRepo.On("GetByID", mock.Anything, id).Return(order, nil)
    orderItemRepo.On("GetByOrderID", mock.Anything, id).Return(items, nil)
    ledgerRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.InventoryLedger")).Return(nil)
    orderItemRepo.On("DeleteByOrderID", mock.Anything, id).Return(nil)
    orderRepo.On("Delete", mock.Anything, id).Return(nil)

    resp, err := svc.DeleteOrder(context.Background(), id.Hex())
    require.NoError(t, err)
    require.NotNil(t, resp)
    assert.True(t, resp.Success)
}

func TestOrderService_UpdateOrder_UnsupportedItems(t *testing.T) {
    svc, orderRepo, _, _, _, _, _, _, _, _ := setupOrderService()
    id := primitive.NewObjectID()
    order := &domain.Order{ID: id, Status: domain.OrderStatusPending}
    orderRepo.On("GetByID", mock.Anything, id).Return(order, nil)

    items := []CreateOrderItemRequest{{ProductID: primitive.NewObjectID().Hex(), Quantity: 1}}
    _, err := svc.UpdateOrder(context.Background(), id.Hex(), UpdateOrderRequest{Items: &items})
    require.Error(t, err)
    assert.Contains(t, err.Error(), "not supported")
}

func TestOrderService_CreateOrder_InvalidProductID(t *testing.T) {
    svc, _, _, _, _, _, _, _, _, _ := setupOrderService()
    email := "buyer@example.com"
    _, err := svc.CreateOrder(context.Background(), CreateOrderRequest{ContactEmail: &email, Items: []CreateOrderItemRequest{{ProductID: "bad", Quantity: 1}}})
    require.Error(t, err)
    assert.Contains(t, err.Error(), "invalid product ID format")
}

func TestOrderService_CreateOrder_ProductNotFound(t *testing.T) {
    svc, _, _, productRepo, _, _, _, _, _, _ := setupOrderService()
    pID := primitive.NewObjectID()
    productRepo.On("GetByID", mock.Anything, pID).Return(nil, nil)
    email := "buyer@example.com"
    _, err := svc.CreateOrder(context.Background(), CreateOrderRequest{ContactEmail: &email, Items: []CreateOrderItemRequest{{ProductID: pID.Hex(), Quantity: 1}}})
    require.Error(t, err)
    assert.Contains(t, err.Error(), "product not found")
}
