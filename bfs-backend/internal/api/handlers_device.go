package api

import (
	"encoding/json"
	"net/http"

	"backend/internal/auth"
	"backend/internal/generated/api/generated"
	"backend/internal/response"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ListDevices returns all device bindings, optionally filtered by type and status.
// (GET /devices)
func (h *Handlers) ListDevices(w http.ResponseWriter, r *http.Request, params generated.ListDevicesParams) {
	ctx := r.Context()

	var deviceType *string
	if params.Type != nil {
		s := string(*params.Type)
		deviceType = &s
	}
	var status *string
	if params.Status != nil {
		s := string(*params.Status)
		status = &s
	}

	devices, err := h.devices.ListAll(ctx, deviceType, status)
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, generated.DeviceList{
		Items: toAPIDevices(devices),
	})
}

// GetDevice returns a single device by ID.
// (GET /devices/{deviceId})
func (h *Handlers) GetDevice(w http.ResponseWriter, r *http.Request, deviceId openapi_types.UUID) {
	device, err := h.devices.GetByID(r.Context(), uuid.UUID(deviceId))
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, toAPIDevice(device))
}

// RevokeDevice revokes a device binding.
// (DELETE /devices/{deviceId})
func (h *Handlers) RevokeDevice(w http.ResponseWriter, r *http.Request, deviceId openapi_types.UUID) {
	if err := h.devices.Revoke(r.Context(), uuid.UUID(deviceId)); err != nil {
		writeEntError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// CreateDevicePairing starts the device pairing flow by generating a pairing code.
// (POST /devices/pairings)
func (h *Handlers) CreateDevicePairing(w http.ResponseWriter, r *http.Request) {
	var body generated.DevicePairingCreate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	result, err := h.devices.CreatePairing(
		r.Context(),
		body.DeviceKey,
		body.Name,
		string(body.Type),
		body.Model,
		body.Os,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, "pairing_failed", err.Error())
		return
	}

	response.WriteJSON(w, http.StatusCreated, generated.DevicePairing{
		Code:      result.Code,
		ExpiresAt: result.ExpiresAt,
	})
}

// GetDevicePairing polls the status of a device pairing by code.
// The device calls this endpoint to check if an admin has approved the pairing.
// (GET /devices/pairings/{code})
func (h *Handlers) GetDevicePairing(w http.ResponseWriter, r *http.Request, code string) {
	status, err := h.devices.GetPairingStatus(r.Context(), code)
	if err != nil {
		writeEntError(w, err)
		return
	}

	// Map ent device status to API pairing status (pending/completed/expired).
	apiStatus := generated.DevicePairingStatusStatus(status.Status)
	switch status.Status {
	case "approved":
		apiStatus = "completed"
	case "rejected", "revoked":
		apiStatus = "expired"
	}

	resp := generated.DevicePairingStatus{
		Status: apiStatus,
	}

	if status.Token != nil {
		resp.Token = status.Token
	}
	if status.Device != nil {
		apiDevice := toAPIDevice(status.Device)
		resp.Device = &apiDevice
	}

	response.WriteJSON(w, http.StatusOK, resp)
}

// CompleteDevicePairing allows an admin to approve a device pairing code.
// (POST /devices/pairings/{code})
func (h *Handlers) CompleteDevicePairing(w http.ResponseWriter, r *http.Request, code string) {
	ctx := r.Context()

	// Admin must be authenticated.
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	// Decode body (contains the code for confirmation).
	var body generated.DevicePairingComplete
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	device, err := h.devices.CompletePairing(ctx, code, userID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "pairing_failed", err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, toAPIDevice(device))
}
