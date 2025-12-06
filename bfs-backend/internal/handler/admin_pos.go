package handler

import (
	"backend/internal/domain"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/response"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AdminPOSHandler struct {
	devices repository.PosDeviceRepository
}

func NewAdminPOSHandler(devices repository.PosDeviceRepository) *AdminPOSHandler {
	return &AdminPOSHandler{devices: devices}
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
