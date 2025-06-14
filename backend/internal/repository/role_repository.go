package repository

import (
	"backend/internal/domain"
	"context"

	"gorm.io/gorm"
)

type RoleRepository interface {
	List(ctx context.Context) ([]domain.Role, error)
	Get(ctx context.Context, id string) (domain.Role, error)
	GetByName(ctx context.Context, name string) (domain.Role, error)
	Create(ctx context.Context, r *domain.Role) error
	Update(ctx context.Context, r *domain.Role) error
	Delete(ctx context.Context, id string) error
}

type roleRepo struct{ db *gorm.DB }

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepo{db}
}

func (r *roleRepo) List(ctx context.Context) ([]domain.Role, error) {
	var out []domain.Role
	return out, r.db.WithContext(ctx).Find(&out).Error
}

func (r *roleRepo) Get(ctx context.Context, id string) (domain.Role, error) {
	var role domain.Role
	return role, r.db.WithContext(ctx).First(&role, "id = ?", id).Error
}

func (r *roleRepo) GetByName(ctx context.Context, name string) (domain.Role, error) {
	var role domain.Role
	return role, r.db.WithContext(ctx).First(&role, "name = ?", name).Error
}

func (r *roleRepo) Create(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepo) Update(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Role{}, "id = ?", id).Error
}
