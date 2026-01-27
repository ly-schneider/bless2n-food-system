package postgres

import (
	"context"

	"backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MenuSlotRepository defines the interface for menu slot data access.
type MenuSlotRepository interface {
	Create(ctx context.Context, slot *model.MenuSlot) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.MenuSlot, error)
	GetByMenuProductID(ctx context.Context, menuProductID uuid.UUID) ([]model.MenuSlot, error)
	Update(ctx context.Context, slot *model.MenuSlot) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByMenuProductID(ctx context.Context, menuProductID uuid.UUID) error
}

type menuSlotRepo struct {
	db *gorm.DB
}

// NewMenuSlotRepository creates a new MenuSlotRepository.
func NewMenuSlotRepository(db *gorm.DB) MenuSlotRepository {
	return &menuSlotRepo{db: db}
}

func (r *menuSlotRepo) Create(ctx context.Context, slot *model.MenuSlot) error {
	return translateError(r.db.WithContext(ctx).Create(slot).Error)
}

func (r *menuSlotRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.MenuSlot, error) {
	var slot model.MenuSlot
	err := r.db.WithContext(ctx).
		Preload("Options").
		Preload("Options.OptionProduct").
		First(&slot, "id = ?", id).Error
	return &slot, translateError(err)
}

func (r *menuSlotRepo) GetByMenuProductID(ctx context.Context, menuProductID uuid.UUID) ([]model.MenuSlot, error) {
	var slots []model.MenuSlot
	err := r.db.WithContext(ctx).
		Preload("Options").
		Preload("Options.OptionProduct").
		Where("menu_product_id = ?", menuProductID).
		Order("sequence ASC").
		Find(&slots).Error
	return slots, translateError(err)
}

func (r *menuSlotRepo) Update(ctx context.Context, slot *model.MenuSlot) error {
	return translateError(r.db.WithContext(ctx).Save(slot).Error)
}

func (r *menuSlotRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return translateError(r.db.WithContext(ctx).Delete(&model.MenuSlot{}, "id = ?", id).Error)
}

func (r *menuSlotRepo) DeleteByMenuProductID(ctx context.Context, menuProductID uuid.UUID) error {
	return translateError(r.db.WithContext(ctx).Delete(&model.MenuSlot{}, "menu_product_id = ?", menuProductID).Error)
}
