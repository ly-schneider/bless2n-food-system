package service

import (
	"context"

	"backend/internal/domain"
	"backend/internal/logger"
	"backend/internal/repository"

	"github.com/google/uuid"
)

type RefreshTokenService interface {
	GetByTokenHash(ctx context.Context, tokenHash string) (domain.RefreshToken, error)
	Create(ctx context.Context, rt *domain.RefreshToken) error
	Revoke(ctx context.Context, id uuid.UUID) error
}

type refreshTokenService struct {
	repo repository.RefreshTokenRepository
}

func NewRefreshTokenService(r repository.RefreshTokenRepository) RefreshTokenService {
	return &refreshTokenService{repo: r}
}
func (s *refreshTokenService) GetByTokenHash(ctx context.Context, tokenHash string) (domain.RefreshToken, error) {
	logger.L.Infow("Getting refresh token by hashed token", "hashed_token", tokenHash)
	rt, err := s.repo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		logger.L.Errorw("Failed to get refresh token by hashed token", "hashed_token", tokenHash, "error", err)
		return rt, err
	}

	if rt.ID == uuid.Nil {
		logger.L.Infow("No refresh token found with hashed token", "hashed_token", tokenHash)
	} else {
		logger.L.Infow("Successfully retrieved refresh token", "id", rt.ID, "hashed_token", tokenHash)
	}
	return rt, nil
}

func (s *refreshTokenService) Create(ctx context.Context, rt *domain.RefreshToken) error {
	logger.L.Infow("Creating refresh token", "user_id", rt.UserID, "hashed_token", rt.TokenHash)

	if err := s.repo.Create(ctx, rt); err != nil {
		logger.L.Errorw("Failed to create refresh token", "user_id", rt.UserID, "error", err)
		return err
	}

	logger.L.Infow("Refresh token created successfully", "id", rt.ID, "user_id", rt.UserID, "hashed_token", rt.TokenHash)
	return nil
}

func (s *refreshTokenService) Revoke(ctx context.Context, id uuid.UUID) error {
	logger.L.Infow("Revoking refresh token", "id", id)

	if err := s.repo.Revoke(ctx, id); err != nil {
		logger.L.Errorw("Failed to revoke refresh token", "id", id, "error", err)
		return err
	}

	logger.L.Infow("Refresh token revoked successfully", "id", id)
	return nil
}
