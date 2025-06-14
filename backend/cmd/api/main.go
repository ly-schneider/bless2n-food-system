package main

import (
	"backend/internal/app"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			app.NewConfig, // viper/envconfig
			app.NewLogger, // zap
			app.NewDB,     // gorm
			app.NewAsynqClient,
			app.NewRepositories,
			app.NewServices,
			app.NewHandlers,
			app.NewRouter,
		),
		fx.Invoke(app.StartHTTPServer), // hides graceful shutdown logic
	).Run()
}
