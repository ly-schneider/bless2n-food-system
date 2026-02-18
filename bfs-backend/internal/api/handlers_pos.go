package api

import (
	"net/http"

	"backend/internal/auth"
	"backend/internal/generated/api/generated"
	"backend/internal/response"
)

// GetCurrentPos returns the POS device information for the current device auth context.
// (GET /pos/me)
func (h *Handlers) GetCurrentPos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	deviceID, ok := auth.GetDeviceID(ctx)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Device authentication required")
		return
	}

	device, err := h.pos.GetDeviceByID(ctx, deviceID)
	if err != nil {
		writeEntError(w, err)
		return
	}

	apiDevice := toAPIDevice(device)
	posDevice := generated.POSDevice{
		Id:        apiDevice.Id,
		Name:      apiDevice.Name,
		Model:     apiDevice.Model,
		Os:        apiDevice.Os,
		Status:    apiDevice.Status,
		CreatedAt: apiDevice.CreatedAt,
	}

	settingsData, err := h.settings.GetSettingsWithProducts(ctx)
	if err == nil && settingsData != nil {
		missingJetons, _ := h.products.CountActiveWithoutJeton(ctx)
		apiSettings := toAPISettings(settingsData, int(missingJetons))
		posDevice.Settings = &apiSettings
	}

	response.WriteJSON(w, http.StatusOK, posDevice)
}

// ListPosDevices returns all POS-type devices with settings.
// (GET /pos/devices)
func (h *Handlers) ListPosDevices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	posType := "POS"
	devices, err := h.devices.ListAll(ctx, &posType, nil)
	if err != nil {
		writeEntError(w, err)
		return
	}

	settingsData, _ := h.settings.GetSettingsWithProducts(ctx)
	var apiSettings *generated.Settings
	if settingsData != nil {
		missingJetons, _ := h.products.CountActiveWithoutJeton(ctx)
		s := toAPISettings(settingsData, int(missingJetons))
		apiSettings = &s
	}

	items := make([]generated.POSDevice, 0, len(devices))
	for _, d := range devices {
		apiDevice := toAPIDevice(d)
		posDevice := generated.POSDevice{
			Id:        apiDevice.Id,
			Name:      apiDevice.Name,
			Model:     apiDevice.Model,
			Os:        apiDevice.Os,
			Status:    apiDevice.Status,
			CreatedAt: apiDevice.CreatedAt,
			Settings:  apiSettings,
		}
		items = append(items, posDevice)
	}

	response.WriteJSON(w, http.StatusOK, generated.POSDeviceList{Items: items})
}
