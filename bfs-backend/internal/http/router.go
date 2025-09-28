package http

import (
	"net/http"

	"backend/internal/handler"
	jwtMiddleware "backend/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(
    authHandler *handler.AuthHandler,
    adminHandler *handler.AdminHandler,
    userHandler *handler.UserHandler,
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
	}

    r.Route("/v1", func(v1 chi.Router) {
        v1.Route("/products", func(product chi.Router) {
            product.Get("/", productHandler.ListProducts)
        })

        // future: orders read endpoints

        v1.Route("/payments", func(pay chi.Router) {
            // Allow optional auth so we can attach user to order if logged in
            pay.Use(jwtMw.OptionalAuth)
            pay.Post("/checkout", paymentHandler.CreateCheckout)
            pay.Post("/webhook", paymentHandler.Webhook)
        })
    })

    return r
}
