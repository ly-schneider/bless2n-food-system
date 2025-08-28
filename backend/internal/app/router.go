package app

import (
	"net/http"

	"backend/internal/config"
	"backend/internal/handler"
	httpRouter "backend/internal/http"
	"backend/internal/middleware"
)

func NewRouter(
	authHandler *handler.AuthHandler,
	jwtMw *middleware.JWTMiddleware,
	securityMw *middleware.SecurityMiddleware,
	cfg config.Config,
) http.Handler {
	enableDocs := cfg.App.AppEnv != "prod"
	return httpRouter.NewRouter(authHandler, jwtMw, securityMw, enableDocs)
}
