package postgres

import (
	"context"

	"backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrderPaymentRepository defines the interface for order payment data access.
type OrderPaymentRepository interface {
	Create(ctx context.Context, payment *model.OrderPayment) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.OrderPayment, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]model.OrderPayment, error)
	Update(ctx context.Context, payment *model.OrderPayment) error
}

type orderPaymentRepo struct {
	db *gorm.DB
}

// NewOrderPaymentRepository creates a new OrderPaymentRepository.
func NewOrderPaymentRepository(db *gorm.DB) OrderPaymentRepository {
	return &orderPaymentRepo{db: db}
}

func (r *orderPaymentRepo) Create(ctx context.Context, payment *model.OrderPayment) error {
	return translateError(r.db.WithContext(ctx).Create(payment).Error)
}

func (r *orderPaymentRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.OrderPayment, error) {
	var payment model.OrderPayment
	err := r.db.WithContext(ctx).First(&payment, "id = ?", id).Error
	return &payment, translateError(err)
}

func (r *orderPaymentRepo) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]model.OrderPayment, error) {
	var payments []model.OrderPayment
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("paid_at ASC").
		Find(&payments).Error
	return payments, translateError(err)
}

func (r *orderPaymentRepo) Update(ctx context.Context, payment *model.OrderPayment) error {
	return translateError(r.db.WithContext(ctx).Save(payment).Error)
}
