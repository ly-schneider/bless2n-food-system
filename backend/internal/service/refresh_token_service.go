package service

import (
	"context"
	"errors"

	"backend/internal/domain"
	"backend/internal/logger"
	"backend/internal/repository"
)

type RefreshTokenService interface {
	GetByTokenHash(ctx context.Context, tokenHash string) (domain.RefreshToken, error)
	Create(ctx context.Context, rt *domain.RefreshToken) error
	Revoke(ctx context.Context, id string) error
}

type refreshTokenService struct {
	repo repository.RefreshTokenRepository
}

func NewRefreshTokenService(r repository.RefreshTokenRepository) RefreshTokenService {
	return &refreshTokenService{repo: r}
}

func (s *refreshTokenService) GetByTokenHash(ctx context.Context, tokenHash string) (domain.RefreshToken, error) {
	logger.L.Infow("Getting refresh token by hashed token", "hashed_token", tokenHash)

	if tokenHash == "" {
		err := errors.New("token hash cannot be empty")
		logger.L.Error(err.Error())
		return domain.RefreshToken{}, err
	}

	rt, err := s.repo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		logger.L.Errorw("Failed to get refresh token by hashed token", "hashed_token", tokenHash, "error", err)
		return rt, err
	}

	if rt.ID == "" {
		logger.L.Infow("No refresh token found with hashed token", "hashed_token", tokenHash)
	} else {
		logger.L.Infow("Successfully retrieved refresh token", "id", rt.ID, "hashed_token", tokenHash)
	}
	return rt, nil
}

func (s *refreshTokenService) Create(ctx context.Context, rt *domain.RefreshToken) error {
	logger.L.Infow("Creating refresh token", "user_id", rt.UserID, "hashed_token", rt.TokenHash)

	if rt.UserID == "" {
		err := errors.New("user ID cannot be empty")
		logger.L.Error(err.Error())
		return err
	}

	if rt.TokenHash == "" {
		err := errors.New("token hash cannot be empty")
		logger.L.Error(err.Error())
		return err
	}

	if err := s.repo.Create(ctx, rt); err != nil {
		logger.L.Errorw("Failed to create refresh token", "user_id", rt.UserID, "error", err)
		return err
	}

	logger.L.Infow("Refresh token created successfully", "id", rt.ID, "user_id", rt.UserID, "hashed_token", rt.TokenHash)
	return nil
}

func (s *refreshTokenService) Revoke(ctx context.Context, id string) error {
	logger.L.Infow("Revoking refresh token", "id", id)

	if id == "" {
		err := errors.New("refresh token ID cannot be empty")
		logger.L.Error(err.Error())
		return err
	}

	if err := s.repo.Revoke(ctx, id); err != nil {
		logger.L.Errorw("Failed to revoke refresh token", "id", id, "error", err)
		return err
	}

	logger.L.Infow("Refresh token revoked successfully", "id", id)
	return nil
}
