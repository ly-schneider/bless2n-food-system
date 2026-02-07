package api

import (
	"encoding/json"
	"net/http"
	"time"

	"backend/internal/auth"
	"backend/internal/generated/api/generated"
	"backend/internal/response"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ListStations returns all station-type devices, optionally filtered by status.
// (GET /stations)
func (h *Handlers) ListStations(w http.ResponseWriter, r *http.Request, params generated.ListStationsParams) {
	ctx := r.Context()

	// Filter devices to STATION type with optional status filter.
	stationType := "STATION"
	var statusStr *string
	if params.Status != nil {
		s := string(*params.Status)
		statusStr = &s
	}

	stations, err := h.devices.ListAll(ctx, &stationType, statusStr)
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, generated.StationList{
		Items: toAPIStations(stations),
	})
}

// GetCurrentStation returns the station associated with the current device auth context.
// (GET /stations/me)
func (h *Handlers) GetCurrentStation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	deviceID, ok := auth.GetDeviceID(ctx)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Device authentication required")
		return
	}

	station, err := h.stations.GetStationByID(ctx, deviceID)
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, toAPIStation(station))
}

// GetStation returns a specific station by ID.
// (GET /stations/{stationId})
func (h *Handlers) GetStation(w http.ResponseWriter, r *http.Request, stationId openapi_types.UUID) {
	station, err := h.stations.GetStationByID(r.Context(), uuid.UUID(stationId))
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, toAPIStation(station))
}

// ListStationProducts returns the products assigned to a station.
// (GET /stations/{stationId}/products)
func (h *Handlers) ListStationProducts(w http.ResponseWriter, r *http.Request, stationId openapi_types.UUID) {
	ctx := r.Context()

	// Get the station to verify it exists and load its product assignments.
	station, err := h.stations.GetStationByID(ctx, uuid.UUID(stationId))
	if err != nil {
		writeEntError(w, err)
		return
	}

	// The station's device products are mapped via the toAPIStation helper,
	// which loads the DeviceProducts edge.
	apiStation := toAPIStation(station)
	var items []generated.StationProduct
	if apiStation.Products != nil {
		items = *apiStation.Products
	} else {
		items = []generated.StationProduct{}
	}

	response.WriteJSON(w, http.StatusOK, generated.StationProductList{
		Items: items,
	})
}

// SetStationProducts replaces the products assigned to a station.
// (PUT /stations/{stationId}/products)
func (h *Handlers) SetStationProducts(w http.ResponseWriter, r *http.Request, stationId openapi_types.UUID) {
	var body generated.StationProductAssignment
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	ctx := r.Context()
	sid := uuid.UUID(stationId)

	productIDs := make([]uuid.UUID, 0, len(body.ProductIds))
	for _, pid := range body.ProductIds {
		productIDs = append(productIDs, uuid.UUID(pid))
	}

	if err := h.stations.SetStationProducts(ctx, sid, productIDs); err != nil {
		writeEntError(w, err)
		return
	}

	// Return the updated station with products.
	station, err := h.stations.GetStationByID(ctx, sid)
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, toAPIStation(station))
}

// RemoveStationProduct removes a single product assignment from a station.
// (DELETE /stations/{stationId}/products/{productId})
func (h *Handlers) RemoveStationProduct(w http.ResponseWriter, r *http.Request, stationId openapi_types.UUID, productId openapi_types.UUID) {
	if err := h.stations.RemoveStationProduct(r.Context(), uuid.UUID(stationId), uuid.UUID(productId)); err != nil {
		writeEntError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RedeemAtStation redeems order items at the current station.
// (POST /stations/redeem)
func (h *Handlers) RedeemAtStation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	deviceID, ok := auth.GetDeviceID(ctx)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Device authentication required")
		return
	}

	var body generated.RedemptionCreate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	orderID := uuid.UUID(body.OrderId)

	// Use the Idempotency-Key header if present.
	idemKey := r.Header.Get("Idempotency-Key")

	result, err := h.stations.RedeemAssigned(ctx, deviceID, orderID, idemKey)
	if err != nil {
		writeError(w, http.StatusBadRequest, "redeem_failed", err.Error())
		return
	}

	// Map the result to the RedemptionResult API type.
	matched := 0
	redeemed := 0
	if v, ok := result["matched"].(int); ok {
		matched = v
	}
	if v, ok := result["redeemed"].(int64); ok {
		redeemed = int(v)
	}

	resp := generated.RedemptionResult{
		Matched:  matched,
		Redeemed: redeemed,
	}

	if items, ok := result["items"].([]map[string]any); ok {
		apiItems := make([]struct {
			Id         *openapi_types.UUID `json:"id,omitempty"`
			Quantity   *int                `json:"quantity,omitempty"`
			RedeemedAt *time.Time          `json:"redeemedAt,omitempty"`
			Title      *string             `json:"title,omitempty"`
		}, 0, len(items))
		for _, item := range items {
			entry := struct {
				Id         *openapi_types.UUID `json:"id,omitempty"`
				Quantity   *int                `json:"quantity,omitempty"`
				RedeemedAt *time.Time          `json:"redeemedAt,omitempty"`
				Title      *string             `json:"title,omitempty"`
			}{}
			if idStr, ok := item["id"].(string); ok {
				if uid, err := uuid.Parse(idStr); err == nil {
					apiID := openapi_types.UUID(uid)
					entry.Id = &apiID
				}
			}
			if title, ok := item["title"].(string); ok {
				entry.Title = &title
			}
			if qty, ok := item["quantity"].(int); ok {
				entry.Quantity = &qty
			}
			apiItems = append(apiItems, entry)
		}
		resp.Items = &apiItems
	}

	response.WriteJSON(w, http.StatusOK, resp)
}

// RevokeStation revokes a station device.
// (DELETE /stations/{stationId})
func (h *Handlers) RevokeStation(w http.ResponseWriter, r *http.Request, stationId openapi_types.UUID) {
	if err := h.devices.Revoke(r.Context(), uuid.UUID(stationId)); err != nil {
		writeEntError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
