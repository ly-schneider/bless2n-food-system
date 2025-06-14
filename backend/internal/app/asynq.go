package app

import (
	"backend/internal/config"
	"backend/internal/queue"

	"github.com/hibiken/asynq"
)

func NewAsynqClient(cfg config.Config) *asynq.Client {
	return queue.NewAsynqClient(cfg.Redis)
}
