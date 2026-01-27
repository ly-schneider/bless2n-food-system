package app

import (
	"net/http"

	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/handler"
	httpRouter "backend/internal/http"
	"backend/internal/middleware"
	"backend/internal/postgres"
	"backend/internal/repository"
	"backend/internal/service"
)

func NewRouter(
	// Handlers (deprecated auth handlers will be removed)
	authHandler *handler.AuthHandler,
	devHandler *handler.DevHandler,
	adminHandler *handler.AdminHandler,
	userHandler *handler.UserHandler,
	orderHandler *handler.OrderHandler,
	stationHandler *handler.StationHandler,
	posHandler *handler.POSHandler,
	categoryHandler *handler.CategoryHandler,
	productHandler *handler.ProductHandler,
	paymentHandler *handler.PaymentHandler,
	redemptionHandler *handler.RedemptionHandler,
	healthHandler *handler.HealthHandler,
	jwksHandler *handler.JWKSHandler,
	// Middleware
	jwtMw *middleware.JWTMiddleware,
	neonAuthMw *auth.NeonAuthMiddleware,
	securityMw *middleware.SecurityMiddleware,
	// MongoDB repositories (deprecated)
	productRepo repository.ProductRepository,
	inventoryRepo repository.InventoryLedgerRepository,
	auditRepo repository.AuditRepository,
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
	userRepo repository.UserRepository,
	menuSlotRepo repository.MenuSlotRepository,
	menuSlotItemRepo repository.MenuSlotItemRepository,
	categoryRepo repository.CategoryRepository,
	adminInviteRepo repository.AdminInviteRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	stationRepo repository.StationRepository,
	posDeviceRepo repository.PosDeviceRepository,
	stationProductRepo repository.StationProductRepository,
	// PostgreSQL repositories
	pgCategoryRepo postgres.CategoryRepository,
	pgProductRepo postgres.ProductRepository,
	pgJetonRepo postgres.JetonRepository,
	pgMenuSlotRepo postgres.MenuSlotRepository,
	pgMenuSlotOptionRepo postgres.MenuSlotOptionRepository,
	pgDeviceRepo postgres.DeviceRepository,
	pgDeviceProductRepo postgres.DeviceProductRepository,
	pgPosSettingsRepo postgres.PosSettingsRepository,
	pgOrderRepo postgres.OrderRepository,
	pgOrderPaymentRepo postgres.OrderPaymentRepository,
	pgOrderLineRepo postgres.OrderLineRepository,
	pgOrderLineRedemptionRepo postgres.OrderLineRedemptionRepository,
	pgInventoryLedgerRepo postgres.InventoryLedgerRepository,
	// Services
	posConfig service.POSConfigService,
	emailSvc service.EmailService,
	jwtSvc service.JWTService,
	cfg config.Config,
) http.Handler {
	isDev := cfg.App.AppEnv == "dev"

	return httpRouter.NewRouter(
		authHandler, devHandler, adminHandler, userHandler, orderHandler, stationHandler, posHandler, categoryHandler, productHandler, paymentHandler, redemptionHandler, healthHandler, jwksHandler,
		jwtMw, neonAuthMw, securityMw,
		productRepo, inventoryRepo, auditRepo, orderRepo, orderItemRepo, userRepo, menuSlotRepo, menuSlotItemRepo, categoryRepo, adminInviteRepo, refreshTokenRepo, stationRepo, posDeviceRepo, stationProductRepo,
		pgCategoryRepo, pgProductRepo, pgJetonRepo, pgMenuSlotRepo, pgMenuSlotOptionRepo, pgDeviceRepo, pgDeviceProductRepo, pgPosSettingsRepo, pgOrderRepo, pgOrderPaymentRepo, pgOrderLineRepo, pgOrderLineRedemptionRepo, pgInventoryLedgerRepo,
		posConfig, emailSvc,
		jwtSvc, isDev,
	)
}
