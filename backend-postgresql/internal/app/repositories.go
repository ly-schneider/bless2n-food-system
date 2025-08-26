package app

import (
	"backend/internal/redis"
	"backend/internal/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func NewRepositories() fx.Option {
	return fx.Options(
		fx.Provide(
			NewUserRepository,
			fx.Annotate(repository.NewAuditLogRepository, fx.As(new(repository.AuditLogRepository))),
			fx.Annotate(repository.NewRefreshTokenRepository, fx.As(new(repository.RefreshTokenRepository))),
			fx.Annotate(repository.NewVerificationTokenRepository, fx.As(new(repository.VerificationTokenRepository))),
		),
	)
}

func NewUserRepository(db *gorm.DB, cache *redis.CacheService, logger *zap.Logger) repository.UserRepository {
	baseRepo := repository.NewUserRepository(db)
	return repository.NewCachedUserRepository(baseRepo, cache, logger)
}
