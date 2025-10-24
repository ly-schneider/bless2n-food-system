package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrderService interface {
	PrepareOrder(ctx context.Context, dto *domain.CreateOrderDTO, userID *bson.ObjectID) (*domain.Order, []*domain.OrderItem, error)
}

type orderService struct {
	productRepo repository.ProductRepository
}

func NewOrderService(productRepo repository.ProductRepository) OrderService {
	return &orderService{productRepo: productRepo}
}

// PrepareOrder validates products and prices, and builds order + order items.
// For menu configurations, component items are stored as children with zero price.
func (s *orderService) PrepareOrder(ctx context.Context, dto *domain.CreateOrderDTO, userID *bson.ObjectID) (*domain.Order, []*domain.OrderItem, error) {
	if dto == nil || len(dto.OrderItems) == 0 {
		return nil, nil, fmt.Errorf("no items provided")
	}

	// Collect all product IDs (parents + configured children)
	productIDSet := make(map[bson.ObjectID]struct{})
	for _, it := range dto.OrderItems {
		productIDSet[it.ProductID] = struct{}{}
	}
	ids := make([]bson.ObjectID, 0, len(productIDSet))
	for id := range productIDSet {
		ids = append(ids, id)
	}

	products, err := s.productRepo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, nil, fmt.Errorf("load products: %w", err)
	}
	prodMap := make(map[bson.ObjectID]*domain.Product, len(products))
	for _, p := range products {
		prodMap[p.ID] = p
	}

	var total domain.Cents
	now := time.Now().UTC()
	orderItems := make([]*domain.OrderItem, 0, len(dto.OrderItems))

	for _, it := range dto.OrderItems {
		p, ok := prodMap[it.ProductID]
		if !ok {
			return nil, nil, fmt.Errorf("unknown product %s", it.ProductID.Hex())
		}
		// Parent item
		parentID := bson.NewObjectID()
		orderItems = append(orderItems, &domain.OrderItem{
			ID:                parentID,
			OrderID:           bson.NilObjectID, // fill after order insert
			ProductID:         p.ID,
			Title:             p.Name,
			Quantity:          it.Quantity,
			PricePerUnitCents: p.PriceCents,
			IsRedeemed:        false,
			RedeemedAt:        nil,
		})
		total += p.PriceCents * domain.Cents(it.Quantity)

		// Configuration items (recorded as zero-price children just for fulfillment)
		if it.MenuSlotItem != nil {
			// We only store a single configured item here if provided
			childID := bson.NewObjectID()
			orderItems = append(orderItems, &domain.OrderItem{
				ID:                childID,
				OrderID:           bson.NilObjectID,
				ProductID:         bson.NilObjectID, // unknown without full menu lookup
				Title:             "Configured Item",
				Quantity:          it.Quantity,
				PricePerUnitCents: 0,
				ParentItemID:      &parentID,
				MenuSlotID:        nil,
				MenuSlotName:      nil,
				IsRedeemed:        false,
				RedeemedAt:        nil,
			})
		}
	}

	ord := &domain.Order{
		ID:           bson.NewObjectID(),
		CustomerID:   userID,
		ContactEmail: dto.ContactEmail,
		TotalCents:   total,
		Status:       domain.OrderStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return ord, orderItems, nil
}
