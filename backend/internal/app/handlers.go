package app

import (
	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/middleware"

	"go.uber.org/fx"
)

func NewHandlers() fx.Option {
	return fx.Options(
		fx.Provide(
			handler.NewAuthHandler,
			handler.NewAdminHandler,
			handler.NewUserHandler,
			handler.NewStationHandler,
			middleware.NewJWTMiddleware,
			NewSecurityMiddleware,
		),
	)
}

func NewSecurityMiddleware(cfg config.Config) *middleware.SecurityMiddleware {
	return middleware.NewSecurityMiddleware(middleware.SecurityConfig{
		EnableHSTS:     cfg.Security.EnableHSTS,
		EnableCSP:      cfg.Security.EnableCSP,
		TrustedOrigins: cfg.Security.TrustedOrigins,
		AppEnv:         cfg.App.AppEnv,
	})
}
