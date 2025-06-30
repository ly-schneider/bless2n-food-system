package main

import (
	"backend/internal/app"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			app.NewConfig,
			app.ProvideAppConfig,
			app.NewLogger,
			app.NewDB,
			app.NewAsynqClient,
			app.NewRouter,
		),
		app.NewRedisServices(),
		app.NewRepositories(),
		app.NewServices(),
		app.NewHandlers(),
		fx.Invoke(app.StartHTTPServer),
	).Run()
}
