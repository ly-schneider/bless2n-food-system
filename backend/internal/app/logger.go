package app

import (
	"backend/internal/config"
	"backend/internal/logger"
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type loggerOut struct {
	fx.Out
	Log *zap.Logger
}

func NewLogger(cfg config.Config, lc fx.Lifecycle) (loggerOut, error) {
	// Initialize logger immediately to avoid nil logger issues
	if err := logger.Init(cfg.Logger); err != nil {
		return loggerOut{}, err
	}

	// Flush on shutdown
	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return logger.Sync()
		},
	})

	// Ensure we never return a nil logger
	if logger.L == nil {
		if err := logger.InitDefault(); err != nil {
			return loggerOut{}, err
		}
	}

	return loggerOut{Log: logger.L.Desugar()}, nil
}
