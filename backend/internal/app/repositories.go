package app

import (
	"backend/internal/repository"

	"go.uber.org/fx"
)

func NewRepositories() fx.Option {
	return fx.Options(
		fx.Provide(
			repository.NewUserRepository,
			repository.NewOTPTokenRepository,
			repository.NewRefreshTokenRepository,
			repository.NewAdminInviteRepository,
			repository.NewStationRepository,
			repository.NewDeviceRepository,
			repository.NewDeviceRequestRepository,
			repository.NewCategoryRepository,
			repository.NewProductRepository,
			repository.NewProductBundleComponentRepository,
			repository.NewStationProductRepository,
			repository.NewInventoryLedgerRepository,
			repository.NewOrderRepository,
			repository.NewOrderItemRepository,
		),
	)
}
