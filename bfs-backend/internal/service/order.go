package service

import (
	"context"
	"fmt"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/inventoryledger"
	"backend/internal/generated/ent/order"
	"backend/internal/generated/ent/product"
	"backend/internal/inventory"
	"backend/internal/repository"

	"github.com/google/uuid"
)

type OrderService interface {
	// GetByID retrieves an order by ID.
	GetByID(ctx context.Context, id uuid.UUID) (*ent.Order, error)
	// GetByIDWithRelations retrieves an order by ID with lines, payments, and redemptions eagerly loaded.
	GetByIDWithRelations(ctx context.Context, id uuid.UUID) (*ent.Order, error)
	// GetOrderLines retrieves all order lines for an order.
	GetOrderLines(ctx context.Context, orderID uuid.UUID) ([]*ent.OrderLine, error)
	ListByCustomerID(ctx context.Context, customerID string) ([]*ent.Order, int64, error)
	ListAdmin(ctx context.Context, params OrderListParams) ([]*ent.Order, int64, error)
	// UpdateStatus updates an order's status.
	UpdateStatus(ctx context.Context, id uuid.UUID, status order.Status) error
	// ListEvents returns months with paid orders for dashboard navigation.
	ListEvents(ctx context.Context) ([]repository.EventMonth, error)
}

type OrderListParams struct {
	Status *order.Status
	From   *string // RFC3339 timestamp
	To     *string // RFC3339 timestamp
	Query  *string
}

type orderService struct {
	orderRepo     repository.OrderRepository
	orderLineRepo repository.OrderLineRepository
	inventoryRepo repository.InventoryLedgerRepository
	inventoryHub  *inventory.Hub
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	orderLineRepo repository.OrderLineRepository,
	inventoryRepo repository.InventoryLedgerRepository,
	inventoryHub *inventory.Hub,
) OrderService {
	return &orderService{
		orderRepo:     orderRepo,
		orderLineRepo: orderLineRepo,
		inventoryRepo: inventoryRepo,
		inventoryHub:  inventoryHub,
	}
}

func (s *orderService) GetByID(ctx context.Context, id uuid.UUID) (*ent.Order, error) {
	return s.orderRepo.GetByID(ctx, id)
}

func (s *orderService) GetByIDWithRelations(ctx context.Context, id uuid.UUID) (*ent.Order, error) {
	return s.orderRepo.GetByIDWithRelations(ctx, id)
}

func (s *orderService) GetOrderLines(ctx context.Context, orderID uuid.UUID) ([]*ent.OrderLine, error) {
	return s.orderLineRepo.GetByOrderID(ctx, orderID)
}

func (s *orderService) ListByCustomerID(ctx context.Context, customerID string) ([]*ent.Order, int64, error) {
	return s.orderRepo.ListByCustomerIDPaginated(ctx, customerID)
}

func (s *orderService) ListAdmin(ctx context.Context, params OrderListParams) ([]*ent.Order, int64, error) {
	// Parse time parameters if provided
	var from, to *time.Time
	if params.From != nil && *params.From != "" {
		t, err := time.Parse(time.RFC3339, *params.From)
		if err == nil {
			from = &t
		}
	}
	if params.To != nil && *params.To != "" {
		t, err := time.Parse(time.RFC3339, *params.To)
		if err == nil {
			to = &t
		}
	}

	return s.orderRepo.ListAdmin(ctx, params.Status, from, to, params.Query)
}

func (s *orderService) UpdateStatus(ctx context.Context, id uuid.UUID, status order.Status) error {
	ord, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !isValidStatusTransition(ord.Status, status) {
		return fmt.Errorf("invalid status transition from %s to %s", ord.Status, status)
	}

	if err := s.orderRepo.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	switch status {
	case order.StatusRefunded:
		if err := s.restoreInventory(ctx, id, inventoryledger.ReasonRefund); err != nil {
			return fmt.Errorf("inventory restore failed: %w", err)
		}
	case order.StatusCancelled:
		if err := s.restoreInventory(ctx, id, inventoryledger.ReasonCancellation); err != nil {
			return fmt.Errorf("inventory restore failed: %w", err)
		}
	}

	return nil
}

func (s *orderService) restoreInventory(ctx context.Context, orderID uuid.UUID, reason inventoryledger.Reason) error {
	lines, err := s.orderLineRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	var entries []repository.InventoryLedgerCreateParams
	for _, line := range lines {
		if line.Quantity <= 0 {
			continue
		}
		if line.Edges.Product == nil || line.Edges.Product.Type != product.TypeSimple {
			continue
		}
		entries = append(entries, repository.InventoryLedgerCreateParams{
			ProductID:   line.ProductID,
			Delta:       line.Quantity,
			Reason:      reason,
			OrderID:     &orderID,
			OrderLineID: &line.ID,
		})
	}

	if len(entries) > 0 {
		if _, err := s.inventoryRepo.CreateMany(ctx, entries); err != nil {
			return err
		}
		s.publishInventoryUpdates(ctx, entries)
	}
	return nil
}

func (s *orderService) publishInventoryUpdates(ctx context.Context, entries []repository.InventoryLedgerCreateParams) {
	if s.inventoryHub == nil {
		return
	}
	productIDs := make([]uuid.UUID, 0, len(entries))
	deltaByProduct := make(map[uuid.UUID]int)
	for _, entry := range entries {
		if _, seen := deltaByProduct[entry.ProductID]; !seen {
			productIDs = append(productIDs, entry.ProductID)
		}
		deltaByProduct[entry.ProductID] += entry.Delta
	}
	stocks, err := s.inventoryRepo.GetCurrentStockBatch(ctx, productIDs)
	if err != nil {
		return
	}
	now := time.Now()
	for _, productID := range productIDs {
		s.inventoryHub.Publish(inventory.Update{
			ProductID: productID,
			NewStock:  stocks[productID],
			Delta:     deltaByProduct[productID],
			Timestamp: now,
		})
	}
}

func isValidStatusTransition(from, to order.Status) bool {
	transitions := map[order.Status][]order.Status{
		order.StatusPending: {
			order.StatusPaid,
			order.StatusCancelled,
		},
		order.StatusPaid: {
			order.StatusCancelled,
			order.StatusRefunded,
		},
		order.StatusCancelled: {
			// Terminal state
		},
		order.StatusRefunded: {
			// Terminal state
		},
	}

	allowed, ok := transitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

func (s *orderService) ListEvents(ctx context.Context) ([]repository.EventMonth, error) {
	return s.orderRepo.GetEventMonths(ctx)
}
