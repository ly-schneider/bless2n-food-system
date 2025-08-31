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

	if enableDocs {
		r.Get("/swagger/*", httpSwagger.WrapHandler)
	}

	r.Route("/v1", func(v1 chi.Router) {
		// Auth routes (public)
		v1.Route("/auth", func(auth chi.Router) {
			auth.Post("/register/customer", authHandler.RegisterCustomer)
			auth.Post("/verify-otp", authHandler.VerifyOTP)
			auth.Post("/resend-otp", authHandler.ResendOTP)
			auth.Post("/request-login-otp", authHandler.RequestLoginOTP)
			auth.Post("/login", authHandler.Login)
			auth.Post("/refresh", authHandler.RefreshToken)
			auth.Post("/logout", authHandler.Logout)
		})

		// Protected routes - require authentication
		v1.Group(func(protected chi.Router) {
			protected.Use(jwtMw.RequireAuth)

			// User routes
			protected.Route("/users", func(user chi.Router) {
				user.Get("/profile", userHandler.GetProfile)
				user.Put("/profile", userHandler.UpdateProfile)
			})

			// Station routes (public access for viewing)
			protected.Route("/stations", func(station chi.Router) {
				station.Get("/", stationHandler.ListStations)
				station.Get("/{id}", stationHandler.GetStation)
				station.Get("/{id}/products", stationHandler.GetStationProducts)
			})

			// Admin-only routes
			protected.Group(func(admin chi.Router) {
				admin.Use(jwtMw.RequireRole("admin"))

				admin.Route("/admin", func(adminRoutes chi.Router) {
					// Customer management
					adminRoutes.Get("/customers", adminHandler.ListCustomers)
					adminRoutes.Put("/customers/{id}/ban", adminHandler.BanCustomer)
					adminRoutes.Delete("/customers/{id}", adminHandler.DeleteCustomer)
					adminRoutes.Post("/invites", adminHandler.InviteAdmin)

					// Station management
					adminRoutes.Post("/stations", stationHandler.CreateStation)
					adminRoutes.Put("/stations/{id}/approve", stationHandler.ApproveStation)
					adminRoutes.Post("/stations/{id}/products", stationHandler.AssignProductsToStation)
					adminRoutes.Delete("/stations/{id}/products/{productId}", stationHandler.RemoveProductFromStation)
				})
			})
		})
	})

	return r
}
