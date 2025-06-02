package services

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type AuthConfig struct {
	ClientID     string
	ClientSecret string
	Domain       string // e.g. "your-tenant.eu.auth0.com"
	BaseURL      string // e.g. "http://localhost:3000"
}

type AuthService struct {
	cfg        AuthConfig
	oidcProv   *oidc.Provider
	oauth2Conf *oauth2.Config
	verifier   *oidc.IDTokenVerifier
}

func NewAuthService(cfg AuthConfig) (*AuthService, error) {
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, "https://"+cfg.Domain+"/")
	if err != nil {
		return nil, fmt.Errorf("oidc provider: %w", err)
	}

	scopes := []string{oidc.ScopeOpenID, "profile", "email"}
	oauthConf := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.BaseURL + "/auth/callback",
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}

	return &AuthService{
		cfg:        cfg,
		oidcProv:   provider,
		oauth2Conf: oauthConf,
		verifier:   provider.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
	}, nil
}

func (s *AuthService) LoginURL(state string) string {
	return s.oauth2Conf.AuthCodeURL(state)
}

func (s *AuthService) HandleCallback(ctx context.Context, code string) (map[string]any, error) {
	token, err := s.oauth2Conf.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("missing id_token")
	}

	idToken, err := s.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("verify id_token: %w", err)
	}

	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("parse claims: %w", err)
	}

	return claims, nil
}

func (s *AuthService) LogoutURL(returnTo string) string {
	return fmt.Sprintf(
		"https://%s/v2/logout?client_id=%s&returnTo=%s",
		s.cfg.Domain, s.cfg.ClientID, returnTo,
	)
}
