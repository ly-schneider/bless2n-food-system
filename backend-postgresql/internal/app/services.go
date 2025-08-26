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
			service.NewAuthService,
			NewEmailService,
			fx.Annotate(service.NewVerificationService, fx.As(new(service.VerificationService))),
		),
	)
}

func NewEmailService(cfg config.Config, logger *zap.Logger) service.EmailService {
	return service.NewEmailService(cfg.Mailgun, logger)
}
