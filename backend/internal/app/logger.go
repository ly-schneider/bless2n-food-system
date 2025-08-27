package app

import (
	"backend/internal/config"
	"go.uber.org/zap"
)

func NewLogger(cfg config.Config) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	if cfg.Logger.Development {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		return nil, err
	}

	return logger, nil
}