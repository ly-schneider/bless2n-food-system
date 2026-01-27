package postgres

import (
	"context"

	"backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrderLineRepository defines the interface for order line data access.
type OrderLineRepository interface {
	Create(ctx context.Context, line *model.OrderLine) error
	CreateBatch(ctx context.Context, lines []model.OrderLine) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.OrderLine, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]model.OrderLine, error)
	GetUnredeemed(ctx context.Context, orderID uuid.UUID) ([]model.OrderLine, error)
	Update(ctx context.Context, line *model.OrderLine) error
}

type orderLineRepo struct {
	db *gorm.DB
}

// NewOrderLineRepository creates a new OrderLineRepository.
func NewOrderLineRepository(db *gorm.DB) OrderLineRepository {
	return &orderLineRepo{db: db}
}

func (r *orderLineRepo) Create(ctx context.Context, line *model.OrderLine) error {
	return translateError(r.db.WithContext(ctx).Create(line).Error)
}

func (r *orderLineRepo) CreateBatch(ctx context.Context, lines []model.OrderLine) error {
	if len(lines) == 0 {
		return nil
	}
	return translateError(r.db.WithContext(ctx).Create(&lines).Error)
}

func (r *orderLineRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.OrderLine, error) {
	var line model.OrderLine
	err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("Redemption").
		First(&line, "id = ?", id).Error
	return &line, translateError(err)
}

func (r *orderLineRepo) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]model.OrderLine, error) {
	var lines []model.OrderLine
	err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("Redemption").
		Where("order_id = ?", orderID).
		Find(&lines).Error
	return lines, translateError(err)
}

func (r *orderLineRepo) GetUnredeemed(ctx context.Context, orderID uuid.UUID) ([]model.OrderLine, error) {
	var lines []model.OrderLine
	err := r.db.WithContext(ctx).
		Preload("Product").
		Joins("LEFT JOIN app.order_line_redemption r ON r.order_line_id = app.order_line.id").
		Where("app.order_line.order_id = ? AND r.id IS NULL", orderID).
		Find(&lines).Error
	return lines, translateError(err)
}

func (r *orderLineRepo) Update(ctx context.Context, line *model.OrderLine) error {
	return translateError(r.db.WithContext(ctx).Save(line).Error)
}
