package service

import (
	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/repository"
	"backend/internal/utils"
	"context"
	"errors"
	"strings"
	"time"

	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/customer"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type UserService interface {
	GetByID(ctx context.Context, userID bson.ObjectID) (*domain.User, error)
	RequestEmailChange(ctx context.Context, userID bson.ObjectID, newEmail, ip, ua string) error
	ConfirmEmailChange(ctx context.Context, userID bson.ObjectID, code string) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID bson.ObjectID, firstName, lastName *string, newEmail *string, role domain.UserRole, ip, ua string) (*domain.User, bool, error)
	DeleteAccount(ctx context.Context, userID bson.ObjectID) error
}

type userService struct {
	cfg           config.Config
	users         repository.UserRepository
	email         EmailService
	emailTokens   repository.EmailChangeTokenRepository
	otps          repository.OTPTokenRepository
	refreshTokens repository.RefreshTokenRepository
}

func NewUserService(cfg config.Config, users repository.UserRepository, email EmailService, emailTokens repository.EmailChangeTokenRepository, otps repository.OTPTokenRepository, rts repository.RefreshTokenRepository) UserService {
	// Ensure Stripe key is set for any Stripe calls from this service
	stripe.Key = cfg.Stripe.SecretKey
	return &userService{cfg: cfg, users: users, email: email, emailTokens: emailTokens, otps: otps, refreshTokens: rts}
}

const (
	emailChangeTTL       = 15 * time.Minute
	maxEmailCodeAttempts = 5
)

func (s *userService) RequestEmailChange(ctx context.Context, userID bson.ObjectID, newEmail, ip, ua string) error {
	newEmail = strings.ToLower(strings.TrimSpace(newEmail))
	if newEmail == "" {
		return errors.New("invalid_email")
	}
	// Load current user
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u.Email == newEmail {
		return errors.New("same_email")
	}
	// Ensure not taken by another user
	if other, err := s.users.FindByEmail(ctx, newEmail); err == nil && other != nil && other.ID != userID {
		return errors.New("email_taken")
	}
	// Generate code and store token
	code, err := utils.GenerateOTP()
	if err != nil {
		return err
	}
	codeHash, err := utils.HashOTPArgon2(code)
	if err != nil {
		return err
	}
	if _, err := s.emailTokens.CreateWithCode(ctx, userID, newEmail, codeHash, time.Now().UTC().Add(emailChangeTTL)); err != nil {
		return err
	}
	// Send to new email address
	return s.email.SendEmailChangeVerification(ctx, newEmail, code, ip, ua, emailChangeTTL)
}

func (s *userService) ConfirmEmailChange(ctx context.Context, userID bson.ObjectID, code string) (*domain.User, error) {
	// Load tokens
	tokens, err := s.emailTokens.FindActiveByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	var matched *domain.EmailChangeToken
	for i := range tokens {
		ok, _ := utils.VerifyOTPArgon2(code, tokens[i].TokenHash)
		if ok {
			matched = &tokens[i]
			break
		}
	}
	if matched == nil {
		if len(tokens) > 0 {
			attempts, _ := s.emailTokens.IncrementAttempts(ctx, tokens[0].ID)
			if attempts >= maxEmailCodeAttempts {
				_ = s.emailTokens.MarkUsed(ctx, tokens[0].ID, time.Now().UTC())
			}
		}
		return nil, errors.New("invalid_code")
	}
	if matched.Attempts >= maxEmailCodeAttempts {
		return nil, errors.New("too_many_attempts")
	}
	// Ensure email still unique (race avoidance)
	if other, err := s.users.FindByEmail(ctx, matched.NewEmail); err == nil && other != nil && other.ID != userID {
		return nil, errors.New("email_taken")
	}
	// Mark token used
	if err := s.emailTokens.MarkUsed(ctx, matched.ID, time.Now().UTC()); err != nil {
		return nil, err
	}
	// Update user email and mark verified
	if err := s.users.UpdateEmail(ctx, userID, matched.NewEmail, true); err != nil {
		return nil, err
	}
	// Update Stripe customer email if present
	u, err := s.users.FindByID(ctx, userID)
	if err == nil && u.StripeCustomerID != nil && *u.StripeCustomerID != "" {
		_, _ = customer.Update(*u.StripeCustomerID, &stripe.CustomerParams{Email: stripe.String(u.Email)})
	}
	return u, nil
}

func (s *userService) DeleteAccount(ctx context.Context, userID bson.ObjectID) error {
	// Load user for Stripe cleanup
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	// Revoke all refresh tokens
	_ = s.refreshTokens.RevokeAllByUser(ctx, userID, "user_deleted")
	// Delete OTP and email-change tokens
	_ = s.otps.DeleteByUser(ctx, userID)
	_ = s.emailTokens.DeleteByUser(ctx, userID)
	// Best-effort: delete Stripe customer if exists
	if u.StripeCustomerID != nil && *u.StripeCustomerID != "" {
		// ignore errors
		_, _ = customer.Del(*u.StripeCustomerID, nil)
	}
	// Finally delete user record
	return s.users.DeleteByID(ctx, userID)
}

// UpdateProfile updates allowed fields. If newEmail is provided and differs, starts verification flow.
// Returns the current user and whether an email change was initiated.
func (s *userService) UpdateProfile(ctx context.Context, userID bson.ObjectID, firstName, lastName *string, newEmail *string, role domain.UserRole, ip, ua string) (*domain.User, bool, error) {
	// Load current user
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, false, err
	}
	// Handle email change
	initiated := false
	if newEmail != nil {
		ne := strings.ToLower(strings.TrimSpace(*newEmail))
		if ne != "" && ne != u.Email {
			if err := s.RequestEmailChange(ctx, userID, ne, ip, ua); err != nil {
				return nil, false, err
			}
			initiated = true
		}
	}
	// Admin can update names
	if role == domain.UserRoleAdmin {
		if err := s.users.UpdateNames(ctx, userID, firstName, lastName); err != nil {
			return nil, initiated, err
		}
	}
	// Reload and return
	u2, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, initiated, err
	}
	return u2, initiated, nil
}
func (s *userService) GetByID(ctx context.Context, userID bson.ObjectID) (*domain.User, error) {
	return s.users.FindByID(ctx, userID)
}
