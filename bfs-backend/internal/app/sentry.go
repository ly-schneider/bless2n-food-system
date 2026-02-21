package app

import (
	"context"
	"time"

	"backend/internal/config"

	"github.com/getsentry/sentry-go"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func SetupSentry(lc fx.Lifecycle, cfg config.Config, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			if cfg.Sentry.DSN == "" {
				logger.Info("sentry disabled (no DSN configured)")
				return nil
			}

			err := sentry.Init(sentry.ClientOptions{
				Dsn:              cfg.Sentry.DSN,
				Environment:      cfg.Sentry.Environment,
				Release:          cfg.Sentry.Release,
				TracesSampleRate: 0.2,
				EnableTracing:    true,
				EnableLogs:       true,
			})
			if err != nil {
				logger.Warn("sentry init failed", zap.Error(err))
				return nil
			}

			logger.Info("sentry initialized", zap.String("environment", cfg.Sentry.Environment))
			return nil
		},
		OnStop: func(context.Context) error {
			sentry.Flush(2 * time.Second)
			return nil
		},
	})
}
