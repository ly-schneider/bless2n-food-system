package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend/internal/config"
	"backend/internal/generated/ent"
	"backend/internal/generated/ent/admininvite"
	"backend/internal/generated/ent/user"
	"backend/internal/repository"

	"github.com/google/uuid"
)

const (
	defaultInviteTTL = 7 * 24 * time.Hour  // 7 days
	maxInviteTTL     = 30 * 24 * time.Hour // 30 days
)

var (
	ErrInviteNotFound    = errors.New("invite not found")
	ErrInviteExpired     = errors.New("invite has expired")
	ErrInviteNotPending  = errors.New("invite is not pending")
	ErrInviteAlreadyUsed = errors.New("invite has already been used")
	ErrInvalidToken      = errors.New("invalid token")
	ErrFirstNameRequired = errors.New("first name is required")
)

type AdminInviteService interface {
	// Admin operations (requires auth)
	List(ctx context.Context, status *string, email *string) ([]*ent.AdminInvite, int64, error)
	GetByID(ctx context.Context, id string) (*ent.AdminInvite, error)
	Create(ctx context.Context, inviterID, email string, expiresInSec *int) (*ent.AdminInvite, error)
	Delete(ctx context.Context, id string) error
	Revoke(ctx context.Context, id string) error
	Resend(ctx context.Context, id string) error

	// Public operations (no auth)
	Verify(ctx context.Context, token string) (*ent.AdminInvite, error)
	Accept(ctx context.Context, token, firstName, lastName string) error
}

type adminInviteService struct {
	cfg        config.Config
	inviteRepo repository.AdminInviteRepository
	userRepo   repository.UserRepository
	emailSvc   EmailService
}

func NewAdminInviteService(
	cfg config.Config,
	inviteRepo repository.AdminInviteRepository,
	userRepo repository.UserRepository,
	emailSvc EmailService,
) AdminInviteService {
	return &adminInviteService{
		cfg:        cfg,
		inviteRepo: inviteRepo,
		userRepo:   userRepo,
		emailSvc:   emailSvc,
	}
}

func (s *adminInviteService) GetByID(ctx context.Context, id string) (*ent.AdminInvite, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, ErrInviteNotFound
	}

	invite, err := s.inviteRepo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInviteNotFound
		}
		return nil, err
	}

	return invite, nil
}

func (s *adminInviteService) List(ctx context.Context, status *string, email *string) ([]*ent.AdminInvite, int64, error) {
	var inviteStatus *admininvite.Status
	if status != nil && *status != "" {
		st := admininvite.Status(*status)
		inviteStatus = &st
	}

	invites, total, err := s.inviteRepo.List(ctx, inviteStatus, email)
	if err != nil {
		return nil, 0, err
	}

	return invites, total, nil
}

func (s *adminInviteService) Create(ctx context.Context, inviterID, email string, expiresInSec *int) (*ent.AdminInvite, error) {
	// Normalize email
	email = strings.ToLower(strings.TrimSpace(email))

	// Calculate expiry
	ttl := defaultInviteTTL
	if expiresInSec != nil && *expiresInSec > 0 {
		ttl = time.Duration(*expiresInSec) * time.Second
		if ttl > maxInviteTTL {
			ttl = maxInviteTTL
		}
	}
	expiresAt := time.Now().Add(ttl)

	// Generate token
	token, tokenHash, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	invite, err := s.inviteRepo.Create(ctx, inviterID, email, tokenHash, admininvite.StatusPending, expiresAt)
	if err != nil {
		return nil, err
	}

	// Build invite URL and send email
	inviteURL := s.buildInviteURL(token)
	if err := s.emailSvc.SendInviteEmail(ctx, email, inviteURL, expiresAt); err != nil {
		// Log but don't fail - invite was created
		// The caller can resend if needed
	}

	return invite, nil
}

func (s *adminInviteService) Delete(ctx context.Context, id string) error {
	uuid, err := parseUUID(id)
	if err != nil {
		return ErrInviteNotFound
	}

	if err := s.inviteRepo.Delete(ctx, uuid); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInviteNotFound
		}
		return err
	}
	return nil
}

func (s *adminInviteService) Revoke(ctx context.Context, id string) error {
	uuid, err := parseUUID(id)
	if err != nil {
		return ErrInviteNotFound
	}

	invite, err := s.inviteRepo.GetByID(ctx, uuid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInviteNotFound
		}
		return err
	}

	if invite.Status != admininvite.StatusPending {
		return ErrInviteNotPending
	}

	return s.inviteRepo.UpdateStatus(ctx, uuid, admininvite.StatusRevoked)
}

func (s *adminInviteService) Resend(ctx context.Context, id string) error {
	uuid, err := parseUUID(id)
	if err != nil {
		return ErrInviteNotFound
	}

	invite, err := s.inviteRepo.GetByID(ctx, uuid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInviteNotFound
		}
		return err
	}

	if invite.Status != admininvite.StatusPending {
		return ErrInviteNotPending
	}

	// Generate new token and extend expiry
	token, tokenHash, err := generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	expiresAt := time.Now().Add(defaultInviteTTL)

	if err := s.inviteRepo.UpdateTokenAndExpiry(ctx, uuid, tokenHash, expiresAt); err != nil {
		return err
	}

	// Send email
	inviteURL := s.buildInviteURL(token)
	return s.emailSvc.SendInviteEmail(ctx, invite.InviteeEmail, inviteURL, expiresAt)
}

func (s *adminInviteService) Verify(ctx context.Context, token string) (*ent.AdminInvite, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}

	tokenHash := hashToken(token)

	invite, err := s.inviteRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	return invite, nil
}

func (s *adminInviteService) Accept(ctx context.Context, token, firstName, lastName string) error {
	if token == "" {
		return ErrInvalidToken
	}

	firstName = strings.TrimSpace(firstName)
	if firstName == "" {
		return ErrFirstNameRequired
	}

	tokenHash := hashToken(token)

	invite, err := s.inviteRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInvalidToken
		}
		return err
	}

	// Check if can be accepted
	if invite.Status != admininvite.StatusPending {
		return ErrInviteAlreadyUsed
	}
	if time.Now().After(invite.ExpiresAt) {
		return ErrInviteExpired
	}

	// Build full name
	name := firstName
	if lastName := strings.TrimSpace(lastName); lastName != "" {
		name = firstName + " " + lastName
	}

	email := strings.ToLower(strings.TrimSpace(invite.InviteeEmail))

	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return err
	}

	if existingUser != nil {
		// Update existing user to admin role
		if err := s.userRepo.UpdateRoleAndName(ctx, existingUser.ID, user.RoleAdmin, name); err != nil {
			return err
		}
	} else {
		// Create new user with admin role
		if _, err := s.userRepo.CreateAdminUser(ctx, email, name); err != nil {
			return err
		}
	}

	// Mark invite as accepted
	return s.inviteRepo.MarkAccepted(ctx, invite.ID)
}

func (s *adminInviteService) buildInviteURL(token string) string {
	baseURL := s.cfg.App.PublicBaseURL
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	return fmt.Sprintf("%s/invite/accept?token=%s", baseURL, token)
}

// generateToken creates a new secure random token and its SHA-256 hash.
func generateToken() (token string, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	hash = hashToken(token)
	return token, hash, nil
}

// hashToken computes SHA-256 hash of a token.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// parseUUID parses a string UUID, returning an error if invalid.
func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
