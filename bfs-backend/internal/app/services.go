package app

import (
	"backend/internal/config"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/service/auth"

	"go.uber.org/fx"
)

func NewServices() fx.Option {
	return fx.Options(
		fx.Provide(
			service.NewEmailService,
			NewJWTService,
			auth.NewOTPService,
			auth.NewTokenService,
			NewAuthService,
			service.NewAdminService,
			service.NewUserService,
			service.NewStationService,
			service.NewCategoryService,
			service.NewProductService,
			service.NewOrderService,
		),
	)
}

func NewAuthService(
	userRepo repository.UserRepository,
	otpService auth.OTPService,
	tokenService auth.TokenService,
) service.AuthService {
	return auth.NewService(userRepo, otpService, tokenService)
}

func NewJWTService(cfg config.Config) service.JWTService {
	return service.NewJWTService(cfg.App.JWTPrivPEMPath, cfg.App.JWTPubPEMPath, cfg.App.JWTIssuer)
}
