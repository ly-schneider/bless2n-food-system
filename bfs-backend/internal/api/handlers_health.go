package api

import (
	"net/http"

	"backend/internal/response"
	"backend/internal/version"
)

// HealthCheck returns a simple health check response.
// This is a standalone handler, not part of the ServerInterface.
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": version.Version,
	})
}
