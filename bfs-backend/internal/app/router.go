package app

import (
	"net/http"

	"backend/internal/api"
	"backend/internal/auth"
	"backend/internal/config"
	httpRouter "backend/internal/http"
	"backend/internal/middleware"

	"go.uber.org/zap"
)

func NewRouter(
	apiHandlers *api.Handlers,
	betterAuthMw *auth.SessionMiddleware,
	deviceAuthMw *auth.DeviceAuthMiddleware,
	securityMw *middleware.SecurityMiddleware,
	logger *zap.Logger,
	cfg config.Config,
) http.Handler {
	isDev := cfg.App.AppEnv == "dev"

	return httpRouter.NewRouter(
		apiHandlers,
		betterAuthMw, deviceAuthMw, securityMw,
		logger, isDev,
	)
}
