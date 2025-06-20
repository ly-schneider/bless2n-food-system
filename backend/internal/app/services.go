package app

import (
	"backend/internal/service"

	"go.uber.org/fx"
)

func NewServices() fx.Option {
	return fx.Options(
		fx.Provide(service.NewAuthService),
	)
}
