package http

import (
	"net/http"

	"backend/internal/domain"
	"backend/internal/handler"
	jwtMiddleware "backend/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	httpSwagger "github.com/swaggo/http-swagger"
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
	jwtMw *jwtMiddleware.JWTMiddleware,
	securityMw *jwtMiddleware.SecurityMiddleware,
	enableDocs bool,
) http.Handler {
	r := chi.NewRouter()

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

	// Health check endpoints (Docker/K8s probes)
	r.Get("/healthz", healthHandler.Healthz)
	r.Get("/readyz", healthHandler.Readyz)

	// JWKS endpoint (public access for JWT verification)
	r.Get("/.well-known/jwks.json", jwksHandler.GetJWKS)

	if enableDocs {
		r.Get("/swagger/*", httpSwagger.WrapHandler)
		r.Get("/dev/email/preview/login", devHandler.PreviewLoginEmail)
		r.Get("/dev/email/preview/email-change", devHandler.PreviewEmailChangeEmail)
	}

	r.Route("/v1", func(v1 chi.Router) {
		// Auth routes
		v1.Route("/auth", func(a chi.Router) {
			a.Post("/otp/request", authHandler.RequestOTP)
			a.Post("/otp/verify", authHandler.VerifyOTP)
			a.Post("/refresh", authHandler.Refresh)
			a.Post("/logout", authHandler.Logout)
			a.With(jwtMw.RequireAuth).Get("/sessions", authHandler.Sessions)
			a.With(jwtMw.RequireAuth).Post("/sessions/{id}/revoke", authHandler.RevokeSession)
		})

		// Me (requires auth)
		v1.With(jwtMw.RequireAuth).Get("/me", authHandler.Me)

		// User profile
		v1.With(jwtMw.RequireAuth).Method("PUT", "/user", http.HandlerFunc(userHandler.UpdateUser))
		v1.With(jwtMw.RequireAuth).Method("DELETE", "/user", http.HandlerFunc(userHandler.DeleteUser))
		v1.With(jwtMw.RequireAuth).Post("/user/email/confirm", userHandler.ConfirmEmailChange)

		v1.Route("/products", func(product chi.Router) {
			product.Get("/", productHandler.ListProducts)
		})

		// Orders (requires auth)
		v1.Route("/orders", func(orders chi.Router) {
			orders.Use(jwtMw.RequireAuth)
			orders.Get("/", orderHandler.ListMyOrders)
		})

		// Example admin protected route
		v1.Route("/admin", func(admin chi.Router) {
			admin.Use(jwtMw.RequireAuth)
			admin.Use(jwtMw.RequireRole(string(domain.UserRoleAdmin)))
			admin.Get("/ping", adminHandler.Ping)
		})

		v1.Route("/payments", func(pay chi.Router) {
			// Allow optional auth so we can attach user to order if logged in
			pay.Use(jwtMw.OptionalAuth)
			pay.Post("/checkout", paymentHandler.CreateCheckout)
			pay.Post("/webhook", paymentHandler.Webhook)
		})
	})

	return r
}
