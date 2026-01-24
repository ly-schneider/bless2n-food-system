package postgres

import (
	"context"

	"backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MenuSlotOptionRepository defines the interface for menu slot option data access.
type MenuSlotOptionRepository interface {
	Create(ctx context.Context, option *model.MenuSlotOption) error
	CreateBatch(ctx context.Context, options []model.MenuSlotOption) error
	GetByMenuSlotID(ctx context.Context, menuSlotID uuid.UUID) ([]model.MenuSlotOption, error)
	Delete(ctx context.Context, menuSlotID, optionProductID uuid.UUID) error
	DeleteByMenuSlotID(ctx context.Context, menuSlotID uuid.UUID) error
}

type menuSlotOptionRepo struct {
	db *gorm.DB
}

// NewMenuSlotOptionRepository creates a new MenuSlotOptionRepository.
func NewMenuSlotOptionRepository(db *gorm.DB) MenuSlotOptionRepository {
	return &menuSlotOptionRepo{db: db}
}

func (r *menuSlotOptionRepo) Create(ctx context.Context, option *model.MenuSlotOption) error {
	return translateError(r.db.WithContext(ctx).Create(option).Error)
}

func (r *menuSlotOptionRepo) CreateBatch(ctx context.Context, options []model.MenuSlotOption) error {
	if len(options) == 0 {
		return nil
	}
	return translateError(r.db.WithContext(ctx).Create(&options).Error)
}

func (r *menuSlotOptionRepo) GetByMenuSlotID(ctx context.Context, menuSlotID uuid.UUID) ([]model.MenuSlotOption, error) {
	var options []model.MenuSlotOption
	err := r.db.WithContext(ctx).
		Preload("OptionProduct").
		Where("menu_slot_id = ?", menuSlotID).
		Find(&options).Error
	return options, translateError(err)
}

func (r *menuSlotOptionRepo) Delete(ctx context.Context, menuSlotID, optionProductID uuid.UUID) error {
	return translateError(r.db.WithContext(ctx).Delete(&model.MenuSlotOption{}, "menu_slot_id = ? AND option_product_id = ?", menuSlotID, optionProductID).Error)
}

func (r *menuSlotOptionRepo) DeleteByMenuSlotID(ctx context.Context, menuSlotID uuid.UUID) error {
	return translateError(r.db.WithContext(ctx).Delete(&model.MenuSlotOption{}, "menu_slot_id = ?", menuSlotID).Error)
}
