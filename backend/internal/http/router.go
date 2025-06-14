package http

import (
	"backend/internal/handlers"
	"backend/internal/http/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	auth handlers.AuthHandler,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logging)
	r.Use(middleware.ErrorHandler)

	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Mount("/auth", auth.Routes())
	})

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return r
}
