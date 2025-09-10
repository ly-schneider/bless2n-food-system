package testutil

import (
    "context"
    "time"

    "backend/internal/domain"

    "github.com/stretchr/testify/mock"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// MockUserRepository provides a mock implementation of the UserRepository interface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		user.ID = primitive.NewObjectID()
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *MockUserRepository) ListCustomers(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *MockUserRepository) CountCustomers(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) Disable(ctx context.Context, id primitive.ObjectID, reason string) error {
	args := m.Called(ctx, id, reason)
	return args.Error(0)
}

func (m *MockUserRepository) Enable(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockCategoryRepository provides a mock implementation of the CategoryRepository interface
type MockCategoryRepository struct {
    mock.Mock
}

func (m *MockCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	args := m.Called(ctx, category)
	if args.Get(0) == nil {
		category.ID = primitive.NewObjectID()
		category.CreatedAt = time.Now()
		category.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *MockCategoryRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *MockCategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCategoryRepository) GetByName(ctx context.Context, name string) (*domain.Category, error) {
    args := m.Called(ctx, name)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *MockCategoryRepository) List(ctx context.Context, activeOnly bool, limit, offset int) ([]*domain.Category, error) {
    args := m.Called(ctx, activeOnly, limit, offset)
    return args.Get(0).([]*domain.Category), args.Error(1)
}

func (m *MockCategoryRepository) SetActive(ctx context.Context, id primitive.ObjectID, isActive bool) error {
    args := m.Called(ctx, id, isActive)
    return args.Error(0)
}

// MockProductRepository provides a mock implementation of the ProductRepository interface
type MockProductRepository struct {
    mock.Mock
}

func (m *MockProductRepository) Create(ctx context.Context, product *domain.Product) error {
	args := m.Called(ctx, product)
	if args.Get(0) == nil {
		product.ID = primitive.NewObjectID()
		product.CreatedAt = time.Now()
		product.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *MockProductRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductRepository) GetByCategoryID(ctx context.Context, categoryID primitive.ObjectID, activeOnly bool, limit, offset int) ([]*domain.Product, error) {
    args := m.Called(ctx, categoryID, activeOnly, limit, offset)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*domain.Product), args.Error(1)
}

func (m *MockProductRepository) Update(ctx context.Context, product *domain.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepository) GetByType(ctx context.Context, productType domain.ProductType, limit, offset int) ([]*domain.Product, error) {
    args := m.Called(ctx, productType, limit, offset)
    return args.Get(0).([]*domain.Product), args.Error(1)
}

func (m *MockProductRepository) SetActive(ctx context.Context, id primitive.ObjectID, isActive bool) error {
    args := m.Called(ctx, id, isActive)
    return args.Error(0)
}

func (m *MockProductRepository) Search(ctx context.Context, query string, limit, offset int) ([]*domain.Product, error) {
    args := m.Called(ctx, query, limit, offset)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*domain.Product), args.Error(1)
}

// MockOrderRepository provides a mock implementation of the OrderRepository interface
type MockOrderRepository struct {
    mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	if args.Get(0) == nil {
		order.ID = primitive.NewObjectID()
		order.CreatedAt = time.Now()
		order.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) GetByCustomerID(ctx context.Context, customerID primitive.ObjectID, limit, offset int) ([]*domain.Order, error) {
    args := m.Called(ctx, customerID, limit, offset)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) GetByContactEmail(ctx context.Context, email string, limit, offset int) ([]*domain.Order, error) {
    args := m.Called(ctx, email, limit, offset)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) GetByStatus(ctx context.Context, status domain.OrderStatus, limit, offset int) ([]*domain.Order, error) {
    args := m.Called(ctx, status, limit, offset)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) Update(ctx context.Context, order *domain.Order) error {
    args := m.Called(ctx, order)
    return args.Error(0)
}

func (m *MockOrderRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status domain.OrderStatus) error {
    args := m.Called(ctx, id, status)
    return args.Error(0)
}

func (m *MockOrderRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrderRepository) List(ctx context.Context, limit, offset int) ([]*domain.Order, error) {
    args := m.Called(ctx, limit, offset)
    return args.Get(0).([]*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) GetRecent(ctx context.Context, limit int) ([]*domain.Order, error) {
    args := m.Called(ctx, limit)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) ListByCustomer(ctx context.Context, customerID primitive.ObjectID, limit, offset int) ([]*domain.Order, error) {
	args := m.Called(ctx, customerID, limit, offset)
	return args.Get(0).([]*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockOrderRepository) CountByCustomer(ctx context.Context, customerID primitive.ObjectID) (int, error) {
	args := m.Called(ctx, customerID)
	return args.Int(0), args.Error(1)
}

// MockEmailService provides a mock implementation of the EmailService interface
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendOTP(ctx context.Context, to, otp string) error {
	args := m.Called(ctx, to, otp)
	return args.Error(0)
}

// MockOTPService provides a mock implementation of the OTPService interface
type MockOTPService struct {
	mock.Mock
}

func (m *MockOTPService) GenerateAndSend(ctx context.Context, userID primitive.ObjectID, email string, tokenType domain.TokenType) error {
	args := m.Called(ctx, userID, email, tokenType)
	return args.Error(0)
}

func (m *MockOTPService) Verify(ctx context.Context, userID primitive.ObjectID, otp string, tokenType domain.TokenType) error {
	args := m.Called(ctx, userID, otp, tokenType)
	return args.Error(0)
}

// Note: TokenService mocks omitted intentionally to avoid coupling to service package.

// MockAdminInviteRepository provides a mock implementation of the AdminInviteRepository interface
type MockAdminInviteRepository struct {
	mock.Mock
}

func (m *MockAdminInviteRepository) Create(ctx context.Context, invite *domain.AdminInvite) error {
    args := m.Called(ctx, invite)
    if args.Get(0) == nil {
        invite.ID = primitive.NewObjectID()
    }
    return args.Error(0)
}

func (m *MockAdminInviteRepository) GetByCode(ctx context.Context, code string) (*domain.AdminInvite, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AdminInvite), args.Error(1)
}

func (m *MockAdminInviteRepository) Update(ctx context.Context, invite *domain.AdminInvite) error {
	args := m.Called(ctx, invite)
	return args.Error(0)
}

func (m *MockAdminInviteRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAdminInviteRepository) List(ctx context.Context, limit, offset int) ([]*domain.AdminInvite, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.AdminInvite), args.Error(1)
}

func (m *MockAdminInviteRepository) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

// MockStationRepository provides a mock implementation of the StationRepository interface
type MockStationRepository struct {
	mock.Mock
}

func (m *MockStationRepository) Create(ctx context.Context, station *domain.Station) error {
	args := m.Called(ctx, station)
	if args.Get(0) == nil {
		station.ID = primitive.NewObjectID()
		station.CreatedAt = time.Now()
		station.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *MockStationRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Station, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Station), args.Error(1)
}

func (m *MockStationRepository) GetByName(ctx context.Context, name string) (*domain.Station, error) {
    args := m.Called(ctx, name)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.Station), args.Error(1)
}

func (m *MockStationRepository) Update(ctx context.Context, station *domain.Station) error {
	args := m.Called(ctx, station)
	return args.Error(0)
}

func (m *MockStationRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStationRepository) List(ctx context.Context, limit, offset int) ([]*domain.Station, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.Station), args.Error(1)
}

func (m *MockStationRepository) Count(ctx context.Context) (int, error) {
    args := m.Called(ctx)
    return args.Int(0), args.Error(1)
}

func (m *MockStationRepository) ListByStatus(ctx context.Context, status domain.StationStatus, limit, offset int) ([]*domain.Station, error) {
    args := m.Called(ctx, status, limit, offset)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*domain.Station), args.Error(1)
}

func (m *MockStationRepository) ApproveStation(ctx context.Context, stationID, adminID primitive.ObjectID) error {
    args := m.Called(ctx, stationID, adminID)
    return args.Error(0)
}

func (m *MockStationRepository) RejectStation(ctx context.Context, stationID, adminID primitive.ObjectID, reason string) error {
    args := m.Called(ctx, stationID, adminID, reason)
    return args.Error(0)
}

// MockProductBundleComponentRepository provides a mock implementation of the ProductBundleComponentRepository interface
type MockProductBundleComponentRepository struct {
	mock.Mock
}

func (m *MockProductBundleComponentRepository) Create(ctx context.Context, component *domain.ProductBundleComponent) error {
    args := m.Called(ctx, component)
    return args.Error(0)
}

func (m *MockProductBundleComponentRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.ProductBundleComponent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProductBundleComponent), args.Error(1)
}

func (m *MockProductBundleComponentRepository) Update(ctx context.Context, component *domain.ProductBundleComponent) error {
	args := m.Called(ctx, component)
	return args.Error(0)
}

func (m *MockProductBundleComponentRepository) Delete(ctx context.Context, bundleID, componentID primitive.ObjectID) error {
    args := m.Called(ctx, bundleID, componentID)
    return args.Error(0)
}

func (m *MockProductBundleComponentRepository) GetByBundleID(ctx context.Context, bundleID primitive.ObjectID) ([]*domain.ProductBundleComponent, error) {
	args := m.Called(ctx, bundleID)
	return args.Get(0).([]*domain.ProductBundleComponent), args.Error(1)
}

func (m *MockProductBundleComponentRepository) DeleteByBundleID(ctx context.Context, bundleID primitive.ObjectID) error {
	args := m.Called(ctx, bundleID)
	return args.Error(0)
}

// MockStationProductRepository provides a mock implementation of the StationProductRepository interface
type MockStationProductRepository struct {
	mock.Mock
}

func (m *MockStationProductRepository) Create(ctx context.Context, stationProduct *domain.StationProduct) error {
    args := m.Called(ctx, stationProduct)
    return args.Error(0)
}

func (m *MockStationProductRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.StationProduct, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.StationProduct), args.Error(1)
}

func (m *MockStationProductRepository) Delete(ctx context.Context, stationID, productID primitive.ObjectID) error {
    args := m.Called(ctx, stationID, productID)
    return args.Error(0)
}

func (m *MockStationProductRepository) GetByStationID(ctx context.Context, stationID primitive.ObjectID) ([]*domain.StationProduct, error) {
	args := m.Called(ctx, stationID)
	return args.Get(0).([]*domain.StationProduct), args.Error(1)
}

func (m *MockStationProductRepository) GetByProductID(ctx context.Context, productID primitive.ObjectID) ([]*domain.StationProduct, error) {
	args := m.Called(ctx, productID)
	return args.Get(0).([]*domain.StationProduct), args.Error(1)
}

func (m *MockStationProductRepository) DeleteByStationID(ctx context.Context, stationID primitive.ObjectID) error {
	args := m.Called(ctx, stationID)
	return args.Error(0)
}

func (m *MockStationProductRepository) DeleteByProductID(ctx context.Context, productID primitive.ObjectID) error {
    args := m.Called(ctx, productID)
    return args.Error(0)
}

// MockOrderItemRepository provides a mock implementation of the OrderItemRepository interface
type MockOrderItemRepository struct{ mock.Mock }

func (m *MockOrderItemRepository) Create(ctx context.Context, item *domain.OrderItem) error {
    args := m.Called(ctx, item)
    if args.Get(0) == nil {
        item.ID = primitive.NewObjectID()
        item.IsRedeemed = false
    }
    return args.Error(0)
}

func (m *MockOrderItemRepository) CreateBatch(ctx context.Context, items []*domain.OrderItem) error {
    args := m.Called(ctx, items)
    return args.Error(0)
}

func (m *MockOrderItemRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.OrderItem, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.OrderItem), args.Error(1)
}

func (m *MockOrderItemRepository) GetByOrderID(ctx context.Context, orderID primitive.ObjectID) ([]*domain.OrderItem, error) {
    args := m.Called(ctx, orderID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*domain.OrderItem), args.Error(1)
}

func (m *MockOrderItemRepository) GetUnredeemedByOrderID(ctx context.Context, orderID primitive.ObjectID) ([]*domain.OrderItem, error) {
    args := m.Called(ctx, orderID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*domain.OrderItem), args.Error(1)
}

func (m *MockOrderItemRepository) GetByProductID(ctx context.Context, productID primitive.ObjectID, unredeemedOnly bool) ([]*domain.OrderItem, error) {
    args := m.Called(ctx, productID, unredeemedOnly)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*domain.OrderItem), args.Error(1)
}

func (m *MockOrderItemRepository) GetByParentItemID(ctx context.Context, parentItemID primitive.ObjectID) ([]*domain.OrderItem, error) {
    args := m.Called(ctx, parentItemID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*domain.OrderItem), args.Error(1)
}

func (m *MockOrderItemRepository) Update(ctx context.Context, item *domain.OrderItem) error {
    args := m.Called(ctx, item)
    return args.Error(0)
}

func (m *MockOrderItemRepository) MarkAsRedeemed(ctx context.Context, id primitive.ObjectID, stationID, deviceID primitive.ObjectID) error {
    args := m.Called(ctx, id, stationID, deviceID)
    return args.Error(0)
}

func (m *MockOrderItemRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

func (m *MockOrderItemRepository) DeleteByOrderID(ctx context.Context, orderID primitive.ObjectID) error {
    args := m.Called(ctx, orderID)
    return args.Error(0)
}

func (m *MockOrderItemRepository) GetRedeemedByStation(ctx context.Context, stationID primitive.ObjectID, limit, offset int) ([]*domain.OrderItem, error) {
    args := m.Called(ctx, stationID, limit, offset)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*domain.OrderItem), args.Error(1)
}

func (m *MockOrderItemRepository) GetRedeemedByDevice(ctx context.Context, deviceID primitive.ObjectID, limit, offset int) ([]*domain.OrderItem, error) {
    args := m.Called(ctx, deviceID, limit, offset)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*domain.OrderItem), args.Error(1)
}

// MockInventoryLedgerRepository provides a mock implementation of InventoryLedgerRepository
type MockInventoryLedgerRepository struct{ mock.Mock }

func (m *MockInventoryLedgerRepository) Create(ctx context.Context, entry *domain.InventoryLedger) error {
    args := m.Called(ctx, entry)
    if args.Get(0) == nil {
        entry.ID = primitive.NewObjectID()
        entry.Timestamp = time.Now()
    }
    return args.Error(0)
}

func (m *MockInventoryLedgerRepository) GetByProductID(ctx context.Context, productID primitive.ObjectID, limit, offset int) ([]*domain.InventoryLedger, error) {
    args := m.Called(ctx, productID, limit, offset)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*domain.InventoryLedger), args.Error(1)
}

func (m *MockInventoryLedgerRepository) GetCurrentStock(ctx context.Context, productID primitive.ObjectID) (int, error) {
    args := m.Called(ctx, productID)
    return args.Int(0), args.Error(1)
}

func (m *MockInventoryLedgerRepository) GetStockMovements(ctx context.Context, productID primitive.ObjectID, reason domain.InventoryReason, limit, offset int) ([]*domain.InventoryLedger, error) {
    args := m.Called(ctx, productID, reason, limit, offset)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*domain.InventoryLedger), args.Error(1)
}

// MockDeviceRepository provides a mock implementation of DeviceRepository
type MockDeviceRepository struct{ mock.Mock }

func (m *MockDeviceRepository) Create(ctx context.Context, device *domain.Device) error {
    args := m.Called(ctx, device)
    if args.Get(0) == nil {
        device.ID = primitive.NewObjectID()
        device.IsActive = true
    }
    return args.Error(0)
}

func (m *MockDeviceRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Device, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.Device), args.Error(1)
}

func (m *MockDeviceRepository) GetByStationID(ctx context.Context, stationID primitive.ObjectID) ([]*domain.Device, error) {
    args := m.Called(ctx, stationID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*domain.Device), args.Error(1)
}

func (m *MockDeviceRepository) Update(ctx context.Context, device *domain.Device) error {
    args := m.Called(ctx, device)
    return args.Error(0)
}

func (m *MockDeviceRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

func (m *MockDeviceRepository) SetActive(ctx context.Context, id primitive.ObjectID, isActive bool) error {
    args := m.Called(ctx, id, isActive)
    return args.Error(0)
}
