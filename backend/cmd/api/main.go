package main

import (
	"backend/internal/app"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			app.NewConfig,
			app.ProvideAppConfig, // Add this line to provide AppConfig
			app.NewLogger,
			app.NewDB,
			app.NewAsynqClient,
			app.NewRouter,
		),
		// Use these fx.Option returning functions directly
		app.NewRepositories(), // Now used as an fx.Option, not as a provider
		app.NewServices(),
		app.NewHandlers(),
		fx.Invoke(app.StartHTTPServer),
	).Run()
}
