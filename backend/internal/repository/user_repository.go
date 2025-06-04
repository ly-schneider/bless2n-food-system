package repository

import (
	"backend/internal/domain"
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	List(ctx context.Context) ([]domain.User, error)
	Get(ctx context.Context, id uuid.UUID) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Create(ctx context.Context, u *domain.User) error
	Update(ctx context.Context, u *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type userRepo struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db}
}

func (r *userRepo) List(ctx context.Context) ([]domain.User, error) {
	var out []domain.User
	return out, r.db.WithContext(ctx).Preload("Role").Find(&out).Error
}

func (r *userRepo) Get(ctx context.Context, id uuid.UUID) (domain.User, error) {
	var u domain.User
	return u, r.db.WithContext(ctx).Preload("Role").First(&u, "id = ?", id).Error
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).Preload("Role").First(&u, "email = ?", email).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		// Return empty user with no error when record not found
		return domain.User{}, nil
	}
	return u, err
}

func (r *userRepo) Create(ctx context.Context, u *domain.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *userRepo) Update(ctx context.Context, u *domain.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *userRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.User{}, "id = ?", id).Error
}
