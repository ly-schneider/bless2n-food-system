package app

import (
	"net/http"

	"backend/internal/handler"
	httpRouter "backend/internal/http"
	"backend/internal/middleware"
)

func NewRouter(
	authHandler *handler.AuthHandler,
	jwtMw *middleware.JWTMiddleware,
	securityMw *middleware.SecurityMiddleware,
) http.Handler {
	return httpRouter.NewRouter(authHandler, jwtMw, securityMw)
}