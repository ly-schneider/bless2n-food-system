package api

import (
	"net/http"

	"backend/internal/response"

	"go.uber.org/zap"
)

func (h *Handlers) GetAndroidLatestVersion(w http.ResponseWriter, r *http.Request) {
	channel := r.URL.Query().Get("channel")
	if channel == "" {
		channel = "production"
	}
	if channel != "staging" && channel != "production" {
		response.WriteError(w, http.StatusBadRequest, "invalid channel")
		return
	}

	release, err := h.androidUpdate.GetLatestRelease(r.Context(), channel)
	if err != nil {
		h.logger.Error("failed to fetch latest android release", zap.Error(err))
		response.WriteError(w, http.StatusBadGateway, "Failed to check for updates")
		return
	}

	if release == nil {
		response.WriteNoContent(w)
		return
	}

	response.WriteJSON(w, http.StatusOK, release)
}
