package app

import (
	"backend/internal/handler"

	"go.uber.org/fx"
)

func NewHandlers() fx.Option {
	return fx.Options(
		fx.Provide(handler.NewAuthHandler),
		fx.Provide(handler.NewUserHandler),
	)
}
