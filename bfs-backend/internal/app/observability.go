package app

import (
	"context"
	"time"

	"backend/internal/config"
	"backend/internal/observability"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/fx"
)

// SetupObservability bootstraps tracing and wires shutdown into the Fx lifecycle.
func SetupObservability(lc fx.Lifecycle, cfg config.Config) error {
	var shutdown = func(context.Context) error { return nil }

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			obsCfg := observability.NewConfigFromEnv()
			if cfg.App.AppEnv != "" {
				obsCfg.ResourceAttributes = append(
					obsCfg.ResourceAttributes,
					attribute.String("deployment.environment", cfg.App.AppEnv),
				)
			}

			initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			var err error
			shutdown, err = observability.Init(initCtx, obsCfg)
			return err
		},
		OnStop: func(ctx context.Context) error {
			stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			return shutdown(stopCtx)
		},
	})

	return nil
}
