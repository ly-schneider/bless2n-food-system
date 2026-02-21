package api

import (
	"net/http"

	"backend/internal/response"
)

func (h *Handlers) GetSystemStatus(w http.ResponseWriter, r *http.Request) {
	enabled, err := h.settings.IsSystemEnabled(r.Context())
	if err != nil {
		enabled = true
	}
	response.WriteJSON(w, http.StatusOK, map[string]bool{
		"enabled": enabled,
	})
}
