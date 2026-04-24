package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"backend/internal/auth"
	"backend/internal/response"
	"backend/internal/service"

	"github.com/google/uuid"
)

type redeemCampaignRequest struct {
	ClaimToken string `json:"claimToken"`
}

// RedeemCampaignAtStation redeems a shared campaign QR at the current station.
// POST /v1/stations/redeem-campaign
func (h *Handlers) RedeemCampaignAtStation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	deviceID, ok := auth.GetDeviceID(ctx)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Device authentication required")
		return
	}

	var body redeemCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}
	token, err := uuid.Parse(body.ClaimToken)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_claim_token", err.Error())
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")

	result, err := h.volunteers.RedeemSharedQR(ctx, token, deviceID, idemKey)
	if err != nil {
		if errors.Is(err, service.ErrVolunteerCampaignNotFound) ||
			errors.Is(err, service.ErrVolunteerCampaignInactive) ||
			errors.Is(err, service.ErrVolunteerCampaignOutsideValid) ||
			errors.Is(err, service.ErrVolunteerMaxRedemptionsReached) ||
			errors.Is(err, service.ErrVolunteerCampaignHasNoProducts) {
			h.writeVolunteerError(w, err)
			return
		}
		writeError(w, http.StatusBadRequest, "redeem_failed", err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"orderId":         result.OrderID.String(),
		"redemptionCount": result.RedemptionCount,
		"maxRedemptions":  result.MaxRedemptions,
		"station":         result.StationResult,
	})
}
