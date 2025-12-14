package http

import (
	"net/http"

	"backend/internal/domain"
	"backend/internal/handler"
	jwtMiddleware "backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	httpSwagger "github.com/swaggo/http-swagger/v2"
	"github.com/swaggo/swag"
)

func NewRouter(
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
	jwtMw *jwtMiddleware.JWTMiddleware,
	securityMw *jwtMiddleware.SecurityMiddleware,
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
	posConfig service.POSConfigService,
	emailSvc service.EmailService,
	jwtSvc service.JWTService,
	isDev bool,
) http.Handler {
	r := chi.NewRouter()
	// wire chi URLParam to admin handler helpers
	handler.ChiURLParamFn = chi.URLParam

	// Security middleware (applied first for all requests)
	r.Use(securityMw.SecurityHeaders)
	r.Use(securityMw.CacheControlForSensitive)
	r.Use(securityMw.CORS)

	// Standard middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(jwtMiddleware.LogServerErrors)

	// Health check endpoint
	r.Get("/health", healthHandler.Health)

	// JWKS endpoint (public access for JWT verification)
	r.Get("/.well-known/jwks.json", jwksHandler.GetJWKS)

	// Swagger documentation (available when docs are generated)
	if swag.GetSwagger(swag.Name) != nil {
		r.Get("/swagger/*", httpSwagger.WrapHandler)
	}

	// Dev-only email preview endpoints (only accessible in dev/local environments)
	if isDev {
		r.Get("/dev/email/preview/login", devHandler.PreviewLoginEmail)
		r.Get("/dev/email/preview/email-change", devHandler.PreviewEmailChangeEmail)
		r.Get("/dev/email/preview/admin-invite", devHandler.PreviewAdminInviteEmail)
	}

	r.Route("/v1", func(v1 chi.Router) {
		// Auth routes
		v1.Route("/auth", func(a chi.Router) {
			a.Post("/otp/request", authHandler.RequestOTP)
			a.Post("/otp/verify", authHandler.VerifyOTP)
			a.Post("/google/code", authHandler.GoogleCode)
			a.Post("/refresh", authHandler.Refresh)
			a.Post("/logout", authHandler.Logout)
		})

		// Users resource (self)
		v1.Route("/users", func(users chi.Router) {
			users.Use(jwtMw.RequireAuth)
			users.Route("/me", func(me chi.Router) {
				me.Get("/", userHandler.GetCurrent)
				me.Method("PATCH", "/", http.HandlerFunc(userHandler.UpdateUser))
				me.Method("DELETE", "/", http.HandlerFunc(userHandler.DeleteUser))
				me.Post("/email-change", userHandler.RequestEmailChange)
				me.Post("/email-change/confirm", userHandler.ConfirmEmailChange)
				me.Get("/sessions", authHandler.Sessions)
				me.Delete("/sessions/{id}", authHandler.RevokeSession)
				me.Delete("/sessions", authHandler.RevokeAllSessions)
			})
		})

		v1.Route("/products", func(product chi.Router) {
			product.Get("/", productHandler.ListProducts)
		})

		// Station public endpoints
		v1.Route("/stations", func(st chi.Router) {
			st.Post("/requests", stationHandler.CreateRequest)
			st.Get("/me", stationHandler.Me)
			st.Post("/verify-qr", stationHandler.VerifyQR)
			st.Post("/redeem", stationHandler.Redeem)
		})

		// POS public endpoints (device-gated via X-Pos-Token)
		v1.Route("/pos", func(pos chi.Router) {
			pos.Post("/requests", posHandler.CreateRequest)
			pos.Get("/me", posHandler.Me)
			pos.Post("/orders", posHandler.CreateOrder)
			pos.Post("/orders/{id}/pay-cash", posHandler.PayCash)
			pos.Post("/orders/{id}/pay-card", posHandler.PayCard)
		})

		// Orders: public details by id, and authenticated list
		v1.Route("/orders", func(orders chi.Router) {
			// Public access to order details by id (no auth)
			orders.Get("/{id}", orderHandler.GetPublicByID)
			orders.Get("/{id}/pickup-qr", stationHandler.GetPickupQR)
			// Authenticated list of own orders
			orders.With(jwtMw.RequireAuth).Get("/", orderHandler.ListMyOrders)
		})

		// Admin routes
		v1.Route("/admin", func(admin chi.Router) {
			admin.Use(jwtMw.RequireAuth)
			// Require admin role, but allow immediate effect on role changes by checking DB if token is not yet updated
			admin.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					claims, ok := jwtMiddleware.GetUserFromContext(r.Context())
					if !ok || claims == nil {
						http.Error(w, "Unauthorized", http.StatusUnauthorized)
						return
					}
					if string(claims.Role) == string(domain.UserRoleAdmin) {
						next.ServeHTTP(w, r)
						return
					}
					// Fallback to DB role check for immediate RBAC change effect
					if userRepo != nil {
						if oid, err := bson.ObjectIDFromHex(claims.Subject); err == nil {
							if u, err := userRepo.FindByID(r.Context(), oid); err == nil && u != nil && u.Role == domain.UserRoleAdmin {
								next.ServeHTTP(w, r)
								return
							}
						}
					}
					http.Error(w, "Forbidden", http.StatusForbidden)
				})
			})
			// CSRF required on state changes
			csrf := jwtMiddleware.NewCSRFMiddleware()
			admin.Use(csrf.Require)
			admin.Get("/ping", adminHandler.Ping)

			// Products admin
			ap := handler.NewAdminProductHandler(productRepo, inventoryRepo, auditRepo, menuSlotItemRepo, categoryRepo, posConfig)
			admin.Patch("/products/{id}/price", http.HandlerFunc(ap.PatchPrice))
			admin.Patch("/products/{id}/active", http.HandlerFunc(ap.PatchActive))
			admin.Post("/products/{id}/inventory-adjust", http.HandlerFunc(ap.AdjustInventory))
			admin.Delete("/products/{id}", http.HandlerFunc(ap.DeleteHard))
			admin.Patch("/products/{id}/category", http.HandlerFunc(ap.PatchCategory))
			admin.Patch("/products/{id}/jeton", http.HandlerFunc(ap.PatchJeton))
			admin.Post("/products", http.HandlerFunc(ap.Create))
			// Orders admin
			ao := handler.NewAdminOrderHandler(orderRepo, orderItemRepo, productRepo, auditRepo)
			admin.Get("/orders", http.HandlerFunc(ao.List))
			admin.Get("/orders/{id}", http.HandlerFunc(ao.GetByID))
			admin.Get("/orders/export.csv", http.HandlerFunc(ao.ExportCSV))
			admin.Patch("/orders/{id}/status", http.HandlerFunc(ao.PatchStatus))
			// Users admin
			au := handler.NewAdminUserHandler(userRepo)
			admin.Get("/users", http.HandlerFunc(au.List))
			admin.Get("/users/{id}", http.HandlerFunc(au.GetByID))
			admin.Patch("/users/{id}", http.HandlerFunc(au.PatchProfile))
			admin.Delete("/users/{id}", http.HandlerFunc(au.Delete))
			admin.Patch("/users/{id}/role", http.HandlerFunc(au.PatchRole))
			// Legacy: promote
			admin.Post("/users/{id}/promote", http.HandlerFunc(au.Promote))
			// Sessions admin
			as := handler.NewAdminSessionsHandler(refreshTokenRepo, userRepo)
			admin.Get("/sessions", http.HandlerFunc(as.List))
			admin.Post("/users/{id}/sessions/revoke", http.HandlerFunc(as.RevokeFamily))
			admin.Post("/users/{id}/sessions/revoke-all", http.HandlerFunc(as.RevokeAll))
			// Categories admin
			ac := handler.NewAdminCategoryHandler(categoryRepo, auditRepo)
			admin.Get("/categories", http.HandlerFunc(ac.List))
			admin.Post("/categories", http.HandlerFunc(ac.Create))
			admin.Patch("/categories/{id}", http.HandlerFunc(ac.Update))
			admin.Delete("/categories/{id}", http.HandlerFunc(ac.Delete))
			// Menus admin
			am := handler.NewAdminMenuHandler(productRepo, categoryRepo, menuSlotRepo, menuSlotItemRepo, auditRepo)
			admin.Get("/menus", http.HandlerFunc(am.List))
			admin.Post("/menus", http.HandlerFunc(am.Create))
			admin.Get("/menus/{id}", http.HandlerFunc(am.Get))
			admin.Patch("/menus/{id}", http.HandlerFunc(am.Update)) // legacy generic update
			admin.Patch("/menus/{id}/active", http.HandlerFunc(am.PatchActive))
			admin.Delete("/menus/{id}", http.HandlerFunc(am.DeleteHard))
			admin.Post("/menus/{id}/slots", http.HandlerFunc(am.CreateSlot))
			admin.Patch("/menus/{id}/slots/{slotId}", http.HandlerFunc(am.RenameSlot))
			admin.Patch("/menus/{id}/slots/reorder", http.HandlerFunc(am.ReorderSlots))
			admin.Delete("/menus/{id}/slots/{slotId}", http.HandlerFunc(am.DeleteSlot))
			admin.Post("/menus/{id}/slots/{slotId}/items", http.HandlerFunc(am.AttachItem))
			admin.Delete("/menus/{id}/slots/{slotId}/items/{productId}", http.HandlerFunc(am.DetachItem))
			// Admin invites (admin-only for creating new admins)
			ai := handler.NewAdminInviteHandler(adminInviteRepo, userRepo, auditRepo, emailSvc, jwtSvc, refreshTokenRepo)
			admin.Get("/invites", http.HandlerFunc(ai.List))
			admin.Post("/invites", http.HandlerFunc(ai.Create))
			admin.Delete("/invites/{id}", http.HandlerFunc(ai.Delete))
			admin.Post("/invites/{id}/revoke", http.HandlerFunc(ai.Revoke))
			admin.Post("/invites/{id}/resend", http.HandlerFunc(ai.Resend))

			// Stations admin
			ast := handler.NewAdminStationHandler(stationRepo, stationProductRepo, productRepo, auditRepo, nil)
			admin.Get("/stations/requests", http.HandlerFunc(ast.ListRequests))
			admin.Get("/stations", http.HandlerFunc(ast.ListStations))
			admin.Post("/stations/requests/{id}/approve", http.HandlerFunc(ast.Approve))
			admin.Post("/stations/requests/{id}/reject", http.HandlerFunc(ast.Reject))
			admin.Post("/stations/requests/{id}/revoke", http.HandlerFunc(ast.Revoke))
			admin.Get("/stations/{id}/products", http.HandlerFunc(ast.ListStationProducts))
			admin.Post("/stations/{id}/products", http.HandlerFunc(ast.AssignProducts))
			admin.Delete("/stations/{id}/products/{productId}", http.HandlerFunc(ast.RemoveProduct))

			// POS admin
			apos := handler.NewAdminPOSHandler(posDeviceRepo, posConfig, productRepo)
			admin.Get("/pos/settings", http.HandlerFunc(apos.GetSettings))
			admin.Patch("/pos/settings", http.HandlerFunc(apos.PatchSettings))
			admin.Get("/pos/requests", http.HandlerFunc(apos.ListRequests))
			admin.Post("/pos/requests/{id}/approve", http.HandlerFunc(apos.Approve))
			admin.Post("/pos/requests/{id}/reject", http.HandlerFunc(apos.Reject))
			admin.Post("/pos/requests/{id}/revoke", http.HandlerFunc(apos.Revoke))
			admin.Get("/pos/devices", http.HandlerFunc(apos.ListDevices))
			admin.Patch("/pos/devices/{id}/config", http.HandlerFunc(apos.PatchConfig))
			admin.Get("/pos/jetons", http.HandlerFunc(apos.ListJetons))
			admin.Post("/pos/jetons", http.HandlerFunc(apos.CreateJeton))
			admin.Patch("/pos/jetons/{id}", http.HandlerFunc(apos.UpdateJeton))
			admin.Delete("/pos/jetons/{id}", http.HandlerFunc(apos.DeleteJeton))
		})

		v1.Route("/payments", func(pay chi.Router) {
			// Allow optional auth so we can attach user to order if logged in
			pay.Use(jwtMw.OptionalAuth)
			// Payment Intents (TWINT via Payment Element)
			pay.Post("/create-intent", paymentHandler.CreateIntent)
			pay.Patch("/attach-email", paymentHandler.AttachEmail)
			pay.Get("/{id}", paymentHandler.GetPayment)
			// Stripe webhook receiver
			pay.Post("/webhook", paymentHandler.Webhook)
		})

		// Public invite endpoints
		v1.Post("/invites/accept", http.HandlerFunc(handler.NewAdminInviteHandler(adminInviteRepo, userRepo, auditRepo, emailSvc, jwtSvc, refreshTokenRepo).Accept))
		v1.Post("/invites/verify", http.HandlerFunc(handler.NewAdminInviteHandler(adminInviteRepo, userRepo, auditRepo, emailSvc, jwtSvc, refreshTokenRepo).Verify))
	})

	return r
}
