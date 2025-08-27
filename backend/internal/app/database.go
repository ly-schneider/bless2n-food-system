package app

import (
	"backend/internal/config"
	"backend/internal/database"
	"go.uber.org/zap"
)

func NewDB(cfg config.Config, logger *zap.Logger) (*database.MongoDB, error) {
	return database.NewMongoDB(cfg, logger)
}