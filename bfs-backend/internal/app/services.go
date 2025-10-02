package app

import (
	"backend/internal/config"
	"backend/internal/service"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewServices() fx.Option {
    return fx.Options(
        fx.Provide(
            service.NewEmailService,
            NewJWTService,
            service.NewAuthService,
            service.NewFederatedAuthService,
            service.NewAdminService,
            service.NewUserService,
            service.NewStationService,
            service.NewCategoryService,
            service.NewProductService,
            service.NewOrderService,
            service.NewPaymentService,
        ),
    )
}

func NewJWTService(cfg config.Config, logger *zap.Logger) service.JWTService {
	logger.Info("Using file-based JWT service")
	return service.NewJWTService(cfg.App.JWTPrivPEMPath, cfg.App.JWTPubPEMPath, cfg.App.JWTIssuer)
}
