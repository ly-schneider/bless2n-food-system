package repository

import (
	"backend/internal/domain"
	"context"

	"gorm.io/gorm"
)

type ProductRepository interface {
	List(ctx context.Context) ([]domain.Product, error)
	Get(ctx context.Context, id uint) (domain.Product, error)
	Create(ctx context.Context, p *domain.Product) error
	Update(ctx context.Context, p *domain.Product) error
	Delete(ctx context.Context, id uint) error
}

type productRepo struct{ db *gorm.DB }

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepo{db}
}

func (r *productRepo) List(ctx context.Context) ([]domain.Product, error) {
	var out []domain.Product
	return out, r.db.WithContext(ctx).Find(&out).Error
}

func (r *productRepo) Get(ctx context.Context, id uint) (domain.Product, error) {
	var p domain.Product
	return p, r.db.WithContext(ctx).First(&p, id).Error
}

func (r *productRepo) Create(ctx context.Context, p *domain.Product) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *productRepo) Update(ctx context.Context, p *domain.Product) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *productRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Product{}, id).Error
}
