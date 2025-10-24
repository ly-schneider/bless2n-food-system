package service

import (
	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/repository"
	"backend/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type FederatedAuthService interface {
	// Google code exchange (Authorization Code + PKCE)
	SignInWithGoogleCode(ctx context.Context, code string, codeVerifier string, redirectURI string, nonce string, deviceLabel string) (*TokenPairResponse, *domain.User, error)
}

type federatedAuthService struct {
	cfg           config.Config
	jwtService    JWTService
	users         repository.UserRepository
	ids           repository.IdentityLinkRepository
	refreshTokens repository.RefreshTokenRepository
	jwks          *jwksCache
}

func NewFederatedAuthService(cfg config.Config, jwt JWTService, users repository.UserRepository, ids repository.IdentityLinkRepository, rts repository.RefreshTokenRepository) FederatedAuthService {
	return &federatedAuthService{cfg: cfg, jwtService: jwt, users: users, ids: ids, refreshTokens: rts, jwks: newJWKSCache()}
}

const (
	federatedRefreshTTL = 7 * 24 * time.Hour
)

// ---------- Google ----------

type googleTokenResp struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	IdToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

func (s *federatedAuthService) SignInWithGoogleCode(ctx context.Context, code string, codeVerifier string, redirectURI string, nonce string, deviceLabel string) (*TokenPairResponse, *domain.User, error) {
	// Exchange code for tokens
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("client_id", s.cfg.OAuth.Google.ClientID)
	form.Set("code_verifier", codeVerifier)
	form.Set("redirect_uri", redirectURI)
	if cs := strings.TrimSpace(s.cfg.OAuth.Google.ClientSecret); cs != "" {
		form.Set("client_secret", cs)
	}
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// best-effort read small error body for logs
		var errObj map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errObj)
		// Log without including secrets
		zap.L().Warn("google token exchange failed", zap.Int("status", resp.StatusCode), zap.Any("error", errObj))
		return nil, nil, errors.New("google_exchange_failed")
	}
	var tr googleTokenResp
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, nil, err
	}
	if tr.IdToken == "" {
		return nil, nil, errors.New("missing_id_token")
	}
	// Verify id_token (iss/aud/exp/signature)
	claims, err := VerifyIDToken(ctx, tr.IdToken, "https://www.googleapis.com/oauth2/v3/certs", []string{"https://accounts.google.com", "accounts.google.com"}, s.cfg.OAuth.Google.ClientID, nonce, s.jwks)
	if err != nil {
		return nil, nil, fmt.Errorf("google_id_token_invalid: %w", err)
	}
	// Email verified?
	emailVerified := false
	switch v := claims.EmailVerified.(type) {
	case bool:
		emailVerified = v
	case string:
		emailVerified = strings.ToLower(v) == "true"
	}
	return s.linkAndIssue(ctx, domain.ProviderGoogle, claims.Sub, claims.Email, emailVerified, claims.Name, claims.Picture, deviceLabel)
}

// linkAndIssue finds/creates the user, ensures identity link, and issues our JWTs + rotated refresh
func (s *federatedAuthService) linkAndIssue(ctx context.Context, provider domain.IdentityProvider, sub string, email string, emailVerified bool, name string, picture string, deviceLabel string) (*TokenPairResponse, *domain.User, error) {
	// 1) Find identity link first
	if link, err := s.ids.FindByProviderAndSub(ctx, provider, sub); err == nil {
		user, err := s.users.FindByID(ctx, link.UserID)
		if err != nil {
			return nil, nil, err
		}
		return s.issueSession(ctx, user, deviceLabel)
	}
	// 2) No link; attempt match by verified email
	var user *domain.User
	var err error
	if email != "" && emailVerified {
		if u, e := s.users.FindByEmail(ctx, strings.ToLower(email)); e == nil && u != nil {
			user = u
		}
	}
	// 3) Create new user if needed
	if user == nil {
		// Use email if present; otherwise synthesize a placeholder tied to subject
		em := email
		if em == "" {
			em = fmt.Sprintf("user-%s@invalid.local", sub)
		}
		user, err = s.users.UpsertCustomerByEmail(ctx, em)
		if err != nil {
			return nil, nil, err
		}
	}
	// 4) Create identity link
	var snapEmail *string
	if email != "" {
		snapEmail = &email
	}
	var snapName *string
	if name != "" {
		snapName = &name
	}
	var snapAvatar *string
	if picture != "" {
		snapAvatar = &picture
	}
	if _, err := s.ids.UpsertLink(ctx, provider, sub, user.ID, snapEmail, snapName, snapAvatar); err != nil {
		return nil, nil, err
	}
	// 5) Issue tokens
	return s.issueSession(ctx, user, deviceLabel)
}

func (s *federatedAuthService) issueSession(ctx context.Context, user *domain.User, clientID string) (*TokenPairResponse, *domain.User, error) {
	access, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, nil, err
	}
	rt, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, nil, err
	}
	now := time.Now().UTC()
	family, err := utils.GenerateRandomURLSafe(24)
	if err != nil {
		return nil, nil, err
	}
	if _, err := s.refreshTokens.Create(ctx, &domain.RefreshToken{
		UserID:     user.ID,
		ClientID:   clientID,
		TokenHash:  utils.HashTokenSHA256(rt),
		IssuedAt:   now,
		LastUsedAt: time.Time{},
		ExpiresAt:  now.Add(federatedRefreshTTL),
		IsRevoked:  false,
		FamilyID:   family,
	}); err != nil {
		return nil, nil, err
	}
	return &TokenPairResponse{AccessToken: access, RefreshToken: rt, TokenType: "Bearer", ExpiresIn: int64(AccessTokenDuration.Seconds())}, user, nil
}
