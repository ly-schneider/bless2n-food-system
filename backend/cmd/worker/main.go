package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"

	"backend/internal/config"
	"backend/internal/logger"
)

type productCreatedPayload struct {
	ProductID uint `json:"product_id"`
}

func handleProductCreated(ctx context.Context, t *asynq.Task) error {
	var p productCreatedPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		logger.L.Errorw("Failed to unmarshal product created payload", "error", err, "payload", string(t.Payload()))
		return asynq.SkipRetry // bad payload – give up
	}

	logger.L.Infow("Processing product created event", "product_id", p.ProductID, "task_id", t.ResultWriter().TaskID())

	// TODO: send e-mail, fire webhook, push analytics, etc.
	logger.L.Infow("Product created event processed successfully", "product_id", p.ProductID)
	return nil
}

func main() {
	// 1. load env / .env
	cfg := config.Load()

	// 2. initialize global logger
	if err := logger.Init(cfg.Logger); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.L.Infow("Starting worker", "env", cfg.App.AppEnv)

	// 3. build Asynq server
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.Redis.GetRedisAddr(),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		},
		asynq.Config{
			Concurrency: 10,
			Queues:      map[string]int{"default": 1},
		},
	)

	// 4. register handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc("product:created", handleProductCreated)

	logger.L.Infow("Registered task handlers", "handlers", []string{"product:created"})

	// 5. run with graceful shutdown
	go func() {
		logger.L.Info("Worker server starting...")
		if err := srv.Run(mux); err != nil {
			logger.L.Fatalw("Asynq server error", "error", err)
		}
	}()

	// Ctrl-C / SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.L.Info("Shutdown signal received, stopping worker...")
	srv.Shutdown() // drains in-flight tasks
	logger.L.Info("Worker stopped gracefully")
}
