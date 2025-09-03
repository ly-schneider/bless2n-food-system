package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/internal/domain"
	"backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderService interface {
	CreateOrder(ctx context.Context, req CreateOrderRequest) (*CreateOrderResponse, error)
	GetOrder(ctx context.Context, orderID string) (*GetOrderResponse, error)
	UpdateOrder(ctx context.Context, orderID string, req UpdateOrderRequest) (*UpdateOrderResponse, error)
	DeleteOrder(ctx context.Context, orderID string) (*DeleteOrderResponse, error)
	ListOrders(ctx context.Context, req ListOrdersRequest) (*ListOrdersResponse, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status domain.OrderStatus) (*UpdateOrderStatusResponse, error)
	GetOrdersByCustomer(ctx context.Context, customerID string, limit, offset int) (*ListOrdersResponse, error)
	GetOrdersByEmail(ctx context.Context, email string, limit, offset int) (*ListOrdersResponse, error)
	GetOrderForRedemption(ctx context.Context, orderID, stationID string) (*OrderRedemptionResponse, error)
	RedeemOrderItems(ctx context.Context, req RedeemOrderItemsRequest) (*RedeemOrderItemsResponse, error)
}

type CreateOrderRequest struct {
	CustomerID   *string                `json:"customer_id,omitempty"`
	ContactEmail *string                `json:"contact_email,omitempty"`
	Items        []CreateOrderItemRequest `json:"items" validate:"required,dive"`
}

type CreateOrderItemRequest struct {
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required,gt=0"`
}

type CreateOrderResponse struct {
	Order   OrderDTO `json:"order"`
	Message string   `json:"message"`
	Success bool     `json:"success"`
}

type GetOrderResponse struct {
	Order OrderDTO          `json:"order"`
	Items []OrderItemDTO    `json:"items"`
}

type UpdateOrderRequest struct {
	ContactEmail *string `json:"contact_email,omitempty"`
	Items        *[]CreateOrderItemRequest `json:"items,omitempty" validate:"omitempty,dive"`
}

type UpdateOrderResponse struct {
	Order   OrderDTO `json:"order"`
	Message string   `json:"message"`
	Success bool     `json:"success"`
}

type DeleteOrderResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type ListOrdersRequest struct {
	Status   *domain.OrderStatus `json:"status,omitempty"`
	Limit    int                 `json:"limit,omitempty"`
	Offset   int                 `json:"offset,omitempty"`
}

type ListOrdersResponse struct {
	Orders []OrderDTO `json:"orders"`
	Total  int        `json:"total"`
}

type UpdateOrderStatusResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type OrderRedemptionResponse struct {
	Order             OrderDTO                  `json:"order"`
	RedeemableItems   []RedeemableOrderItemDTO  `json:"redeemable_items"`
	AlreadyRedeemed   []OrderItemDTO            `json:"already_redeemed"`
	Message           string                    `json:"message"`
}

type RedeemableOrderItemDTO struct {
	ID                string  `json:"id"`
	ProductID         string  `json:"product_id"`
	ProductName       string  `json:"product_name"`
	Type              string  `json:"type"`
	Quantity          int     `json:"quantity"`
	PricePerUnit      float64 `json:"price_per_unit"`
	IsBundle          bool    `json:"is_bundle"`
	ComponentItems    []OrderItemDTO `json:"component_items,omitempty"`
}

type RedeemOrderItemsRequest struct {
	OrderID   string `json:"order_id" validate:"required"`
	StationID string `json:"station_id" validate:"required"`
	DeviceID  string `json:"device_id" validate:"required"`
}

type RedeemOrderItemsResponse struct {
	RedeemedItems []OrderItemDTO `json:"redeemed_items"`
	Message       string         `json:"message"`
	Success       bool           `json:"success"`
}

type OrderDTO struct {
	ID           string  `json:"id"`
	CustomerID   *string `json:"customer_id,omitempty"`
	ContactEmail *string `json:"contact_email,omitempty"`
	Total        float64 `json:"total"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type OrderItemDTO struct {
	ID                string  `json:"id"`
	OrderID           string  `json:"order_id"`
	ProductID         string  `json:"product_id"`
	ParentItemID      *string `json:"parent_item_id,omitempty"`
	Type              string  `json:"type"`
	Title             string  `json:"title"`
	Quantity          int     `json:"quantity"`
	PricePerUnit      float64 `json:"price_per_unit"`
	IsRedeemed        bool    `json:"is_redeemed"`
	RedeemedAt        *string `json:"redeemed_at,omitempty"`
	RedeemedStationID *string `json:"redeemed_station_id,omitempty"`
	RedeemedDeviceID  *string `json:"redeemed_device_id,omitempty"`
}

type orderService struct {
	orderRepo           repository.OrderRepository
	orderItemRepo       repository.OrderItemRepository
	productRepo         repository.ProductRepository
	bundleComponentRepo repository.ProductBundleComponentRepository
	inventoryLedgerRepo repository.InventoryLedgerRepository
	userRepo            repository.UserRepository
	stationProductRepo  repository.StationProductRepository
	stationRepo         repository.StationRepository
	deviceRepo          repository.DeviceRepository
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
	productRepo repository.ProductRepository,
	bundleComponentRepo repository.ProductBundleComponentRepository,
	inventoryLedgerRepo repository.InventoryLedgerRepository,
	userRepo repository.UserRepository,
	stationProductRepo repository.StationProductRepository,
	stationRepo repository.StationRepository,
	deviceRepo repository.DeviceRepository,
) OrderService {
	return &orderService{
		orderRepo:           orderRepo,
		orderItemRepo:       orderItemRepo,
		productRepo:         productRepo,
		bundleComponentRepo: bundleComponentRepo,
		inventoryLedgerRepo: inventoryLedgerRepo,
		userRepo:            userRepo,
		stationProductRepo:  stationProductRepo,
		stationRepo:         stationRepo,
		deviceRepo:          deviceRepo,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*CreateOrderResponse, error) {
	if req.CustomerID == nil && req.ContactEmail == nil {
		return nil, errors.New("either customer_id or contact_email must be provided")
	}

	if len(req.Items) == 0 {
		return nil, errors.New("order must contain at least one item")
	}

	var customerID *primitive.ObjectID
	if req.CustomerID != nil {
		id, err := primitive.ObjectIDFromHex(*req.CustomerID)
		if err != nil {
			return nil, errors.New("invalid customer ID format")
		}
		
		// Verify customer exists
		customer, err := s.userRepo.GetByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to verify customer: %w", err)
		}
		if customer == nil {
			return nil, errors.New("customer not found")
		}
		customerID = &id
	}

	// Calculate total and validate items
	var orderItems []*domain.OrderItem
	var total float64

	for _, itemReq := range req.Items {
		productID, err := primitive.ObjectIDFromHex(itemReq.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid product ID format: %s", itemReq.ProductID)
		}

		product, err := s.productRepo.GetByID(ctx, productID)
		if err != nil {
			return nil, fmt.Errorf("failed to get product: %w", err)
		}
		if product == nil {
			return nil, fmt.Errorf("product not found: %s", itemReq.ProductID)
		}

		if !product.IsActive {
			return nil, fmt.Errorf("product is not active: %s", product.Name)
		}

		// Check inventory availability
		currentStock, err := s.inventoryLedgerRepo.GetCurrentStock(ctx, productID)
		if err != nil {
			return nil, fmt.Errorf("failed to check inventory for product %s: %w", product.Name, err)
		}

		if product.Type == domain.ProductTypeSimple {
			if currentStock < itemReq.Quantity {
				return nil, fmt.Errorf("insufficient stock for product %s: available %d, requested %d", product.Name, currentStock, itemReq.Quantity)
			}

			// Create simple product order item
			orderItem := &domain.OrderItem{
				ProductID:    productID,
				Type:         domain.OrderItemTypeSimple,
				Title:        product.Name,
				Quantity:     itemReq.Quantity,
				PricePerUnit: product.Price,
			}
			orderItems = append(orderItems, orderItem)
			total += product.Price * float64(itemReq.Quantity)

		} else if product.Type == domain.ProductTypeBundle {
			// Handle bundle products
			bundleComponents, err := s.bundleComponentRepo.GetByBundleID(ctx, productID)
			if err != nil {
				return nil, fmt.Errorf("failed to get bundle components: %w", err)
			}

			if len(bundleComponents) == 0 {
				return nil, fmt.Errorf("bundle has no components: %s", product.Name)
			}

			// Check stock for all components
			for _, component := range bundleComponents {
				componentStock, err := s.inventoryLedgerRepo.GetCurrentStock(ctx, component.ComponentProductID)
				if err != nil {
					return nil, fmt.Errorf("failed to check component stock: %w", err)
				}

				requiredComponentQty := component.Quantity * itemReq.Quantity
				if componentStock < requiredComponentQty {
					componentProduct, _ := s.productRepo.GetByID(ctx, component.ComponentProductID)
					componentName := "Unknown"
					if componentProduct != nil {
						componentName = componentProduct.Name
					}
					return nil, fmt.Errorf("insufficient stock for bundle component %s: available %d, required %d", componentName, componentStock, requiredComponentQty)
				}
			}

			// Create bundle order item
			bundleItem := &domain.OrderItem{
				ProductID:    productID,
				Type:         domain.OrderItemTypeBundle,
				Title:        product.Name,
				Quantity:     itemReq.Quantity,
				PricePerUnit: product.Price,
			}
			orderItems = append(orderItems, bundleItem)
			total += product.Price * float64(itemReq.Quantity)
		}
	}

	// Create the order
	order := &domain.Order{
		CustomerID:   customerID,
		ContactEmail: req.ContactEmail,
		Total:        total,
		Status:       domain.OrderStatusPending,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Create order items and update inventory
	for _, orderItem := range orderItems {
		orderItem.OrderID = order.ID
		if err := s.orderItemRepo.Create(ctx, orderItem); err != nil {
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}

		// Handle inventory updates
		if orderItem.Type == domain.OrderItemTypeSimple {
			// Deduct inventory for simple products
			inventoryEntry := &domain.InventoryLedger{
				ProductID: orderItem.ProductID,
				Delta:     -orderItem.Quantity,
				Reason:    domain.InventoryReasonSale,
			}
			if err := s.inventoryLedgerRepo.Create(ctx, inventoryEntry); err != nil {
				return nil, fmt.Errorf("failed to update inventory: %w", err)
			}

		} else if orderItem.Type == domain.OrderItemTypeBundle {
			// Create component items and deduct inventory
			bundleComponents, err := s.bundleComponentRepo.GetByBundleID(ctx, orderItem.ProductID)
			if err != nil {
				return nil, fmt.Errorf("failed to get bundle components: %w", err)
			}

			for _, component := range bundleComponents {
				componentProduct, err := s.productRepo.GetByID(ctx, component.ComponentProductID)
				if err != nil {
					return nil, fmt.Errorf("failed to get component product: %w", err)
				}

				componentQuantity := component.Quantity * orderItem.Quantity

				// Create component order item
				componentOrderItem := &domain.OrderItem{
					OrderID:      order.ID,
					ProductID:    component.ComponentProductID,
					ParentItemID: &orderItem.ID,
					Type:         domain.OrderItemTypeComponent,
					Title:        componentProduct.Name,
					Quantity:     componentQuantity,
					PricePerUnit: 0, // Component items don't have individual prices
				}

				if err := s.orderItemRepo.Create(ctx, componentOrderItem); err != nil {
					return nil, fmt.Errorf("failed to create component order item: %w", err)
				}

				// Deduct component inventory
				inventoryEntry := &domain.InventoryLedger{
					ProductID: component.ComponentProductID,
					Delta:     -componentQuantity,
					Reason:    domain.InventoryReasonSale,
				}
				if err := s.inventoryLedgerRepo.Create(ctx, inventoryEntry); err != nil {
					return nil, fmt.Errorf("failed to update component inventory: %w", err)
				}
			}
		}
	}

	return &CreateOrderResponse{
		Order:   s.toOrderDTO(order),
		Message: "Order created successfully",
		Success: true,
	}, nil
}

func (s *orderService) GetOrder(ctx context.Context, orderID string) (*GetOrderResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return nil, errors.New("invalid order ID format")
	}

	order, err := s.orderRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	items, err := s.orderItemRepo.GetByOrderID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}

	itemDTOs := make([]OrderItemDTO, len(items))
	for i, item := range items {
		itemDTOs[i] = s.toOrderItemDTO(item)
	}

	return &GetOrderResponse{
		Order: s.toOrderDTO(order),
		Items: itemDTOs,
	}, nil
}

func (s *orderService) UpdateOrder(ctx context.Context, orderID string, req UpdateOrderRequest) (*UpdateOrderResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return nil, errors.New("invalid order ID format")
	}

	order, err := s.orderRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	if order.Status != domain.OrderStatusPending {
		return nil, errors.New("can only update pending orders")
	}

	if req.ContactEmail != nil {
		order.ContactEmail = req.ContactEmail
	}

	// Handle items update if provided
	if req.Items != nil {
		// This would require reversing inventory changes and creating new ones
		// For now, we'll just update the contact email
		// Full item update implementation would be more complex
		return nil, errors.New("updating order items is not supported yet")
	}

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	return &UpdateOrderResponse{
		Order:   s.toOrderDTO(order),
		Message: "Order updated successfully",
		Success: true,
	}, nil
}

func (s *orderService) DeleteOrder(ctx context.Context, orderID string) (*DeleteOrderResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return nil, errors.New("invalid order ID format")
	}

	order, err := s.orderRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	if order.Status != domain.OrderStatusPending && order.Status != domain.OrderStatusCancelled {
		return nil, errors.New("can only delete pending or cancelled orders")
	}

	// Get order items to reverse inventory
	items, err := s.orderItemRepo.GetByOrderID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}

	// Reverse inventory changes if order was pending
	if order.Status == domain.OrderStatusPending {
		for _, item := range items {
			if item.Type == domain.OrderItemTypeSimple || item.Type == domain.OrderItemTypeComponent {
				inventoryEntry := &domain.InventoryLedger{
					ProductID: item.ProductID,
					Delta:     item.Quantity, // Positive to add back
					Reason:    domain.InventoryReasonRefund,
				}
				if err := s.inventoryLedgerRepo.Create(ctx, inventoryEntry); err != nil {
					return nil, fmt.Errorf("failed to reverse inventory: %w", err)
				}
			}
		}
	}

	// Delete order items
	if err := s.orderItemRepo.DeleteByOrderID(ctx, objectID); err != nil {
		return nil, fmt.Errorf("failed to delete order items: %w", err)
	}

	// Delete order
	if err := s.orderRepo.Delete(ctx, objectID); err != nil {
		return nil, fmt.Errorf("failed to delete order: %w", err)
	}

	return &DeleteOrderResponse{
		Message: "Order deleted successfully",
		Success: true,
	}, nil
}

func (s *orderService) ListOrders(ctx context.Context, req ListOrdersRequest) (*ListOrdersResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	var orders []*domain.Order
	var err error

	if req.Status != nil {
		orders, err = s.orderRepo.GetByStatus(ctx, *req.Status, limit, offset)
	} else {
		orders, err = s.orderRepo.List(ctx, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	orderDTOs := make([]OrderDTO, len(orders))
	for i, order := range orders {
		orderDTOs[i] = s.toOrderDTO(order)
	}

	return &ListOrdersResponse{
		Orders: orderDTOs,
		Total:  len(orderDTOs),
	}, nil
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, orderID string, status domain.OrderStatus) (*UpdateOrderStatusResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return nil, errors.New("invalid order ID format")
	}

	order, err := s.orderRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	if err := s.orderRepo.UpdateStatus(ctx, objectID, status); err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	return &UpdateOrderStatusResponse{
		Message: fmt.Sprintf("Order status updated to %s", status),
		Success: true,
	}, nil
}

func (s *orderService) GetOrdersByCustomer(ctx context.Context, customerID string, limit, offset int) (*ListOrdersResponse, error) {
	id, err := primitive.ObjectIDFromHex(customerID)
	if err != nil {
		return nil, errors.New("invalid customer ID format")
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	orders, err := s.orderRepo.GetByCustomerID(ctx, id, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by customer: %w", err)
	}

	orderDTOs := make([]OrderDTO, len(orders))
	for i, order := range orders {
		orderDTOs[i] = s.toOrderDTO(order)
	}

	return &ListOrdersResponse{
		Orders: orderDTOs,
		Total:  len(orderDTOs),
	}, nil
}

func (s *orderService) GetOrdersByEmail(ctx context.Context, email string, limit, offset int) (*ListOrdersResponse, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	orders, err := s.orderRepo.GetByContactEmail(ctx, email, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by email: %w", err)
	}

	orderDTOs := make([]OrderDTO, len(orders))
	for i, order := range orders {
		orderDTOs[i] = s.toOrderDTO(order)
	}

	return &ListOrdersResponse{
		Orders: orderDTOs,
		Total:  len(orderDTOs),
	}, nil
}

func (s *orderService) toOrderDTO(order *domain.Order) OrderDTO {
	dto := OrderDTO{
		ID:        order.ID.Hex(),
		Total:     order.Total,
		Status:    string(order.Status),
		CreatedAt: order.CreatedAt.Format(time.RFC3339),
		UpdatedAt: order.UpdatedAt.Format(time.RFC3339),
	}

	if order.CustomerID != nil {
		customerIDStr := order.CustomerID.Hex()
		dto.CustomerID = &customerIDStr
	}

	if order.ContactEmail != nil {
		dto.ContactEmail = order.ContactEmail
	}

	return dto
}

func (s *orderService) toOrderItemDTO(item *domain.OrderItem) OrderItemDTO {
	dto := OrderItemDTO{
		ID:           item.ID.Hex(),
		OrderID:      item.OrderID.Hex(),
		ProductID:    item.ProductID.Hex(),
		Type:         string(item.Type),
		Title:        item.Title,
		Quantity:     item.Quantity,
		PricePerUnit: item.PricePerUnit,
		IsRedeemed:   item.IsRedeemed,
	}

	if item.ParentItemID != nil {
		parentID := item.ParentItemID.Hex()
		dto.ParentItemID = &parentID
	}

	if item.RedeemedAt != nil {
		redeemedAt := item.RedeemedAt.Format(time.RFC3339)
		dto.RedeemedAt = &redeemedAt
	}

	if item.RedeemedStationID != nil {
		stationID := item.RedeemedStationID.Hex()
		dto.RedeemedStationID = &stationID
	}

	if item.RedeemedDeviceID != nil {
		deviceID := item.RedeemedDeviceID.Hex()
		dto.RedeemedDeviceID = &deviceID
	}

	return dto
}

func (s *orderService) GetOrderForRedemption(ctx context.Context, orderID, stationID string) (*OrderRedemptionResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return nil, errors.New("invalid order ID format")
	}

	stationObjectID, err := primitive.ObjectIDFromHex(stationID)
	if err != nil {
		return nil, errors.New("invalid station ID format")
	}

	// Verify station exists
	station, err := s.stationRepo.GetByID(ctx, stationObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get station: %w", err)
	}
	if station == nil {
		return nil, errors.New("station not found")
	}

	// Get the order
	order, err := s.orderRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	// Only allow redemption for paid orders
	if order.Status != domain.OrderStatusPaid {
		return nil, fmt.Errorf("order must be paid to redeem items. Current status: %s", order.Status)
	}

	// Get all order items
	allOrderItems, err := s.orderItemRepo.GetByOrderID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}

	// Get products available at this station
	stationProducts, err := s.stationProductRepo.GetByStationID(ctx, stationObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get station products: %w", err)
	}

	// Create a map of products available at this station
	stationProductMap := make(map[primitive.ObjectID]bool)
	for _, sp := range stationProducts {
		stationProductMap[sp.ProductID] = true
	}

	var redeemableItems []RedeemableOrderItemDTO
	var alreadyRedeemed []OrderItemDTO

	// Process order items to find what can be redeemed at this station
	for _, item := range allOrderItems {
		// Skip component items for now, we'll handle them through their parent bundles
		if item.Type == domain.OrderItemTypeComponent {
			continue
		}

		if item.IsRedeemed {
			alreadyRedeemed = append(alreadyRedeemed, s.toOrderItemDTO(item))
			continue
		}

		// Check if this product is available at this station
		if item.Type == domain.OrderItemTypeSimple {
			if stationProductMap[item.ProductID] {
				product, _ := s.productRepo.GetByID(ctx, item.ProductID)
				productName := item.Title
				if product != nil {
					productName = product.Name
				}

				redeemableItem := RedeemableOrderItemDTO{
					ID:           item.ID.Hex(),
					ProductID:    item.ProductID.Hex(),
					ProductName:  productName,
					Type:         string(item.Type),
					Quantity:     item.Quantity,
					PricePerUnit: item.PricePerUnit,
					IsBundle:     false,
				}
				redeemableItems = append(redeemableItems, redeemableItem)
			}
		} else if item.Type == domain.OrderItemTypeBundle {
			// For bundles, check if any components are available at this station
			componentItems, err := s.orderItemRepo.GetByParentItemID(ctx, item.ID)
			if err != nil {
				continue // Skip this bundle if we can't get components
			}

			var redeemableComponents []OrderItemDTO
			hasRedeemableComponents := false

			for _, component := range componentItems {
				if stationProductMap[component.ProductID] && !component.IsRedeemed {
					redeemableComponents = append(redeemableComponents, s.toOrderItemDTO(component))
					hasRedeemableComponents = true
				}
			}

			if hasRedeemableComponents {
				product, _ := s.productRepo.GetByID(ctx, item.ProductID)
				productName := item.Title
				if product != nil {
					productName = product.Name
				}

				redeemableItem := RedeemableOrderItemDTO{
					ID:             item.ID.Hex(),
					ProductID:      item.ProductID.Hex(),
					ProductName:    productName,
					Type:           string(item.Type),
					Quantity:       item.Quantity,
					PricePerUnit:   item.PricePerUnit,
					IsBundle:       true,
					ComponentItems: redeemableComponents,
				}
				redeemableItems = append(redeemableItems, redeemableItem)
			}
		}
	}

	message := "Order items available for redemption at this station"
	if len(redeemableItems) == 0 {
		if len(alreadyRedeemed) > 0 {
			message = "All items for this station have already been redeemed"
		} else {
			message = "No items available for redemption at this station"
		}
	}

	return &OrderRedemptionResponse{
		Order:           s.toOrderDTO(order),
		RedeemableItems: redeemableItems,
		AlreadyRedeemed: alreadyRedeemed,
		Message:         message,
	}, nil
}

func (s *orderService) RedeemOrderItems(ctx context.Context, req RedeemOrderItemsRequest) (*RedeemOrderItemsResponse, error) {
	orderObjectID, err := primitive.ObjectIDFromHex(req.OrderID)
	if err != nil {
		return nil, errors.New("invalid order ID format")
	}

	stationObjectID, err := primitive.ObjectIDFromHex(req.StationID)
	if err != nil {
		return nil, errors.New("invalid station ID format")
	}

	deviceObjectID, err := primitive.ObjectIDFromHex(req.DeviceID)
	if err != nil {
		return nil, errors.New("invalid device ID format")
	}

	// Verify device exists and belongs to the station
	device, err := s.deviceRepo.GetByID(ctx, deviceObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	if device == nil {
		return nil, errors.New("device not found")
	}
	if device.StationID != stationObjectID {
		return nil, errors.New("device does not belong to the specified station")
	}
	if !device.IsActive {
		return nil, errors.New("device is not active")
	}

	// Get the order
	order, err := s.orderRepo.GetByID(ctx, orderObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	// Only allow redemption for paid orders
	if order.Status != domain.OrderStatusPaid {
		return nil, fmt.Errorf("order must be paid to redeem items. Current status: %s", order.Status)
	}

	// Get products available at this station
	stationProducts, err := s.stationProductRepo.GetByStationID(ctx, stationObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get station products: %w", err)
	}

	// Create a map of products available at this station
	stationProductMap := make(map[primitive.ObjectID]bool)
	for _, sp := range stationProducts {
		stationProductMap[sp.ProductID] = true
	}

	// Get all order items
	allOrderItems, err := s.orderItemRepo.GetByOrderID(ctx, orderObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}

	var redeemedItems []OrderItemDTO
	itemsToRedeem := make([]primitive.ObjectID, 0)

	// Find items that can be redeemed at this station
	for _, item := range allOrderItems {
		// Skip already redeemed items
		if item.IsRedeemed {
			continue
		}

		// Handle simple products
		if item.Type == domain.OrderItemTypeSimple {
			if stationProductMap[item.ProductID] {
				itemsToRedeem = append(itemsToRedeem, item.ID)
				redeemedItems = append(redeemedItems, s.toOrderItemDTO(item))
			}
		}
		// Handle bundle components
		if item.Type == domain.OrderItemTypeComponent {
			if stationProductMap[item.ProductID] {
				itemsToRedeem = append(itemsToRedeem, item.ID)
				redeemedItems = append(redeemedItems, s.toOrderItemDTO(item))
			}
		}
	}

	// Check if any items were found for redemption
	if len(itemsToRedeem) == 0 {
		return nil, errors.New("no items available for redemption at this station or all items already redeemed")
	}

	// Mark items as redeemed
	for _, itemID := range itemsToRedeem {
		if err := s.orderItemRepo.MarkAsRedeemed(ctx, itemID, stationObjectID, deviceObjectID); err != nil {
			return nil, fmt.Errorf("failed to mark item as redeemed: %w", err)
		}
	}

	// Check if all order items are now redeemed
	remainingItems, err := s.orderItemRepo.GetUnredeemedByOrderID(ctx, orderObjectID)
	if err == nil && len(remainingItems) == 0 {
		// All items redeemed, we could potentially update order status here
		// For now, we'll leave the order as "paid" since that indicates it's complete
	}

	return &RedeemOrderItemsResponse{
		RedeemedItems: redeemedItems,
		Message:       fmt.Sprintf("Successfully redeemed %d items at this station", len(redeemedItems)),
		Success:       true,
	}, nil
}