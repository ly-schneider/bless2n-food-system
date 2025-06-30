package app

import (
	"backend/internal/config"
	"backend/internal/service"

	"go.uber.org/fx"
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

func NewEmailService(cfg config.Config) service.EmailService {
	return service.NewEmailService(cfg.Mailgun)
}
