package app

import (
	"backend/internal/inventory"
	"backend/internal/repository"
	"backend/internal/service"

	"go.uber.org/fx"
)

func NewServices() fx.Option {
	return fx.Options(
		fx.Provide(
			service.NewPaymentService,
			service.NewSettingsService,
			service.NewProductService,
			service.NewCategoryService,
			service.NewOrderService,
			service.NewStationService,
			service.NewPOSService,
			repository.NewIdempotencyRepository,
			service.NewEmailService,
			service.NewAdminInviteService,
			service.NewUserService,
			service.NewDeviceService,
			inventory.NewHub,
			service.NewElvantoService,
			service.NewClub100Service,
			service.NewAndroidUpdateService,
		),
	)
}
