package api

import (
	"net/http"

	"backend/internal/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handlers) GetAdminOpsOverview(w http.ResponseWriter, r *http.Request) {
	summary, err := h.dashboard.GetOpsOverview(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ops_overview_failed", err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, summary)
}

func (h *Handlers) GetAdminStationSummary(w http.ResponseWriter, r *http.Request) {
	stationID, err := uuid.Parse(chi.URLParam(r, "stationId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_station_id", "Ungültige Station.")
		return
	}

	summary, err := h.dashboard.GetStationDetail(r.Context(), stationID)
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, summary)
}
