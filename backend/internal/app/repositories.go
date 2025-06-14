package app

import (
	"backend/internal/repository"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

func NewRepositories(db *gorm.DB) fx.Option {
	return fx.Options(
		fx.Provide(repository.NewUserRepository),
		fx.Provide(repository.NewRoleRepository),
		fx.Provide(repository.NewAuditLogRepository),
		fx.Provide(repository.NewRefreshTokenRepository),
	)
}
