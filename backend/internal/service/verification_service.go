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
)

type VerificationService interface {
	SendVerificationCode(ctx context.Context, userID model.NanoID14, email, name string) error
	VerifyCode(ctx context.Context, userID model.NanoID14, code string) error
}

type verificationService struct {
	verificationTokenRepo repository.VerificationTokenRepository
	emailService          EmailService
}

func NewVerificationService(
	verificationTokenRepo repository.VerificationTokenRepository,
	emailService EmailService,
) VerificationService {
	return &verificationService{
		verificationTokenRepo: verificationTokenRepo,
		emailService:          emailService,
	}
}

func (s *verificationService) SendVerificationCode(ctx context.Context, userID model.NanoID14, email, name string) error {
	// Delete any existing verification tokens for this user
	if err := s.verificationTokenRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete existing verification tokens: %w", err)
	}

	// Generate 6-digit code
	code := s.generateSixDigitCode()

	// Hash the code before storing
	codeHashString, err := utils.HashOTP(code)
	if err != nil {
		return fmt.Errorf("failed to hash verification code: %w", err)
	}
	codeHash := []byte(codeHashString)

	// Set expiration to 15 minutes from now
	expiresAt := time.Now().Add(15 * time.Minute)

	// Store the hashed code
	if err := s.verificationTokenRepo.Create(ctx, userID, codeHash, expiresAt); err != nil {
		return fmt.Errorf("failed to create verification token: %w", err)
	}

	// Send email with the plain code
	if err := s.emailService.SendVerificationEmail(ctx, email, name, code); err != nil {
		// If email sending fails, clean up the token
		_ = s.verificationTokenRepo.DeleteByUserID(ctx, userID)
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

func (s *verificationService) VerifyCode(ctx context.Context, userID model.NanoID14, code string) error {
	// Find the verification token
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

	// Verify the provided code against the stored hash
	if !utils.VerifyOTP(code, string(token.TokenHash)) {
		return fmt.Errorf("invalid verification code")
	}

	// Delete the token after successful verification
	if err := s.verificationTokenRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete verification token: %w", err)
	}

	return nil
}

func (s *verificationService) generateSixDigitCode() string {
	// Generate a random 6-digit code
	code := rand.Intn(900000) + 100000 // Ensures 6 digits (100000-999999)
	return fmt.Sprintf("%06d", code)
}

