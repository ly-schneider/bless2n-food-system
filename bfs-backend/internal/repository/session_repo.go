package repository

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/session"
	"backend/internal/generated/ent/user"
)

type SessionWithUser struct {
	UserID    string
	Role      user.Role
	Email     string
	Name      string
	UpdatedAt time.Time // needed for sliding refresh check
}

type SessionRepository interface {
	GetByToken(ctx context.Context, token string) (*SessionWithUser, error)
	RefreshSession(ctx context.Context, token string, expiresIn time.Duration) error
	CreateSession(ctx context.Context, userID string, expiresIn time.Duration) (string, error)
}

type sessionRepo struct {
	client *ent.Client
}

func NewSessionRepository(client *ent.Client) SessionRepository {
	return &sessionRepo{client: client}
}

func (r *sessionRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *sessionRepo) GetByToken(ctx context.Context, token string) (*SessionWithUser, error) {
	e, err := r.ec(ctx).Session.Query().
		Where(session.TokenEQ(token), session.ExpiresAtGT(time.Now().UTC())).
		WithUser().
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}

	u := e.Edges.User
	if u == nil {
		return nil, ErrNotFound
	}

	var email, name string
	if u.Email != nil {
		email = *u.Email
	}
	if u.Name != nil {
		name = *u.Name
	}

	return &SessionWithUser{
		UserID:    e.UserID,
		Role:      u.Role,
		Email:     email,
		Name:      name,
		UpdatedAt: e.UpdatedAt,
	}, nil
}

func (r *sessionRepo) RefreshSession(ctx context.Context, token string, expiresIn time.Duration) error {
	now := time.Now().UTC()
	n, err := r.ec(ctx).Session.Update().
		Where(session.TokenEQ(token)).
		SetUpdatedAt(now).
		SetExpiresAt(now.Add(expiresIn)).
		Save(ctx)
	if err != nil {
		return translateError(err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sessionRepo) CreateSession(ctx context.Context, userID string, expiresIn time.Duration) (string, error) {
	// Generate random token: 32 bytes, base64url no padding
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	// Generate random session ID: "sess_" + 24 random bytes base64url
	idBytes := make([]byte, 24)
	if _, err := rand.Read(idBytes); err != nil {
		return "", fmt.Errorf("generate session id: %w", err)
	}
	sessionID := "sess_" + base64.RawURLEncoding.EncodeToString(idBytes)

	now := time.Now().UTC()
	_, err := r.ec(ctx).Session.Create().
		SetID(sessionID).
		SetToken(token).
		SetUserID(userID).
		SetExpiresAt(now.Add(expiresIn)).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Save(ctx)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	return token, nil
}
