package auth

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"backend/internal/domain"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/utils"
)

type OTPService interface {
	GenerateAndSend(ctx context.Context, userID primitive.ObjectID, email string, tokenType domain.TokenType) error
	Verify(ctx context.Context, userID primitive.ObjectID, otp string, tokenType domain.TokenType) error
}

type otpService struct {
	otpRepo      repository.OTPTokenRepository
	emailService service.EmailService
	logger       *zap.Logger
}

func NewOTPService(
	otpRepo repository.OTPTokenRepository,
	emailService service.EmailService,
	logger *zap.Logger,
) OTPService {
	return &otpService{
		otpRepo:      otpRepo,
		emailService: emailService,
		logger:       logger,
	}
}

func (s *otpService) GenerateAndSend(ctx context.Context, userID primitive.ObjectID, email string, tokenType domain.TokenType) error {
	// Generate OTP
	otp, err := utils.GenerateOTP()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Hash OTP using Argon2id
	otpHash, err := utils.HashOTPArgon2(otp)
	if err != nil {
		return fmt.Errorf("failed to hash OTP: %w", err)
	}

	// Create OTP token
	otpToken := &domain.OTPToken{
		UserID:    userID,
		TokenHash: otpHash,
		Type:      tokenType,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		CreatedAt: time.Now(),
	}

	// Save OTP token
	if err := s.otpRepo.Create(ctx, otpToken); err != nil {
		return fmt.Errorf("failed to save OTP token: %w", err)
	}

	// Send OTP via email
	if err := s.emailService.SendOTP(ctx, email, otp); err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	return nil
}

func (s *otpService) Verify(ctx context.Context, userID primitive.ObjectID, otp string, tokenType domain.TokenType) error {
	// Get latest OTP token for the user
	otpToken, err := s.otpRepo.GetLatestByUserAndType(ctx, userID, tokenType)
	if err != nil {
		s.logger.Error("failed to get OTP token", zap.Error(err))
		return fmt.Errorf("failed to get OTP token: %w", err)
	}
	if otpToken == nil {
		return fmt.Errorf("no valid OTP found. Please request a new one")
	}

	// Check expiry, usage, and attempts
	if time.Now().After(otpToken.ExpiresAt) {
		return fmt.Errorf("OTP has expired. Please request a new one")
	}
	if otpToken.UsedAt != nil {
		return fmt.Errorf("OTP has already been used. Please request a new one")
	}
	if otpToken.Attempts >= 3 {
		return fmt.Errorf("too many attempts. Please request a new OTP")
	}

	// Verify OTP using Argon2id
	ok, verr := utils.VerifyOTPArgon2(otp, otpToken.TokenHash)
	if verr != nil || !ok {
		// Increment attempts on any verification failure
		if err := s.otpRepo.IncrementAttempts(ctx, otpToken.ID); err != nil {
			s.logger.Error("failed to increment OTP attempts", zap.Error(err))
		}
		return fmt.Errorf("invalid OTP code")
	}

	// Mark OTP as used
	if err := s.otpRepo.MarkAsUsed(ctx, otpToken.ID); err != nil {
		s.logger.Error("failed to mark OTP as used", zap.Error(err))
		return fmt.Errorf("failed to mark OTP as used: %w", err)
	}

	return nil
}