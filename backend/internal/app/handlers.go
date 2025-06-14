package app

import (
	"backend/internal/handlers"

	"go.uber.org/fx"
)

func NewHandlers() fx.Option {
	return fx.Options(
		fx.Provide(handlers.NewAuthHandler),
	)
}
