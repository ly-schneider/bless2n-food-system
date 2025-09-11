package main

import (
	_ "backend/docs"
	"backend/internal/app"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// @title Bless2n Food System API
// @version 1.0
// @description Internal BlessThun Food System API
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

		fx.Invoke(app.StartHTTPServer),
	).Run()
}
