package repository

import (
	"backend/internal/domain"
	"backend/internal/model"
	"context"
	"time"

	"gorm.io/gorm"
)

type VerificationTokenRepository interface {
	Create(ctx context.Context, userID model.NanoID14, tokenHash string, expiresAt time.Time) error
	FindByUserID(ctx context.Context, userID model.NanoID14) (*domain.VerificationToken, error)
	DeleteByUserID(ctx context.Context, userID model.NanoID14) error
	DeleteExpired(ctx context.Context) error
}

type verificationTokenRepository struct {
	db *gorm.DB
}

func NewVerificationTokenRepository(db *gorm.DB) VerificationTokenRepository {
	return &verificationTokenRepository{db: db}
}

func (r *verificationTokenRepository) Create(ctx context.Context, userID model.NanoID14, tokenHash string, expiresAt time.Time) error {
	token := &domain.VerificationToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	// Set the current user ID in PostgreSQL session for RLS policies
	err := r.db.WithContext(ctx).Exec("SELECT set_config('app.current_user_id', ?, false)", string(userID)).Error
	if err != nil {
		return err
	}

	result := r.db.WithContext(ctx).Create(token)
	return result.Error
}

func (r *verificationTokenRepository) FindByUserID(ctx context.Context, userID model.NanoID14) (*domain.VerificationToken, error) {
	var token domain.VerificationToken

	// Set the current user ID in PostgreSQL session for RLS policies
	err := r.db.WithContext(ctx).Exec("SELECT set_config('app.current_user_id', ?, false)", string(userID)).Error
	if err != nil {
		return nil, err
	}

	result := r.db.WithContext(ctx).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		First(&token)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domain.ErrVerificationTokenNotFound
		}
		return nil, result.Error
	}

	return &token, nil
}

func (r *verificationTokenRepository) DeleteByUserID(ctx context.Context, userID model.NanoID14) error {
	// Set the current user ID in PostgreSQL session for RLS policies
	err := r.db.WithContext(ctx).Exec("SELECT set_config('app.current_user_id', ?, false)", string(userID)).Error
	if err != nil {
		return err
	}

	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&domain.VerificationToken{})

	return result.Error
}

func (r *verificationTokenRepository) DeleteExpired(ctx context.Context) error {
	result := r.db.WithContext(ctx).
		Where("expires_at <= ?", time.Now()).
		Delete(&domain.VerificationToken{})

	return result.Error
}
