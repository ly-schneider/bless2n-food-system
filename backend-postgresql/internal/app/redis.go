package app

import (
	"backend/internal/config"
	"backend/internal/jobs"
	"backend/internal/redis"

	"github.com/hibiken/asynq"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewRedisServices() fx.Option {
	return fx.Options(
		fx.Provide(
			NewRedisClient,
			NewCacheService,
			NewSessionCacheService,
			NewRateLimiterService,
			NewJobService,
		),
	)
}

func NewRedisClient(cfg config.Config, logger *zap.Logger) (*redis.Client, error) {
	return redis.NewClient(cfg.Redis, logger)
}

func NewCacheService(client *redis.Client, logger *zap.Logger) *redis.CacheService {
	return redis.NewCacheService(client, logger)
}

func NewSessionCacheService(cache *redis.CacheService, logger *zap.Logger) *redis.SessionCacheService {
	return redis.NewSessionCacheService(cache, logger)
}

func NewRateLimiterService(cache *redis.CacheService, logger *zap.Logger) *redis.RateLimiterService {
	return redis.NewRateLimiterService(cache, logger)
}


func NewJobService(client *asynq.Client, logger *zap.Logger) *jobs.JobService {
	return jobs.NewJobService(client, logger)
}