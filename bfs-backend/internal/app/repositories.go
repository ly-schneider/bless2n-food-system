package app

import (
	"backend/internal/generated/ent"
	"backend/internal/repository"

	"go.uber.org/fx"
)

func newCachedSessionRepository(client *ent.Client) repository.SessionRepository {
	return repository.NewCachedSessionRepository(repository.NewSessionRepository(client))
}

func newCachedDeviceBindingRepository(client *ent.Client) repository.DeviceBindingRepository {
	return repository.NewCachedDeviceBindingRepository(repository.NewDeviceBindingRepository(client))
}

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
			newCachedSessionRepository,
			newCachedDeviceBindingRepository,
			repository.NewClub100RedemptionRepository,
			repository.NewVolunteerCampaignRepository,
			repository.NewVolunteerRedemptionRepository,
		),
	)
}
