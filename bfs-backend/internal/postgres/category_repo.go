package postgres

import (
	"context"

	"backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CategoryRepository defines the interface for category data access.
type CategoryRepository interface {
	Create(ctx context.Context, category *model.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Category, error)
	GetAll(ctx context.Context) ([]model.Category, error)
	GetAllActive(ctx context.Context) ([]model.Category, error)
	Update(ctx context.Context, category *model.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type categoryRepo struct {
	db *gorm.DB
}

// NewCategoryRepository creates a new CategoryRepository.
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepo{db: db}
}

func (r *categoryRepo) Create(ctx context.Context, category *model.Category) error {
	return translateError(r.db.WithContext(ctx).Create(category).Error)
}

func (r *categoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	var category model.Category
	err := r.db.WithContext(ctx).First(&category, "id = ?", id).Error
	return &category, translateError(err)
}

func (r *categoryRepo) GetAll(ctx context.Context) ([]model.Category, error) {
	var categories []model.Category
	err := r.db.WithContext(ctx).Order("position ASC").Find(&categories).Error
	return categories, translateError(err)
}

func (r *categoryRepo) GetAllActive(ctx context.Context) ([]model.Category, error) {
	var categories []model.Category
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("position ASC").Find(&categories).Error
	return categories, translateError(err)
}

func (r *categoryRepo) Update(ctx context.Context, category *model.Category) error {
	return translateError(r.db.WithContext(ctx).Save(category).Error)
}

func (r *categoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return translateError(r.db.WithContext(ctx).Delete(&model.Category{}, "id = ?", id).Error)
}
