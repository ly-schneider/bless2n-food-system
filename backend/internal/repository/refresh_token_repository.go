package repository

import (
	"backend/internal/domain"
	"context"
	"fmt"

	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *domain.RefreshToken) error
}

type refreshTokenRepo struct{ db *gorm.DB }

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepo{db}
}

func (r *refreshTokenRepo) Create(ctx context.Context, rt *domain.RefreshToken) error {
	if err := r.db.WithContext(ctx).Create(rt).Error; err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}
	return nil
}
