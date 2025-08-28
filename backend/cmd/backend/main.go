package main

import (
	docs "backend/docs"
	"backend/internal/app"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// @title Backend API
// @version 1.0
// @description Internal API for our backend
// @BasePath /v1
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	fx.New(
		fx.Provide(
			app.NewConfig,
			app.ProvideAppConfig,
			app.NewLogger,
			app.NewDB,
			app.NewRouter,
		),
		fx.WithLogger(func(lc fx.Lifecycle, l *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: l}
		}),

		app.NewRepositories(),
		app.NewServices(),
		app.NewHandlers(),

		fx.Invoke(func() {
			docs.SwaggerInfo.BasePath = "/v1"
		}),

		fx.Invoke(app.StartHTTPServer),
	).Run()
}
