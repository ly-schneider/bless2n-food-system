package repository

import (
	"context"
	"strings"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/verification"
)

type VerificationRepository interface {
	// GetRecentOTP retrieves the most recent valid OTP for the given identifier.
	// Identifier format: "{type}-otp-{email}" (e.g., "sign-in-otp-user@example.com")
	// Returns the OTP string parsed from value (format: "{otp}:{attempts}"), or ErrNotFound if none exists.
	GetRecentOTP(ctx context.Context, identifier string) (string, error)
}

type verificationRepo struct {
	client *ent.Client
}

func NewVerificationRepository(client *ent.Client) VerificationRepository {
	return &verificationRepo{client: client}
}

func (r *verificationRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *verificationRepo) GetRecentOTP(ctx context.Context, identifier string) (string, error) {
	e, err := r.ec(ctx).Verification.Query().
		Where(
			verification.IdentifierEQ(identifier),
			verification.ExpiresAtGT(time.Now()),
		).
		Order(verification.ByCreatedAt(entDescOpt())).
		First(ctx)
	if err != nil {
		return "", translateError(err)
	}

	// Value format is "{otp}:{attempts}" - extract the OTP
	parts := strings.SplitN(e.Value, ":", 2)
	if len(parts) == 0 || parts[0] == "" {
		return "", ErrNotFound
	}

	return parts[0], nil
}
