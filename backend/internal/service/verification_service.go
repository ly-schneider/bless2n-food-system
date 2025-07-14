package service

import (
	"backend/internal/domain"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/utils"
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

type VerificationService interface {
	SendVerificationCode(ctx context.Context, userID model.NanoID14) error
	VerifyCode(ctx context.Context, userID model.NanoID14, code string) error
}

type verificationService struct {
	verificationTokenRepo repository.VerificationTokenRepository
	userRepo              repository.UserRepository
	emailService          EmailService
	logger                *zap.Logger
}

func NewVerificationService(
	verificationTokenRepo repository.VerificationTokenRepository,
	userRepo repository.UserRepository,
	emailService EmailService,
	logger *zap.Logger,
) VerificationService {
	return &verificationService{
		verificationTokenRepo: verificationTokenRepo,
		userRepo:              userRepo,
		emailService:          emailService,
		logger:                logger,
	}
}

func (s *verificationService) SendVerificationCode(ctx context.Context, userID model.NanoID14) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user by ID: %w", err)
	}

	if err := s.verificationTokenRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete existing verification tokens: %w", err)
	}

	code := s.generateSixDigitCode()

	codeHash, err := utils.HashOTP(code)
	if err != nil {
		return fmt.Errorf("failed to hash verification code: %w", err)
	}

	expiresAt := time.Now().Add(15 * time.Minute)

	if err := s.verificationTokenRepo.Create(ctx, userID, codeHash, expiresAt); err != nil {
		return fmt.Errorf("failed to create verification token: %w", err)
	}

	// Send email with the plain code
	if err := s.emailService.SendVerificationEmail(ctx, user.Email, user.FirstName, code); err != nil {
		// If email sending fails, clean up the token
		_ = s.verificationTokenRepo.DeleteByUserID(ctx, userID)
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

func (s *verificationService) VerifyCode(ctx context.Context, userID model.NanoID14, code string) error {
	token, err := s.verificationTokenRepo.FindByUserID(ctx, userID)
	if err != nil {
		if err == domain.ErrVerificationTokenNotFound {
			return domain.ErrVerificationTokenNotFound
		}
		return fmt.Errorf("failed to find verification token: %w", err)
	}

	// Check if token has expired
	if time.Now().After(token.ExpiresAt) {
		// Clean up expired token
		_ = s.verificationTokenRepo.DeleteByUserID(ctx, userID)
		return domain.ErrVerificationTokenExpired
	}

	if !utils.VerifyOTP(code, string(token.TokenHash)) {
		return fmt.Errorf("invalid verification code")
	}

	// Get the user and update is_verified to true
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.IsVerified = true
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user verification status: %w", err)
	}

	if err := s.verificationTokenRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete verification token: %w", err)
	}

	return nil
}

func (s *verificationService) generateSixDigitCode() string {
	code := rand.Intn(900000) + 100000
	return fmt.Sprintf("%06d", code)
}
