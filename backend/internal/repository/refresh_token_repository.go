package repository

import (
	"backend/internal/domain"
	"context"
	"errors"

	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	GetByTokenHash(ctx context.Context, tokenHash string) (domain.RefreshToken, error)
	Create(ctx context.Context, rt *domain.RefreshToken) error
	Revoke(ctx context.Context, id string) error
}

type refreshTokenRepo struct{ db *gorm.DB }

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepo{db}
}

func (r *refreshTokenRepo) GetByTokenHash(ctx context.Context, tokenHash string) (domain.RefreshToken, error) {
	var rt domain.RefreshToken
	err := r.db.WithContext(ctx).First(&rt, "token_hash = ?", tokenHash).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		// Return empty refresh token with no error when record not found
		return domain.RefreshToken{}, nil
	}
	return rt, err
}

func (r *refreshTokenRepo) Create(ctx context.Context, rt *domain.RefreshToken) error {
	return r.db.WithContext(ctx).Create(rt).Error
}

func (r *refreshTokenRepo) Revoke(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.RefreshToken{}).Where("id = ?", id).Update("revoked", true).Error
}
