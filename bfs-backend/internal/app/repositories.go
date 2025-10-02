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
            repository.NewEmailChangeTokenRepository,
            repository.NewRefreshTokenRepository,
            repository.NewIdentityLinkRepository,
            repository.NewAdminInviteRepository,
            repository.NewStationRepository,
            repository.NewStationRequestRepository,
            repository.NewCategoryRepository,
            repository.NewProductRepository,
            repository.NewMenuSlotRepository,
            repository.NewMenuSlotItemRepository,
            repository.NewStationProductRepository,
            repository.NewInventoryLedgerRepository,
            repository.NewOrderRepository,
            repository.NewOrderItemRepository,
        ),
    )
}
