package repository

import (
	"backend/internal/domain"
	"backend/internal/logger"
	"backend/internal/utils"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	GetByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	GetValidTokenForUser(ctx context.Context, plainToken string) (*domain.RefreshToken, error)
	Create(ctx context.Context, rt *domain.RefreshToken) error
	RevokeByHash(ctx context.Context, tokenHash string) error
}

type refreshTokenRepo struct{ db *gorm.DB }

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepo{db}
}

func (r *refreshTokenRepo) GetByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var rt domain.RefreshToken
	if err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&rt).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.L.Error(ctx, "refresh token not found for token_hash: ", tokenHash)
			return nil, domain.ErrRefreshTokenNotFound
		}
		return nil, domain.ErrRefreshTokenNotFound
	}

	if rt.IsRevoked {
		return nil, domain.ErrRefreshTokenRevoked
	}

	if time.Now().After(rt.ExpiresAt) {
		return nil, domain.ErrRefreshTokenExpired
	}

	return &rt, nil
}

func (r *refreshTokenRepo) Create(ctx context.Context, rt *domain.RefreshToken) error {
	if err := r.db.WithContext(ctx).Create(rt).Error; err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}
	return nil
}

func (r *refreshTokenRepo) GetValidTokenForUser(ctx context.Context, plainToken string) (*domain.RefreshToken, error) {
	var tokens []domain.RefreshToken
	if err := r.db.WithContext(ctx).
		Where("is_revoked = ? AND expires_at > ?", false, time.Now()).
		Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("failed to get active tokens: %w", err)
	}

	// Verify the plain token against each stored hash
	for _, token := range tokens {
		if utils.VerifyToken(plainToken, token.TokenHash) {
			return &token, nil
		}
	}

	return nil, domain.ErrRefreshTokenNotFound
}

func (r *refreshTokenRepo) RevokeByHash(ctx context.Context, tokenHash string) error {
	if err := r.db.WithContext(ctx).Model(&domain.RefreshToken{}).
		Where("token_hash = ?", tokenHash).
		Update("is_revoked", true).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.L.Error(ctx, "refresh token not found for token_hash: ", tokenHash)
			return domain.ErrRefreshTokenNotFound
		}
		return domain.ErrRefreshTokenNotFound
	}
	return nil
}
