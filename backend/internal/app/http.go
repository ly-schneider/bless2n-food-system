package app

import (
	"backend/internal/config"
	"backend/internal/handlers"
	"backend/internal/http"
	"context"
	h "net/http"
	"time"

	"go.uber.org/fx"
)

type HTTPParams struct {
	fx.In
	Router h.Handler
	Cfg    config.AppConfig
	Auth   handlers.AuthHandler
}

func NewRouter(p HTTPParams) h.Handler { return http.NewRouter(p.Auth) }

func StartHTTPServer(lc fx.Lifecycle, p HTTPParams) {
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
