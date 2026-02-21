package app

import (
	"backend/internal/api"
	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/middleware"
	"backend/internal/service"

	"go.uber.org/fx"
)

func NewHandlers() fx.Option {
	return fx.Options(
		fx.Provide(
			NewSecurityMiddleware,
			NewSystemDisableMiddleware,
			auth.NewDeviceAuthMiddleware,
			api.NewHandlers,
		),
	)
}

func NewSystemDisableMiddleware(settings service.SettingsService, sessionMw *auth.SessionMiddleware) *middleware.SystemDisableMiddleware {
	return middleware.NewSystemDisableMiddleware(settings, sessionMw)
}

func NewSecurityMiddleware(cfg config.Config) *middleware.SecurityMiddleware {
	return middleware.NewSecurityMiddleware(middleware.SecurityConfig{
		EnableHSTS:     cfg.Security.EnableHSTS,
		EnableCSP:      cfg.Security.EnableCSP,
		TrustedOrigins: cfg.Security.TrustedOrigins,
		AppEnv:         cfg.App.AppEnv,
	})
}
