package app

import (
    "net/http"

    "backend/internal/config"
    "backend/internal/handler"
    httpRouter "backend/internal/http"
    "backend/internal/middleware"
    "backend/internal/repository"
    "backend/internal/service"
)

func NewRouter(
    authHandler *handler.AuthHandler,
    devHandler *handler.DevHandler,
    adminHandler *handler.AdminHandler,
    userHandler *handler.UserHandler,
    orderHandler *handler.OrderHandler,
    stationHandler *handler.StationHandler,
    categoryHandler *handler.CategoryHandler,
    productHandler *handler.ProductHandler,
    paymentHandler *handler.PaymentHandler,
    redemptionHandler *handler.RedemptionHandler,
    healthHandler *handler.HealthHandler,
    jwksHandler *handler.JWKSHandler,
    jwtMw *middleware.JWTMiddleware,
    securityMw *middleware.SecurityMiddleware,
    productRepo repository.ProductRepository,
    inventoryRepo repository.InventoryLedgerRepository,
    auditRepo repository.AuditRepository,
    orderRepo repository.OrderRepository,
    userRepo repository.UserRepository,
    menuSlotRepo repository.MenuSlotRepository,
    menuSlotItemRepo repository.MenuSlotItemRepository,
    categoryRepo repository.CategoryRepository,
    adminInviteRepo repository.AdminInviteRepository,
    refreshTokenRepo repository.RefreshTokenRepository,
    emailSvc service.EmailService,
    cfg config.Config,
) http.Handler {
    enableDocs := cfg.App.AppEnv != "prod"
    return httpRouter.NewRouter(
        authHandler, devHandler, adminHandler, userHandler, orderHandler, stationHandler, categoryHandler, productHandler, paymentHandler, redemptionHandler, healthHandler, jwksHandler,
        jwtMw, securityMw,
        productRepo, inventoryRepo, auditRepo, orderRepo, userRepo, menuSlotRepo, menuSlotItemRepo, categoryRepo, adminInviteRepo, refreshTokenRepo, emailSvc,
        enableDocs,
    )
}
