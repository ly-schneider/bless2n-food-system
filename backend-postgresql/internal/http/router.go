package http

import (
	"backend/internal/handler"
	"backend/internal/http/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	auth handler.AuthHandler,
	user handler.UserHandler,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logging)
	r.Use(middleware.ErrorHandler)
	r.Use(middleware.JWT)
	r.Use(middleware.CORS)
	r.Use(middleware.PublicIP)
	r.Use(middleware.UserAgent)

	r.Route("/v1", func(v1 chi.Router) {
		v1.Mount("/auth", auth.Routes())
		v1.Mount("/users", user.Routes())
	})

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return r
}
