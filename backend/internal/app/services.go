package app

import (
	"backend/internal/service"

	"go.uber.org/fx"
)

// NewServices provides all application services with their dependencies
func NewServices() fx.Option {
	return fx.Options(
		fx.Provide(service.NewAuthService),
		fx.Provide(service.NewUserService),
	)
}
