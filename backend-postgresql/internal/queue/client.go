package queue

import (
	"backend/internal/config"

	"github.com/hibiken/asynq"
)

func NewAsynqClient(cfg config.RedisConfig) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}
