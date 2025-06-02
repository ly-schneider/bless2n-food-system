package queue

import (
	"backend/internal/config"

	"github.com/hibiken/asynq"
)

func NewAsynqClient(cfg config.Config) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Redis.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
}
