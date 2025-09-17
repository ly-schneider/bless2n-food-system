package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"backend/db/seeds/dev"
	"backend/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	var (
		reset = flag.Bool("reset", false, "Reset database before seeding")
		force = flag.Bool("force", false, "Force seeding even if APP_ENV != dev")
	)
	flag.Parse()

	// Initialize logger
	logger := initLogger()
	defer func() { _ = logger.Sync() }()

	// Load configuration
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	logger.Info("Starting MongoDB seeding",
		zap.Bool("reset", *reset),
		zap.Bool("force", *force),
	)

	if err := seedMongo(ctx, cfg, *reset, *force, logger); err != nil {
		logger.Fatal("MongoDB seeding failed", zap.Error(err))
	}

	logger.Info("Seeding completed successfully")
}

func seedMongo(ctx context.Context, cfg config.Config, reset, force bool, logger *zap.Logger) error {
	// Create MongoDB client
	clientOptions := options.Client().ApplyURI(cfg.Mongo.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("Failed to disconnect MongoDB client", zap.Error(err))
		}
	}()

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Configure seeding with fixed healthy amounts
	seedConfig := dev.MongoConfig{
		DatabaseName:       cfg.Mongo.Database,
		ResetBeforeSeeding: reset,
		BaselineDir:        "./db/seeds/dev",
	}

	// Run seeding
	return dev.SeedMongo(ctx, client, seedConfig, logger, force)
}

func initLogger() *zap.Logger {
	// Create a simple development logger for CLI use
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	config.EncoderConfig.TimeKey = "time"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.MessageKey = "msg"
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	logger, err := config.Build()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	return logger
}
