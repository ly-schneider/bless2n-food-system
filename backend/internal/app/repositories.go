package app

import (
	"backend/internal/repository"

	"go.uber.org/fx"
)

func NewRepositories() fx.Option {
	return fx.Options(
		fx.Provide(repository.NewUserRepository),
		fx.Provide(repository.NewRefreshTokenRepository),
	)
}
