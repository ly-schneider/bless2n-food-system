package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynq"
	"go.uber.org/fx"

	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/handlers"
	"backend/internal/logger"
	"backend/internal/queue"
	"backend/internal/repository"
	"backend/internal/service"
)

func main() {
	// Initialize global logger first
	cfg := config.Load()
	if err := logger.Init(cfg.Logger); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	app := fx.New(
		fx.Provide(
			func() config.Config { return cfg },
			db.New,
			queue.NewAsynqClient,
			repository.NewUserRepository,
			repository.NewRoleRepository,
			repository.NewAuditLogRepository,
			repository.NewRefreshTokenRepository,
			service.NewUserService,
			service.NewRoleService,
			service.NewAuthService,
			service.NewRefreshTokenService,
			handlers.NewAuthHandler,
			newRouter,
		),
		fx.Invoke(registerRoutes, registerAsynqClient),
	)
	app.Run()
}

func newRouter() *chi.Mux { return chi.NewRouter() }

func registerRoutes(lc fx.Lifecycle, r *chi.Mux, cfg config.Config,
	ah handlers.AuthHandler) {
	r.Mount("/auth", ah.Routes())

	server := &http.Server{Addr: ":" + cfg.App.AppPort, Handler: r}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error { go server.ListenAndServe(); return nil },
		OnStop:  func(ctx context.Context) error { return server.Shutdown(ctx) },
	})
}

func registerAsynqClient(lc fx.Lifecycle, client *asynq.Client) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return client.Close()
		},
	})
}
