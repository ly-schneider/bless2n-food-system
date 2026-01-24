package app

import (
	"backend/internal/postgres"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

// NewPostgresRepositories returns an fx.Option that provides all PostgreSQL repositories.
func NewPostgresRepositories() fx.Option {
	return fx.Options(
		fx.Provide(
			func(db *gorm.DB) postgres.CategoryRepository {
				return postgres.NewCategoryRepository(db)
			},
			func(db *gorm.DB) postgres.ProductRepository {
				return postgres.NewProductRepository(db)
			},
			func(db *gorm.DB) postgres.JetonRepository {
				return postgres.NewJetonRepository(db)
			},
			func(db *gorm.DB) postgres.MenuSlotRepository {
				return postgres.NewMenuSlotRepository(db)
			},
			func(db *gorm.DB) postgres.MenuSlotOptionRepository {
				return postgres.NewMenuSlotOptionRepository(db)
			},
			func(db *gorm.DB) postgres.DeviceRepository {
				return postgres.NewDeviceRepository(db)
			},
			func(db *gorm.DB) postgres.DeviceProductRepository {
				return postgres.NewDeviceProductRepository(db)
			},
			func(db *gorm.DB) postgres.PosSettingsRepository {
				return postgres.NewPosSettingsRepository(db)
			},
			func(db *gorm.DB) postgres.OrderRepository {
				return postgres.NewOrderRepository(db)
			},
			func(db *gorm.DB) postgres.OrderPaymentRepository {
				return postgres.NewOrderPaymentRepository(db)
			},
			func(db *gorm.DB) postgres.OrderLineRepository {
				return postgres.NewOrderLineRepository(db)
			},
			func(db *gorm.DB) postgres.OrderLineRedemptionRepository {
				return postgres.NewOrderLineRedemptionRepository(db)
			},
			func(db *gorm.DB) postgres.InventoryLedgerRepository {
				return postgres.NewInventoryLedgerRepository(db)
			},
			func(db *gorm.DB) *postgres.TxManager {
				return postgres.NewTxManager(db)
			},
		),
	)
}
