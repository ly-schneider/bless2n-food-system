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
	adminHandler *handler.AdminHandler,
	userHandler *handler.UserHandler,
	stationHandler *handler.StationHandler,
	categoryHandler *handler.CategoryHandler,
	productHandler *handler.ProductHandler,
	orderHandler *handler.OrderHandler,
	redemptionHandler *handler.RedemptionHandler,
	jwtMw *middleware.JWTMiddleware,
	securityMw *middleware.SecurityMiddleware,
	cfg config.Config,
) http.Handler {
	enableDocs := cfg.App.AppEnv != "prod"
	return httpRouter.NewRouter(authHandler, adminHandler, userHandler, stationHandler, categoryHandler, productHandler, orderHandler, redemptionHandler, jwtMw, securityMw, enableDocs)
}
