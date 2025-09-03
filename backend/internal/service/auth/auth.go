package auth

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"backend/internal/domain"
	"backend/internal/repository"
	"backend/internal/service"
)

type authService struct {
	userRepo     repository.UserRepository
	otpService   OTPService
	tokenService TokenService
}

func NewService(
	userRepo repository.UserRepository,
	otpService OTPService,
	tokenService TokenService,
) service.AuthService {
	return &authService{
		userRepo:     userRepo,
		otpService:   otpService,
		tokenService: tokenService,
	}
}

func (s *authService) RegisterCustomer(ctx context.Context, req service.RegisterCustomerRequest) (*service.RegisterCustomerResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		zap.L().Error("failed to check existing user", zap.Error(err))
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Create new customer user
	var user = &domain.User{
		Email:      req.Email,
		Role:       domain.UserRoleCustomer,
		IsVerified: false,
		IsDisabled: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		zap.L().Error("failed to create user", zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	zap.L().Info("customer registered successfully",
		zap.String("user_id", user.ID.Hex()),
		zap.String("email", user.Email))

	return &service.RegisterCustomerResponse{
		Message: "Registration successful.",
		UserID:  user.ID.Hex(),
	}, nil
}

func (s *authService) RequestOTP(ctx context.Context, req service.RequestOTPRequest) (*service.RequestOTPResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		zap.L().Error("failed to get user by email", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.IsDisabled {
		return nil, fmt.Errorf("account is disabled: %s", *user.DisabledReason)
	}

	// Generate and send new login OTP (works for both verified and unverified users)
	if err := s.otpService.GenerateAndSend(ctx, user.ID, req.Email, domain.TokenTypeLogin); err != nil {
		zap.L().Error("failed to generate and send login OTP", zap.Error(err))
		return nil, fmt.Errorf("failed to send login code: %w", err)
	}

	zap.L().Info("login OTP sent successfully",
		zap.String("user_id", user.ID.Hex()),
		zap.String("email", user.Email),
		zap.Bool("user_verified", user.IsVerified))

	return &service.RequestOTPResponse{
		Message: "Login code sent to your email.",
	}, nil
}

func (s *authService) Login(ctx context.Context, req service.LoginRequest) (*service.LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		zap.L().Error("failed to get user by email", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if user.IsDisabled {
		return nil, fmt.Errorf("account is disabled: %s", *user.DisabledReason)
	}

	// Verify OTP
	if err := s.otpService.Verify(ctx, user.ID, req.OTP, domain.TokenTypeLogin); err != nil {
		return nil, err
	}

	// Mark user as verified if not already verified (first-time login verification)
	if !user.IsVerified {
		user.IsVerified = true
		user.UpdatedAt = time.Now()
		if err := s.userRepo.Update(ctx, user); err != nil {
			zap.L().Error("failed to mark user as verified during login", zap.Error(err))
			return nil, fmt.Errorf("failed to verify user: %w", err)
		}
	}

	// Generate token pair
	tokenPair, err := s.tokenService.GenerateTokenPair(ctx, user, req.ClientID)
	if err != nil {
		return nil, err
	}

	zap.L().Info("user logged in successfully",
		zap.String("user_id", user.ID.Hex()),
		zap.String("email", user.Email),
		zap.String("client_id", req.ClientID),
		zap.Bool("first_time_verification", !user.IsVerified))

	return &service.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         user,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, req service.RefreshTokenRequest) (*service.RefreshTokenResponse, error) {
	// Refresh token pair
	tokenPair, err := s.tokenService.RefreshTokenPair(ctx, req.RefreshToken, req.ClientID)
	if err != nil {
		return nil, err
	}

	return &service.RefreshTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

func (s *authService) Logout(ctx context.Context, req service.LogoutRequest) (*service.LogoutResponse, error) {
	// Revoke token family
	if err := s.tokenService.RevokeTokenFamily(ctx, req.RefreshToken, "user_logout"); err != nil {
		return nil, err
	}

	return &service.LogoutResponse{Message: "Logged out successfully"}, nil
}
