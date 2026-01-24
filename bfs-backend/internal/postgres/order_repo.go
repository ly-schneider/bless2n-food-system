package postgres

import (
	"context"
	"time"

	"backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrderRepository defines the interface for order data access.
type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error)
	GetByIDWithRelations(ctx context.Context, id uuid.UUID) (*model.Order, error)
	GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]model.Order, error)
	GetByStatus(ctx context.Context, status model.OrderStatus) ([]model.Order, error)
	GetByDateRange(ctx context.Context, start, end time.Time) ([]model.Order, error)
	GetRecent(ctx context.Context, limit int) ([]model.Order, error)
	Update(ctx context.Context, order *model.Order) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus) error
}

type orderRepo struct {
	db *gorm.DB
}

// NewOrderRepository creates a new OrderRepository.
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepo{db: db}
}

func (r *orderRepo) Create(ctx context.Context, order *model.Order) error {
	return translateError(r.db.WithContext(ctx).Create(order).Error)
}

func (r *orderRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).First(&order, "id = ?", id).Error
	return &order, translateError(err)
}

func (r *orderRepo) GetByIDWithRelations(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).
		Preload("Payments").
		Preload("Lines").
		Preload("Lines.Product").
		Preload("Lines.ChildLines").
		Preload("Lines.Redemption").
		First(&order, "id = ?", id).Error
	return &order, translateError(err)
}

func (r *orderRepo) GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.WithContext(ctx).
		Preload("Lines").
		Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, translateError(err)
}

func (r *orderRepo) GetByStatus(ctx context.Context, status model.OrderStatus) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, translateError(err)
}

func (r *orderRepo) GetByDateRange(ctx context.Context, start, end time.Time) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.WithContext(ctx).
		Where("created_at BETWEEN ? AND ?", start, end).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, translateError(err)
}

func (r *orderRepo) GetRecent(ctx context.Context, limit int) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Find(&orders).Error
	return orders, translateError(err)
}

func (r *orderRepo) Update(ctx context.Context, order *model.Order) error {
	return translateError(r.db.WithContext(ctx).Save(order).Error)
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus) error {
	return translateError(r.db.WithContext(ctx).
		Model(&model.Order{}).
		Where("id = ?", id).
		Update("status", status).Error)
}
