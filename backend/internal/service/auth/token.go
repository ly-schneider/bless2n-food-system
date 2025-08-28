package auth

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"backend/internal/domain"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/utils"
)

const RefreshTokenDuration = 7 * 24 * time.Hour // 7 days

type TokenService interface {
	GenerateTokenPair(ctx context.Context, user *domain.User, clientID string) (*service.TokenPairResponse, error)
	RefreshTokenPair(ctx context.Context, refreshToken, clientID string) (*service.TokenPairResponse, error)
	RevokeTokenFamily(ctx context.Context, refreshToken, reason string) error
}

type tokenService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtService       service.JWTService
}

func NewTokenService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtService service.JWTService,
) TokenService {
	return &tokenService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
	}
}

func (s *tokenService) GenerateTokenPair(ctx context.Context, user *domain.User, clientID string) (*service.TokenPairResponse, error) {
	// Revoke all existing refresh tokens for this user and client
	err := s.refreshTokenRepo.RevokeByClientID(ctx, user.ID, clientID, "new_login")
	if err != nil {
		zap.L().Error("failed to revoke existing refresh tokens", zap.Error(err))
		return nil, fmt.Errorf("failed to revoke existing refresh tokens: %w", err)
	}

	// Generate token pair
	tokenPair, err := s.jwtService.GenerateTokenPair(user, clientID)
	if err != nil {
		zap.L().Error("failed to generate token pair", zap.Error(err))
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Generate family ID for refresh token rotation
	familyID, err := utils.GenerateFamilyID()
	if err != nil {
		zap.L().Error("failed to generate family ID", zap.Error(err))
		return nil, fmt.Errorf("failed to generate family ID: %w", err)
	}

	// Store refresh token
	refreshToken := &domain.RefreshToken{
		UserID:    user.ID,
		ClientID:  clientID,
		FamilyID:  familyID,
		ExpiresAt: time.Now().Add(RefreshTokenDuration),
	}

	if err := s.refreshTokenRepo.CreateWithPlainToken(ctx, refreshToken, tokenPair.RefreshToken); err != nil {
		zap.L().Error("failed to store refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	zap.L().Info("token pair generated successfully",
		zap.String("user_id", user.ID.Hex()),
		zap.String("client_id", clientID))

	return tokenPair, nil
}

func (s *tokenService) RefreshTokenPair(ctx context.Context, refreshToken, clientID string) (*service.TokenPairResponse, error) {
	// Find the refresh token
	storedToken, err := s.refreshTokenRepo.GetValidTokenForUser(ctx, refreshToken)
	if err != nil {
		if err == domain.ErrRefreshTokenNotFound {
			zap.L().Warn("refresh token not found or invalid")
			return nil, fmt.Errorf("invalid refresh token")
		}
		zap.L().Error("failed to get refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}

	// Check client ID match
	if storedToken.ClientID != clientID {
		zap.L().Warn("client ID mismatch for refresh token",
			zap.String("stored_client_id", storedToken.ClientID),
			zap.String("request_client_id", clientID))
		return nil, fmt.Errorf("invalid client")
	}

	return s.rotateRefreshToken(ctx, storedToken, clientID)
}

func (s *tokenService) rotateRefreshToken(ctx context.Context, storedToken *domain.RefreshToken, clientID string) (*service.TokenPairResponse, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, storedToken.UserID)
	if err != nil {
		zap.L().Error("failed to get user", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.IsDisabled {
		return nil, fmt.Errorf("account is disabled")
	}

	// Generate new token pair
	tokenPair, err := s.jwtService.GenerateTokenPair(user, clientID)
	if err != nil {
		zap.L().Error("failed to generate token pair", zap.Error(err))
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Revoke the old refresh token (refresh token rotation)
	if err := s.refreshTokenRepo.RevokeByID(ctx, storedToken.ID, "token_rotation"); err != nil {
		zap.L().Error("failed to revoke old refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to rotate token: %w", err)
	}

	// Create new refresh token with same family ID
	newRefreshToken := &domain.RefreshToken{
		UserID:    storedToken.UserID,
		ClientID:  clientID,
		FamilyID:  storedToken.FamilyID, // Keep same family ID for rotation
		ExpiresAt: time.Now().Add(RefreshTokenDuration),
	}

	if err := s.refreshTokenRepo.CreateWithPlainToken(ctx, newRefreshToken, tokenPair.RefreshToken); err != nil {
		zap.L().Error("failed to store new refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Update last used timestamp of the stored token for tracking
	if err := s.refreshTokenRepo.UpdateLastUsed(ctx, storedToken.ID); err != nil {
		zap.L().Warn("failed to update last used timestamp", zap.Error(err))
	}

	zap.L().Info("refresh token rotated successfully",
		zap.String("user_id", storedToken.UserID.Hex()),
		zap.String("client_id", clientID))

	return &service.TokenPairResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

func (s *tokenService) RevokeTokenFamily(ctx context.Context, refreshToken, reason string) error {
	// Find the refresh token
	storedToken, err := s.refreshTokenRepo.GetValidTokenForUser(ctx, refreshToken)
	if err != nil {
		if err == domain.ErrRefreshTokenNotFound {
			// Token not found, but that's okay for logout
			return nil
		}
		zap.L().Error("failed to get refresh token", zap.Error(err))
		return fmt.Errorf("failed to validate refresh token: %w", err)
	}

	// Revoke all tokens in the same family
	if err := s.refreshTokenRepo.RevokeByFamilyID(ctx, storedToken.FamilyID, reason); err != nil {
		zap.L().Error("failed to revoke token family", zap.Error(err))
		return fmt.Errorf("failed to revoke tokens: %w", err)
	}

	zap.L().Info("token family revoked successfully",
		zap.String("user_id", storedToken.UserID.Hex()),
		zap.String("family_id", storedToken.FamilyID),
		zap.String("reason", reason))

	return nil
}
