package api

import (
	"net/http"

	"backend/internal/response"

	"go.uber.org/zap"
)

func (h *Handlers) GetAndroidLatestVersion(w http.ResponseWriter, r *http.Request) {
	release, err := h.androidUpdate.GetLatestRelease(r.Context())
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
