package main

import (
	"backend/internal/app"
	"net/http"
	"os"
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// @title Bless2n Food System API
// @version 1.0
// @description Internal BlessThun Food System API
// @schemes https http
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Lightweight container healthcheck: call local /health and exit.
	for _, arg := range os.Args[1:] {
		if arg == "--healthcheck" {
			port := os.Getenv("APP_PORT")
			if port == "" {
				port = "8080"
			}
			url := "http://127.0.0.1:" + port + "/ping"
			client := &http.Client{Timeout: 2 * time.Second}
			resp, err := client.Get(url)
			if err != nil {
				os.Exit(1)
			}
			if resp.StatusCode == http.StatusOK {
				os.Exit(0)
			}
			os.Exit(1)
		}
	}

	fx.New(
		fx.Provide(
			app.NewConfig,
			app.ProvideAppConfig,
			app.NewLogger,
			app.NewDB,
			app.ProvideDatabase,
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
