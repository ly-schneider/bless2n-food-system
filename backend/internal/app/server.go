package app

import (
	"context"
	"net/http"

	"backend/internal/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func StartHTTPServer(lc fx.Lifecycle, cfg config.Config, router http.Handler, logger *zap.Logger) {
	server := &http.Server{
		Addr:    ":" + cfg.App.AppPort,
		Handler: router,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting HTTP server", zap.String("port", cfg.App.AppPort))
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Fatal("failed to start HTTP server", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping HTTP server")
			return server.Shutdown(ctx)
		},
	})
}