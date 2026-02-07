package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"backend/internal/generated/api/generated"
	"backend/internal/generated/ent"
	"backend/internal/generated/ent/settings"
	"backend/internal/response"
	"backend/internal/service"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// GetSettings returns the current application settings.
// (GET /settings)
func (h *Handlers) GetSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	settingsData, err := h.settings.GetSettingsWithProducts(ctx)
	if err != nil {
		writeEntError(w, err)
		return
	}

	missingJetons, _ := h.products.CountActiveWithoutJeton(ctx)

	apiSettings := toAPISettings(settingsData, int(missingJetons))
	response.WriteJSON(w, http.StatusOK, apiSettings)
}

// UpdateSettings updates the application settings.
// (PATCH /settings)
func (h *Handlers) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body generated.SettingsUpdate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	if body.PosMode != nil {
		if err := h.settings.SetPosMode(ctx, settings.PosMode(*body.PosMode)); err != nil {
			var missingErr service.MissingJetonForActiveProductsError
			if errors.As(err, &missingErr) {
				response.WriteJSON(w, http.StatusBadRequest, map[string]any{
					"error":   "missing_jetons",
					"message": "Active products missing jeton",
					"missing": missingErr.Count,
				})
				return
			}
			writeError(w, http.StatusBadRequest, "update_failed", err.Error())
			return
		}
	}

	if body.Club100FreeProductIds != nil || body.Club100MaxRedemptions != nil {
		var productIDs []uuid.UUID
		if body.Club100FreeProductIds != nil {
			productIDs = make([]uuid.UUID, 0, len(*body.Club100FreeProductIds))
			for _, id := range *body.Club100FreeProductIds {
				productIDs = append(productIDs, uuid.UUID(id))
			}
		}
		if err := h.settings.SetClub100Settings(ctx, productIDs, body.Club100MaxRedemptions); err != nil {
			writeError(w, http.StatusBadRequest, "update_failed", err.Error())
			return
		}
	}

	settingsData, err := h.settings.GetSettingsWithProducts(ctx)
	if err != nil {
		writeEntError(w, err)
		return
	}

	missingJetons, _ := h.products.CountActiveWithoutJeton(ctx)
	response.WriteJSON(w, http.StatusOK, toAPISettings(settingsData, int(missingJetons)))
}

func toAPISettings(e *ent.Settings, missingJetons int) generated.Settings {
	s := generated.Settings{
		Id:                    e.ID,
		PosMode:               generated.PosFulfillmentMode(e.PosMode),
		Club100MaxRedemptions: e.Club100MaxRedemptions,
		UpdatedAt:             ptr(e.UpdatedAt),
	}

	if missingJetons > 0 {
		s.MissingJetons = ptr(missingJetons)
	}

	if e.Edges.Club100FreeProducts != nil {
		ids := make([]openapi_types.UUID, 0, len(e.Edges.Club100FreeProducts))
		for _, p := range e.Edges.Club100FreeProducts {
			ids = append(ids, openapi_types.UUID(p.ID))
		}
		s.Club100FreeProductIds = &ids
	}

	return s
}
