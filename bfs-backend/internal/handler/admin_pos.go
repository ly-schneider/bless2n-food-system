package handler

import (
	"backend/internal/domain"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/response"
	"backend/internal/service"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AdminPOSHandler struct {
	devices  repository.PosDeviceRepository
	config   service.POSConfigService
	products repository.ProductRepository
}

func NewAdminPOSHandler(devices repository.PosDeviceRepository, config service.POSConfigService, products repository.ProductRepository) *AdminPOSHandler {
	return &AdminPOSHandler{devices: devices, config: config, products: products}
}

// GetSettings godoc
// @Summary Get POS settings
// @Tags admin-pos
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/pos/settings [get]
func (h *AdminPOSHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.config.GetSettings(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	mode := domain.PosModeQRCode
	if settings != nil && settings.Mode != "" {
		mode = settings.Mode
	}
	resp := map[string]any{"mode": mode}
	if h.products != nil {
		if missing, err := h.products.CountActiveWithoutJeton(r.Context()); err == nil {
			resp["missingJetons"] = missing
		}
	}
	response.WriteJSON(w, http.StatusOK, resp)
}

// PatchSettings godoc
// @Summary Update POS settings
// @Tags admin-pos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body map[string]string true "Settings payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemDetails
// @Router /v1/admin/pos/settings [patch]
func (h *AdminPOSHandler) PatchSettings(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Mode domain.PosFulfillmentMode `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if body.Mode == "" {
		response.WriteError(w, http.StatusBadRequest, "mode required")
		return
	}
	if err := h.config.SetMode(r.Context(), body.Mode); err != nil {
		if e, ok := err.(service.MissingJetonForActiveProductsError); ok {
			response.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "missing_jetons", "missing": e.Count})
			return
		}
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"mode": body.Mode})
}

// ListRequests godoc
// @Summary List POS requests
// @Tags admin-pos
// @Security BearerAuth
// @Produce json
// @Param status query string false "Status filter"
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/pos/requests [get]
func (h *AdminPOSHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	var status *domain.PosRequestStatus
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.PosRequestStatus(s)
		status = &st
	}
	items, err := h.devices.List(r.Context(), status)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to list requests")
		return
	}
	type dto struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Model      string `json:"model"`
		OS         string `json:"os"`
		Status     string `json:"status"`
		Token      string `json:"token"`
		CreatedAt  any    `json:"createdAt"`
		ApprovedAt any    `json:"approvedAt,omitempty"`
	}
	out := make([]dto, 0, len(items))
	for _, it := range items {
		out = append(out, dto{ID: it.ID.Hex(), Name: it.Name, Model: it.Model, OS: it.OS, Token: it.DeviceToken, Status: string(it.Status), CreatedAt: it.CreatedAt, ApprovedAt: it.ApprovedAt})
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"items": out})
}

// ListDevices godoc
// @Summary List POS devices
// @Tags admin-pos
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/pos/devices [get]
func (h *AdminPOSHandler) ListDevices(w http.ResponseWriter, r *http.Request) {
	var status *domain.PosRequestStatus
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.PosRequestStatus(s)
		status = &st
	}
	items, err := h.devices.List(r.Context(), status)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to list devices")
		return
	}
	type dto struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Model       string  `json:"model,omitempty"`
		OS          string  `json:"os,omitempty"`
		Token       string  `json:"token"`
		Approved    bool    `json:"approved"`
		Status      string  `json:"status"`
		ApprovedAt  any     `json:"approvedAt,omitempty"`
		CreatedAt   any     `json:"createdAt"`
		CardCapable *bool   `json:"cardCapable,omitempty"`
		PrinterMAC  *string `json:"printerMac,omitempty"`
		PrinterUUID *string `json:"printerUuid,omitempty"`
	}
	out := make([]dto, 0, len(items))
	for _, it := range items {
		out = append(out, dto{
			ID: it.ID.Hex(), Name: it.Name, Model: it.Model, OS: it.OS, Token: it.DeviceToken,
			Approved: it.Approved, Status: string(it.Status), ApprovedAt: it.ApprovedAt, CreatedAt: it.CreatedAt,
			CardCapable: it.CardCapable, PrinterMAC: it.PrinterMAC, PrinterUUID: it.PrinterUUID,
		})
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"items": out})
}

// Approve godoc
// @Summary Approve POS request
// @Tags admin-pos
// @Security BearerAuth
// @Param id path string true "Request ID"
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/pos/requests/{id}/approve [post]
func (h *AdminPOSHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if _, err := h.devices.FindByID(r.Context(), oid); err != nil {
		response.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	var decidedBy *bson.ObjectID
	if claims, ok := middleware.GetUserFromContext(r.Context()); ok && claims != nil && claims.Subject != "" {
		if u, err := bson.ObjectIDFromHex(claims.Subject); err == nil {
			decidedBy = &u
		}
	}
	if err := h.devices.UpdateStatus(r.Context(), oid, domain.PosRequestStatusApproved, decidedBy, time.Now().UTC()); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "update failed")
		return
	}
	response.WriteJSON(w, http.StatusOK, response.Ack{Message: "approved"})
}

// Revoke godoc
// @Summary Revoke POS device access
// @Tags admin-pos
// @Security BearerAuth
// @Param id path string true "Device ID"
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/pos/requests/{id}/revoke [post]
func (h *AdminPOSHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if _, err := h.devices.FindByID(r.Context(), oid); err != nil {
		response.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	var decidedBy *bson.ObjectID
	if claims, ok := middleware.GetUserFromContext(r.Context()); ok && claims != nil && claims.Subject != "" {
		if u, err := bson.ObjectIDFromHex(claims.Subject); err == nil {
			decidedBy = &u
		}
	}
	if err := h.devices.UpdateStatus(r.Context(), oid, domain.PosRequestStatusRevoked, decidedBy, time.Now().UTC()); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "update failed")
		return
	}
	response.WriteJSON(w, http.StatusOK, response.Ack{Message: "revoked"})
}

// Reject godoc
// @Summary Reject POS request
// @Tags admin-pos
// @Security BearerAuth
// @Param id path string true "Request ID"
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/pos/requests/{id}/reject [post]
func (h *AdminPOSHandler) Reject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if _, err := h.devices.FindByID(r.Context(), oid); err != nil {
		response.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	var decidedBy *bson.ObjectID
	if claims, ok := middleware.GetUserFromContext(r.Context()); ok && claims != nil && claims.Subject != "" {
		if u, err := bson.ObjectIDFromHex(claims.Subject); err == nil {
			decidedBy = &u
		}
	}
	if err := h.devices.UpdateStatus(r.Context(), oid, domain.PosRequestStatusRejected, decidedBy, time.Now().UTC()); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "update failed")
		return
	}
	response.WriteJSON(w, http.StatusOK, response.Ack{Message: "rejected"})
}

// PatchConfig godoc
// @Summary Update POS device config
// @Tags admin-pos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/pos/devices/{id}/config [patch]
func (h *AdminPOSHandler) PatchConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		CardCapable *bool   `json:"cardCapable"`
		PrinterMAC  *string `json:"printerMac"`
		PrinterUUID *string `json:"printerUuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.devices.UpdateConfig(r.Context(), oid, body.CardCapable, body.PrinterMAC, body.PrinterUUID); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "update failed")
		return
	}
	response.WriteJSON(w, http.StatusOK, response.Ack{Message: "updated"})
}

// ListJetons godoc
// @Summary List jetons
// @Tags admin-pos
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/pos/jetons [get]
func (h *AdminPOSHandler) ListJetons(w http.ResponseWriter, r *http.Request) {
	items, err := h.config.ListJetons(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to list jetons")
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

// CreateJeton godoc
// @Summary Create jeton
// @Tags admin-pos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body map[string]interface{} true "Jeton payload"
// @Success 201 {object} domain.JetonDTO
// @Failure 400 {object} response.ProblemDetails
// @Router /v1/admin/pos/jetons [post]
func (h *AdminPOSHandler) CreateJeton(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name         string  `json:"name"`
		PaletteColor string  `json:"paletteColor"`
		HexColor     *string `json:"hexColor,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	dto, err := h.config.CreateJeton(r.Context(), body.Name, body.PaletteColor, body.HexColor)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusCreated, dto)
}

// UpdateJeton godoc
// @Summary Update jeton
// @Tags admin-pos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Jeton ID"
// @Param payload body map[string]interface{} true "Jeton payload"
// @Success 200 {object} domain.JetonDTO
// @Failure 400 {object} response.ProblemDetails
// @Router /v1/admin/pos/jetons/{id} [patch]
func (h *AdminPOSHandler) UpdateJeton(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		Name         string  `json:"name"`
		PaletteColor string  `json:"paletteColor"`
		HexColor     *string `json:"hexColor,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	dto, err := h.config.UpdateJeton(r.Context(), oid, body.Name, body.PaletteColor, body.HexColor)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, dto)
}

// DeleteJeton godoc
// @Summary Delete jeton
// @Tags admin-pos
// @Security BearerAuth
// @Param id path string true "Jeton ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ProblemDetails
// @Failure 409 {object} response.ProblemDetails
// @Router /v1/admin/pos/jetons/{id} [delete]
func (h *AdminPOSHandler) DeleteJeton(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.config.DeleteJeton(r.Context(), oid); err != nil {
		if e, ok := err.(service.JetonInUseError); ok {
			response.WriteJSON(w, http.StatusConflict, map[string]any{"error": "jeton_in_use", "usage": e.Count})
			return
		}
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteNoContent(w)
}
