package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"backend/internal/generated/api/generated"
	"backend/internal/response"
	"backend/internal/service"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ListJetons returns all jetons.
// (GET /jetons)
func (h *Handlers) ListJetons(w http.ResponseWriter, r *http.Request) {
	jetons, err := h.settings.ListJetons(r.Context())
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, generated.JetonList{
		Items: toAPIJetons(jetons),
	})
}

// CreateJeton creates a new jeton.
// (POST /jetons)
func (h *Handlers) CreateJeton(w http.ResponseWriter, r *http.Request) {
	var body generated.JetonCreate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	jeton, err := h.settings.CreateJeton(r.Context(), body.Name, body.Color)
	if err != nil {
		writeError(w, http.StatusBadRequest, "create_failed", err.Error())
		return
	}
	response.WriteJSON(w, http.StatusCreated, toAPIJeton(jeton))
}

// UpdateJeton updates an existing jeton.
// (PATCH /jetons/{jetonId})
func (h *Handlers) UpdateJeton(w http.ResponseWriter, r *http.Request, jetonId openapi_types.UUID) {
	var body generated.JetonUpdate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	// For partial updates, we need the current values. Fetch the jeton first.
	// The service handles merging with current values.
	name := derefStr(body.Name)
	color := derefStr(body.Color)

	jeton, err := h.settings.UpdateJeton(r.Context(), uuid.UUID(jetonId), name, color)
	if err != nil {
		writeError(w, http.StatusBadRequest, "update_failed", err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, toAPIJeton(jeton))
}

// DeleteJeton removes a jeton.
// (DELETE /jetons/{jetonId})
func (h *Handlers) DeleteJeton(w http.ResponseWriter, r *http.Request, jetonId openapi_types.UUID) {
	if err := h.settings.DeleteJeton(r.Context(), uuid.UUID(jetonId)); err != nil {
		var inUse service.JetonInUseError
		if errors.As(err, &inUse) {
			details := map[string]any{"usage": inUse.Count}
			response.WriteJSON(w, http.StatusConflict, generated.Error{
				Code:    "jeton_in_use",
				Message: "Dieser Jeton ist noch Produkten zugewiesen. Bitte entferne zuerst die Zuweisungen.",
				Details: &details,
			})
			return
		}
		writeEntError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
