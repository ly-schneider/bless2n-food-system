package http

import (
	"net/http"

	"backend/internal/api"
	"backend/internal/auth"
	"backend/internal/generated/api/generated"
	"backend/internal/middleware"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func NewRouter(
	apiHandlers *api.Handlers,
	betterAuthMw *auth.SessionMiddleware,
	deviceAuthMw *auth.DeviceAuthMiddleware,
	securityMw *middleware.SecurityMiddleware,
	systemMw *middleware.SystemDisableMiddleware,
	logger *zap.Logger,
	isDev bool,
) http.Handler {
	// ── SSE router (minimal middleware to preserve flushing) ──────────
	sseRouter := chi.NewRouter()
	sseRouter.Use(securityMw.CORS)
	sseRouter.Use(systemMw.RequireEnabled)
	sseRouter.Get("/v1/inventory/stream", apiHandlers.StreamInventory)

	// ── Main router with full middleware stack ────────────────────────
	r := chi.NewRouter()
	sentryHandler := sentryhttp.New(sentryhttp.Options{Repanic: true})
	r.Use(sentryHandler.Handle)
	r.Use(securityMw.SecurityHeaders)
	r.Use(securityMw.CacheControlForSensitive)
	r.Use(securityMw.CORS)
	r.Use(chiMw.Logger)
	r.Use(chiMw.Recoverer)
	r.Use(chiMw.RequestID)
	r.Use(chiMw.RealIP)
	r.Use(chiMw.Heartbeat("/ping"))
	r.Use(middleware.LogServerErrors)

	// ── Create the generated wrapper ──────────────────────────────────
	wrapper := generated.ServerInterfaceWrapper{
		Handler: apiHandlers,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
	}

	// ── Health check ──────────────────────────────────────────────────
	r.Get("/health", apiHandlers.HealthCheck)

	// ── Android update check (public, no auth) ──────────────────────
	r.Get("/v1/android/latest", apiHandlers.GetAndroidLatestVersion)

	// ── API documentation (Scalar) ───────────────────────────────────
	r.Get("/docs", DocsScalarHandler())
	r.Get("/docs/openapi.json", DocsOpenAPIHandler())

	r.Route("/v1", func(v1 chi.Router) {

		// ── System status (always accessible, no auth) ───────────
		v1.Get("/system/status", apiHandlers.GetSystemStatus)

		// ── Exempt routes (no auth, always accessible) ───────────
		v1.Post("/devices/pairings", wrapper.CreateDevicePairing)
		v1.Get("/devices/pairings/{code}", wrapper.GetDevicePairing)
		v1.Post("/invites/verify", wrapper.VerifyInvite)
		v1.Post("/invites/accept", wrapper.AcceptInvite)
		v1.Post("/payments/webhooks/payrexx", wrapper.HandlePayrexxWebhook)
		v1.Get("/payments/{paymentId}", wrapper.GetPayment)

		// ── Public routes (no auth, blocked when system disabled) ─
		v1.Group(func(pub chi.Router) {
			pub.Use(systemMw.RequireEnabled)
			pub.Post("/auth/otp-email", wrapper.SendOtpEmail)
			pub.Get("/products", wrapper.ListProducts)
			pub.Get("/products/{productId}", wrapper.GetProduct)
			pub.Get("/orders/{orderId}", wrapper.GetOrder)
			pub.Get("/menus", wrapper.ListMenus)
			pub.Get("/menus/{menuId}", wrapper.GetMenu)
			pub.Get("/categories", wrapper.ListCategories)
			pub.Get("/categories/{categoryId}", wrapper.GetCategory)
		})

		// ── Device-authenticated routes (blocked when disabled) ──
		v1.Group(func(station chi.Router) {
			station.Use(systemMw.RequireEnabled)
			station.Use(deviceAuthMw.RequireDevice(auth.DeviceTypeStation))
			station.Get("/stations/me", wrapper.GetCurrentStation)
			station.Post("/stations/redeem", wrapper.RedeemAtStation)
		})

		v1.Group(func(pos chi.Router) {
			pos.Use(systemMw.RequireEnabled)
			pos.Use(deviceAuthMw.RequireDevice(auth.DeviceTypePOS))
			pos.Get("/pos/me", wrapper.GetCurrentPos)
			pos.Get("/club100/people", wrapper.ListClub100People)
			pos.Get("/club100/remaining/{elvantoPersonId}", wrapper.GetClub100Remaining)
		})

		// ── Orders (anonymous allowed, blocked when disabled) ────
		v1.Group(func(orders chi.Router) {
			orders.Use(systemMw.RequireEnabled)
			orders.Use(betterAuthMw.OptionalAuth())
			orders.Post("/orders", wrapper.CreateOrder)
			orders.Get("/orders/{orderId}/payment", wrapper.GetOrderPayment)
			orders.Post("/orders/{orderId}/payment", wrapper.CreateOrderPayment)
		})

		// ── User-authenticated routes (blocked when disabled) ────
		v1.Group(func(authed chi.Router) {
			authed.Use(systemMw.RequireEnabled)
			authed.Use(betterAuthMw.RequireAuth())
			authed.Get("/orders", wrapper.ListOrders)
		})

		// ── Admin routes (auth + admin RBAC + CSRF, NOT blocked) ─
		v1.Group(func(admin chi.Router) {
			admin.Use(betterAuthMw.RequireAuth())
			admin.Use(betterAuthMw.RequirePermission(auth.PermAdminAccess))
			csrf := middleware.NewCSRFMiddleware()
			admin.Use(csrf.Require)

			admin.Post("/products", wrapper.CreateProduct)
			admin.Patch("/products/{productId}", wrapper.UpdateProduct)
			admin.Delete("/products/{productId}", wrapper.DeleteProduct)
			admin.Post("/products/{productId}/image", wrapper.UploadProductImage)
			admin.Delete("/products/{productId}/image", wrapper.DeleteProductImage)
			admin.Get("/products/{productId}/inventory", wrapper.GetProductInventory)
			admin.Get("/products/{productId}/inventory/history", wrapper.GetProductInventoryHistory)
			admin.Patch("/products/{productId}/inventory", wrapper.AdjustProductInventory)

			admin.Post("/categories", wrapper.CreateCategory)
			admin.Patch("/categories/{categoryId}", wrapper.UpdateCategory)
			admin.Delete("/categories/{categoryId}", wrapper.DeleteCategory)

			admin.Post("/menus", wrapper.CreateMenu)
			admin.Patch("/menus/{menuId}", wrapper.UpdateMenu)
			admin.Delete("/menus/{menuId}", wrapper.DeleteMenu)
			admin.Post("/menus/{menuId}/slots", wrapper.CreateMenuSlot)
			admin.Patch("/menus/{menuId}/slots/reorder", wrapper.ReorderMenuSlots)
			admin.Patch("/menus/{menuId}/slots/{slotId}", wrapper.UpdateMenuSlot)
			admin.Delete("/menus/{menuId}/slots/{slotId}", wrapper.DeleteMenuSlot)
			admin.Post("/menus/{menuId}/slots/{slotId}/options", wrapper.AddSlotOption)
			admin.Delete("/menus/{menuId}/slots/{slotId}/options/{optionProductId}", wrapper.RemoveSlotOption)

			admin.Patch("/orders/{orderId}", wrapper.UpdateOrderStatus)

			admin.Get("/stations", wrapper.ListStations)
			admin.Get("/stations/{stationId}", wrapper.GetStation)
			admin.Get("/stations/{stationId}/products", wrapper.ListStationProducts)
			admin.Put("/stations/{stationId}/products", wrapper.SetStationProducts)
			admin.Delete("/stations/{stationId}/products/{productId}", wrapper.RemoveStationProduct)
			admin.Delete("/stations/{stationId}", wrapper.RevokeStation)

			admin.Get("/pos/devices", wrapper.ListPosDevices)

			admin.Get("/settings", wrapper.GetSettings)
			admin.Patch("/settings", wrapper.UpdateSettings)

			admin.Get("/jetons", wrapper.ListJetons)
			admin.Post("/jetons", wrapper.CreateJeton)
			admin.Patch("/jetons/{jetonId}", wrapper.UpdateJeton)
			admin.Delete("/jetons/{jetonId}", wrapper.DeleteJeton)

			admin.Get("/users", wrapper.ListUsers)
			admin.Get("/users/{userId}", wrapper.GetUser)
			admin.Patch("/users/{userId}", wrapper.UpdateUser)
			admin.Delete("/users/{userId}", wrapper.DeleteUser)

			admin.Get("/devices", wrapper.ListDevices)
			admin.Get("/devices/{deviceId}", wrapper.GetDevice)
			admin.Delete("/devices/{deviceId}", wrapper.RevokeDevice)
			admin.Post("/devices/pairings/{code}", wrapper.CompleteDevicePairing)

			admin.Get("/invites", wrapper.ListInvites)
			admin.Post("/invites", wrapper.CreateInvite)
			admin.Get("/invites/{inviteId}", wrapper.GetInvite)
			admin.Delete("/invites/{inviteId}", wrapper.DeleteInvite)

			admin.Get("/events", wrapper.ListEvents)
		})
	})

	// Compose: SSE bypasses main middleware, everything else uses full stack
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/v1/inventory/stream" {
			sseRouter.ServeHTTP(w, req)
			return
		}
		r.ServeHTTP(w, req)
	})
}
