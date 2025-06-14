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
	Lc  fx.Lifecycle `optional:"true"`
}

func NewLogger(cfg config.Config, lc fx.Lifecycle) (loggerOut, error) {
	if err := logger.Init(cfg.Logger); err != nil {
		return loggerOut{}, err
	}
	// Flush on shutdown
	lc.Append(fx.Hook{OnStop: func(_ context.Context) error { return logger.L.Sync() }})
	return loggerOut{Log: logger.L.Desugar()}, nil
}
