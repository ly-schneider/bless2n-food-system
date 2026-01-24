package app

import (
	"context"

	"backend/internal/auth"
	"backend/internal/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// NewJWKSClient creates a new JWKS client for Neon Auth.
func NewJWKSClient(cfg config.Config, logger *zap.Logger) (*auth.JWKSClient, error) {
	if cfg.NeonAuth.URL == "" {
		logger.Warn("NEON_AUTH_URL not configured, auth will not work")
		return nil, nil
	}
	return auth.NewJWKSClient(cfg.NeonAuth.URL, logger), nil
}

// NewNeonAuthMiddleware creates a new Neon Auth middleware.
func NewNeonAuthMiddleware(cfg config.Config, jwksClient *auth.JWKSClient, logger *zap.Logger) *auth.NeonAuthMiddleware {
	if jwksClient == nil {
		logger.Warn("JWKS client not configured, auth middleware will reject all requests")
		return nil
	}
	return auth.NewNeonAuthMiddleware(jwksClient, cfg.NeonAuth.URL, cfg.NeonAuth.Audience, logger)
}

// StartJWKSClient starts the JWKS client background refresh.
func StartJWKSClient(lc fx.Lifecycle, jwksClient *auth.JWKSClient, logger *zap.Logger) {
	if jwksClient == nil {
		return
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting JWKS client")
			return jwksClient.Start(ctx)
		},
		OnStop: func(_ context.Context) error {
			logger.Info("stopping JWKS client")
			jwksClient.Stop()
			return nil
		},
	})
}
