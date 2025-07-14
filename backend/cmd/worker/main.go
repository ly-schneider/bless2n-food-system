package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"backend/internal/config"
	"backend/internal/jobs"
	"backend/internal/logger"
	"backend/internal/repository"
	"backend/internal/service"
)

func main() {
	cfg := config.Load()

	if err := logger.Init(cfg.Logger); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.L.Infow("Starting worker", "env", cfg.App.AppEnv)

	// Initialize database connection
	db, err := gorm.Open(postgres.Open(cfg.DB.GetDSN()), &gorm.Config{})
	if err != nil {
		logger.L.Fatalw("Failed to connect to database", "error", err)
	}

	// Initialize repositories
	verificationTokenRepo := repository.NewVerificationTokenRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	emailService := service.NewEmailService(cfg.Mailgun, logger.GetLogger())
	verificationService := service.NewVerificationService(verificationTokenRepo, userRepo, emailService, logger.GetLogger())

	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.Redis.GetRedisAddr(),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	handlers := jobs.NewJobHandlers(logger.GetLogger(), verificationService)

	mux := asynq.NewServeMux()
	mux.HandleFunc(jobs.TypeEmailVerification, handlers.HandleEmailVerification)

	logger.L.Infow("Registered task handlers", "handlers", []string{
		jobs.TypeEmailVerification,
	})

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
	srv.Shutdown()
	logger.L.Info("Worker stopped gracefully")
}
