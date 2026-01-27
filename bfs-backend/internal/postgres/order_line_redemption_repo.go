package postgres

import (
	"context"

	"backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrderLineRedemptionRepository defines the interface for order line redemption data access.
type OrderLineRedemptionRepository interface {
	Create(ctx context.Context, redemption *model.OrderLineRedemption) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.OrderLineRedemption, error)
	GetByOrderLineID(ctx context.Context, orderLineID uuid.UUID) (*model.OrderLineRedemption, error)
	ExistsByOrderLineID(ctx context.Context, orderLineID uuid.UUID) (bool, error)
}

type orderLineRedemptionRepo struct {
	db *gorm.DB
}

// NewOrderLineRedemptionRepository creates a new OrderLineRedemptionRepository.
func NewOrderLineRedemptionRepository(db *gorm.DB) OrderLineRedemptionRepository {
	return &orderLineRedemptionRepo{db: db}
}

func (r *orderLineRedemptionRepo) Create(ctx context.Context, redemption *model.OrderLineRedemption) error {
	return translateError(r.db.WithContext(ctx).Create(redemption).Error)
}

func (r *orderLineRedemptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.OrderLineRedemption, error) {
	var redemption model.OrderLineRedemption
	err := r.db.WithContext(ctx).First(&redemption, "id = ?", id).Error
	return &redemption, translateError(err)
}

func (r *orderLineRedemptionRepo) GetByOrderLineID(ctx context.Context, orderLineID uuid.UUID) (*model.OrderLineRedemption, error) {
	var redemption model.OrderLineRedemption
	err := r.db.WithContext(ctx).First(&redemption, "order_line_id = ?", orderLineID).Error
	return &redemption, translateError(err)
}

func (r *orderLineRedemptionRepo) ExistsByOrderLineID(ctx context.Context, orderLineID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.OrderLineRedemption{}).
		Where("order_line_id = ?", orderLineID).
		Count(&count).Error
	return count > 0, translateError(err)
}
