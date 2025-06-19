package app

import (
	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/http"
	"context"
	h "net/http"
	"time"

	"go.uber.org/fx"
)

type HTTPParams struct {
	fx.In
	Cfg  config.AppConfig
	Auth handler.AuthHandler
	// Add any other handlers here when needed
}

func NewRouter(p HTTPParams) h.Handler {
	return http.NewRouter(p.Auth)
}

type ServerParams struct {
	fx.In
	Router h.Handler
	Cfg    config.AppConfig
}

func StartHTTPServer(lc fx.Lifecycle, p ServerParams) {
	srv := &h.Server{
		Addr:              ":" + p.Cfg.AppPort,
		Handler:           p.Router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go srv.ListenAndServe()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}
