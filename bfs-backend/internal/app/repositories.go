package app

import (
	"backend/internal/repository"

	"go.uber.org/fx"
)

func NewRepositories() fx.Option {
	return fx.Options(
		fx.Provide(
			repository.NewCategoryRepository,
			repository.NewProductRepository,
			repository.NewJetonRepository,
			repository.NewMenuSlotRepository,
			repository.NewMenuSlotOptionRepository,
			repository.NewDeviceRepository,
			repository.NewDeviceProductRepository,
			repository.NewSettingsRepository,
			repository.NewOrderRepository,
			repository.NewOrderPaymentRepository,
			repository.NewOrderLineRepository,
			repository.NewOrderLineRedemptionRepository,
			repository.NewInventoryLedgerRepository,
			repository.NewAdminInviteRepository,
			repository.NewUserRepository,
			repository.NewVerificationRepository,
			repository.NewSessionRepository,
			repository.NewDeviceBindingRepository,
			repository.NewClub100RedemptionRepository,
		),
	)
}
