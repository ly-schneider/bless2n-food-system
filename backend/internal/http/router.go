package http

import (
	"net/http"

	"backend/internal/handler"
	jwtMiddleware "backend/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	authHandler *handler.AuthHandler,
	jwtMw *jwtMiddleware.JWTMiddleware,
	securityMw *jwtMiddleware.SecurityMiddleware,
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

			// Example protected routes (add your application routes here)
			protected.Get("/profile", func(w http.ResponseWriter, r *http.Request) {
				// This is just an example - replace with actual profile handler
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"message": "Protected route accessed successfully"}`))
			})

			// Admin-only routes
			protected.Group(func(admin chi.Router) {
				admin.Use(jwtMw.RequireRole("admin"))

				// Example admin route
				admin.Get("/admin/dashboard", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"message": "Admin dashboard accessed"}`))
				})
			})
		})
	})

	return r
}
