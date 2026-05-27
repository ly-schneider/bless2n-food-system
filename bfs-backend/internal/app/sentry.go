package app

import (
	"context"
	"os"
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

			// Without an explicit ServerName, sentry.NewMeter() falls back to
			// os.Hostname() on every call — and HTTPMetrics middleware calls
			// NewMeter per request. Resolve it once at startup.
			hostname, _ := os.Hostname()

			err := sentry.Init(sentry.ClientOptions{
				Dsn:              cfg.Sentry.DSN,
				Environment:      cfg.Sentry.Environment,
				Release:          cfg.Sentry.Release,
				ServerName:       hostname,
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
