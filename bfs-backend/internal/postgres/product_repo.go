package postgres

import (
	"context"

	"backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductRepository defines the interface for product data access.
type ProductRepository interface {
	Create(ctx context.Context, product *model.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error)
	GetByIDWithRelations(ctx context.Context, id uuid.UUID) (*model.Product, error)
	GetAll(ctx context.Context) ([]model.Product, error)
	GetAllActive(ctx context.Context) ([]model.Product, error)
	GetByCategory(ctx context.Context, categoryID uuid.UUID) ([]model.Product, error)
	GetByCategoryActive(ctx context.Context, categoryID uuid.UUID) ([]model.Product, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Product, error)
	Update(ctx context.Context, product *model.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type productRepo struct {
	db *gorm.DB
}

// NewProductRepository creates a new ProductRepository.
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepo{db: db}
}

func (r *productRepo) Create(ctx context.Context, product *model.Product) error {
	return translateError(r.db.WithContext(ctx).Create(product).Error)
}

func (r *productRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	var product model.Product
	err := r.db.WithContext(ctx).First(&product, "id = ?", id).Error
	return &product, translateError(err)
}

func (r *productRepo) GetByIDWithRelations(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	var product model.Product
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Jeton").
		Preload("MenuSlots").
		Preload("MenuSlots.Options").
		Preload("MenuSlots.Options.OptionProduct").
		First(&product, "id = ?", id).Error
	return &product, translateError(err)
}

func (r *productRepo) GetAll(ctx context.Context) ([]model.Product, error) {
	var products []model.Product
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Jeton").
		Order("name ASC").
		Find(&products).Error
	return products, translateError(err)
}

func (r *productRepo) GetAllActive(ctx context.Context) ([]model.Product, error) {
	var products []model.Product
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Jeton").
		Where("is_active = ?", true).
		Order("name ASC").
		Find(&products).Error
	return products, translateError(err)
}

func (r *productRepo) GetByCategory(ctx context.Context, categoryID uuid.UUID) ([]model.Product, error) {
	var products []model.Product
	err := r.db.WithContext(ctx).
		Preload("Jeton").
		Where("category_id = ?", categoryID).
		Order("name ASC").
		Find(&products).Error
	return products, translateError(err)
}

func (r *productRepo) GetByCategoryActive(ctx context.Context, categoryID uuid.UUID) ([]model.Product, error) {
	var products []model.Product
	err := r.db.WithContext(ctx).
		Preload("Jeton").
		Where("category_id = ? AND is_active = ?", categoryID, true).
		Order("name ASC").
		Find(&products).Error
	return products, translateError(err)
}

func (r *productRepo) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Product, error) {
	var products []model.Product
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Jeton").
		Where("id IN ?", ids).
		Find(&products).Error
	return products, translateError(err)
}

func (r *productRepo) Update(ctx context.Context, product *model.Product) error {
	return translateError(r.db.WithContext(ctx).Save(product).Error)
}

func (r *productRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return translateError(r.db.WithContext(ctx).Delete(&model.Product{}, "id = ?", id).Error)
}
