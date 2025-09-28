package app

import (
	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/middleware"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewHandlers() fx.Option {
    return fx.Options(
        fx.Provide(
            handler.NewAuthHandler,
            handler.NewAdminHandler,
            handler.NewUserHandler,
            handler.NewStationHandler,
            handler.NewCategoryHandler,
            handler.NewProductHandler,
            handler.NewPaymentHandler,
            handler.NewRedemptionHandler,
            NewHealthHandler,
            handler.NewJWKSHandler,
            middleware.NewJWTMiddleware,
            NewSecurityMiddleware,
        ),
    )
}

func NewHealthHandler(logger *zap.Logger, db *mongo.Database) *handler.HealthHandler {
	return handler.NewHealthHandler(logger, db, nil)
}

func NewSecurityMiddleware(cfg config.Config) *middleware.SecurityMiddleware {
	return middleware.NewSecurityMiddleware(middleware.SecurityConfig{
		EnableHSTS:     cfg.Security.EnableHSTS,
		EnableCSP:      cfg.Security.EnableCSP,
		TrustedOrigins: cfg.Security.TrustedOrigins,
		AppEnv:         cfg.App.AppEnv,
	})
}
